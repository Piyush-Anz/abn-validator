package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/bryonbaker/utils"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const (
	// DefaultServiceVersion keeps the service version handy
	DefaultServiceVersion string = "unknown"

	// BootConfigurationFile contains the name of the file containing the bootstrap configuratiopn.
	// It must exist in the same directory as the executable.
	BootConfigurationFile = "boot-config.json"
)

// ApplicationConfig holds the startup config used to bootstrap all other config.
type ApplicationConfig struct {
	Version            string `json:"Version,omitempty"`
	ListeningPort      string `json:"ListeningPort,omitempty"`
	AusGovGUID         string `json:"UniqueID,omitempty"`
	AusGovURL          string `json:"AusGovURL,omitempty"`
	NameRuleServerURL  string `json:"NameRuleServerURL,omitempty"`
	LNameRuleServerURL string `json:"LNameRuleServerURL,omitempty"`
	ABNRuleServerURL   string `json:"ABNRuleServerURL,omitempty"`
	CallbackFunction   string `json:"CallbackFunction,omitempty"`
	Username           string `json:"Username,omitempty"`
	Password           string `json:"Password,omitempty"`
}

// Holds the application's config
var applicationConfig ApplicationConfig

// ##########################################
// ### ABN Validator Web Service Messages ###
// ##########################################

// ABNSvcRequest defines the format of the HTML json request.
type ABNSvcRequest struct {
	Abn string `json:"Abn,omitempty"`
}

// ABNSvcResponse defines the format of the HTML json response.
type ABNSvcResponse struct {
	AbnStatus string `json:"AbnStatus,omitempty"`
	// EntityName string `json:"EntityName,omitempty"`
	Message string `json:"Message,omitempty"`
}

// ####################################
// ### Aus Gov Web Service Messages ###
// ####################################

// AbnLookupResponse contains the response to the web-service call to the
// Aus Gov ABN lookup service
type AbnLookupResponse struct {
	Abn             string   `json:"Abn,omitempty"`
	AbnStatus       string   `json:"AbnStatus,omitempty"`
	AddressDate     string   `json:"AddressDate,omitempty"`
	AddressPostcode string   `json:"AddressPostcode,omitempty"`
	AddressState    string   `json:"AddressState,omitempty"`
	BusinessNames   []string `json:"BusinessName,omitempty"`
	EntityName      string   `json:"EntityName,omitempty"`
	EntityTypeCode  string   `json:"EntityTypeCode,omitempty"`
	EntityTypeName  string   `json:"EntityTypeName,omitempty"`
	Gst             string   `json:"Gst,omitempty"`
	Message         string   `json:"Message,omitempty"`
}

// AliveResponse defines the format of the HTML json response.
type AliveResponse struct {
	Alive bool `json:"alive,omitempty"`
}

// ###########################################
// ### Messages for testing the browser UI ###
// ###########################################

// TestFormRequest holds the startup config used to bootstrap all other config.
type TestFormRequest struct {
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	ABN       string `json:"abn,omitempty"`
}

// TestFormResponse holds the startup config used to bootstrap all other config.
type TestFormResponse struct {
	ValidFirstName bool   `json:"validFirstName,omitempty"`
	ValidLastName  bool   `json:"validLastName,omitempty"`
	AbnStatus      bool   `json:"abnStatus,omitempty"`
	Message        string `json:"message,omitempty"`
}

// HomeHandler is the simple request handler that takes no onput parameters.
func homeHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Println("HomeHandler()")

	var jsonResponse AbnLookupResponse
	jsonResponse.AbnStatus = "Current"

	json.NewEncoder(w).Encode(jsonResponse)
}

// abnCheckHandler is a request that has an input parameter of the client ID.
func abnLookupHandler(w http.ResponseWriter, req *http.Request) {

	var abnLookup ABNSvcRequest
	_ = json.NewDecoder(req.Body).Decode(&abnLookup)
	fmt.Fprint(os.Stdout, "\nINFO: Received request to look up ABN: ", abnLookup.Abn)

	// Call out to ABR and get the ABN details
	abnDetails, err := getAbnFromAusGov(abnLookup.Abn)

	var jsonResponse ABNSvcResponse
	if err == nil {
		jsonResponse.AbnStatus = abnDetails.AbnStatus
		// jsonResponse.EntityName = abnDetails.EntityName
		if abnDetails.Message != "" {
			// 	jsonResponse.Message = abnDetails.EntityName
			// } else {
			jsonResponse.Message = abnDetails.Message
		}

	} else {
		jsonResponse.AbnStatus = ""
		jsonResponse.Message = "\nERROR: Aus Gov ABN service returned: \"" + err.Error() + "\""
	}

	w.Header().Set("Content-Type", "application/json")

	// Build the response.
	json.NewEncoder(w).Encode(jsonResponse)
}

func getAbnFromAusGov(abn string) (AbnLookupResponse, error) {
	var err error
	var abnDetails AbnLookupResponse

	// This is a work around because I cannot get the http methods to add the query parameters to the url.
	var fullURL = applicationConfig.AusGovURL + "?abn=" + abn +
		"&callback=" + applicationConfig.CallbackFunction +
		"&guid=" + applicationConfig.AusGovGUID

	// Create an empty JSON parameter list to include in the POST.
	var jsonParms = make(map[string]string)
	requestBody, err := json.Marshal(jsonParms)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nERROR: getAbnFromAusGov(): Cannot Marshal empty map: "+err.Error())
		abnDetails.Message = "ERROR: ABN Lookup service error. Consult logs."
	}

	response, err := http.Post(fullURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nERROR: http.PostForm() returned: "+err.Error())
		abnDetails.Message = "ERROR: ABN Lookup service error. Consult logs."
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\nERROR: ioutil.ReadAll() returned: "+err.Error())
		abnDetails.Message = "ERROR: ABN Lookup service error. Consult logs."
	} else {
		s := string(body)

		// Strip out the JSON from the response
		s = strings.TrimPrefix(s, applicationConfig.CallbackFunction+"(")
		s = strings.TrimSuffix(s, ")")

		fmt.Fprintf(os.Stdout, "\nINFO: POST response:\n")
		fmt.Fprintln(os.Stdout, s)

		err := json.Unmarshal([]byte(s), &abnDetails)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nERROR: Cannot unmarshal JSON into ABN Details. "+err.Error())
		}
	}

	return abnDetails, err
}

// aliveCheckHandler is a request that has an input parameter of the client ID.
func alivezHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Liveness probe...")

	var alivezResponse AliveResponse
	alivezResponse.Alive = true

	w.Header().Set("Content-Type", "application/json")

	// Build the response.
	json.NewEncoder(w).Encode(alivezResponse)
}

// formHandler is a debug route for the browser form
func formHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Println("formHandler()")

	var resp TestFormResponse

	var jsonReq TestFormRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&jsonReq)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Cannot decode request JSON. Error: ", err.Error())
		resp.Message = "ERROR"
	} else {
		fmt.Println("INFO: First name: ", jsonReq.FirstName)
		fmt.Println("INFO: Last name: ", jsonReq.LastName)
		fmt.Println("INFO: ABN: ", jsonReq.ABN)

		rules := RuleInputs{firstName: jsonReq.FirstName, lastName: jsonReq.LastName, abn: jsonReq.ABN}
		validateRules(rules)

		// TODO: Build the JSON response.

		w.Header().Set("Content-Type", "application/json")
	}

	// Build the response.
	json.NewEncoder(w).Encode(resp)
}

// formTestHandler is a debug route for the browser form
func formTestHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Println("formTestHandler()")

	var resp TestFormResponse

	var jsonReq TestFormRequest
	decoder := json.NewDecoder(req.Body)
	err := decoder.Decode(&jsonReq)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Cannot decode request JSON. Error: ", err.Error())
		resp.Message = "ERROR"
	} else {
		fmt.Println("INFO: First name: ", jsonReq.FirstName)

		// Don't bother calling it if the ABN is not supplied.
		if jsonReq.ABN != "" {
			abnResp, err := getAbnFromAusGov(jsonReq.ABN)
			if err == nil {
				if abnResp.AbnStatus == "Active" {
					resp.AbnStatus = true
				} else {
					resp.AbnStatus = false
				}
			}

			if abnResp.Message == "" {
				resp.Message = abnResp.EntityName
			} else {
				resp.Message = abnResp.Message
			}
		}

		// Validation of names.
		if jsonReq.FirstName != "" {
			resp.ValidFirstName = true
		}

		if jsonReq.LastName != "" {
			resp.ValidLastName = true
		}
	}
	w.Header().Set("Content-Type", "application/json")

	// Build the response.
	json.NewEncoder(w).Encode(resp)
}

// Initialise loads the boot configuration information.
func initialise() ApplicationConfig {
	fmt.Println("initialise()")

	var bootConfig utils.BootConfig
	bootConfig, err := utils.LoadBootConfig(BootConfigurationFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Cannot load bot configuration")
		os.Exit(-1)
	}

	var appConfig ApplicationConfig
	fmt.Fprintln(os.Stdout, "INFO: Loading application config from: ", bootConfig.ApplicationConfig)

	jsonFile, err := os.Open(bootConfig.ApplicationConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Configuration file: "+bootConfig.ApplicationConfig+" does not exist\n")
		os.Exit(-1)
	}

	defer jsonFile.Close()
	byteVal, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: Cannot read the json file contents.\n")
		os.Exit(-1)
	}

	utils.Must(json.Unmarshal(byteVal, &appConfig))

	sDec, _ := base64.StdEncoding.DecodeString(appConfig.AusGovGUID)
	appConfig.AusGovGUID = string(sDec)

	return appConfig
}

func main() {
	fmt.Println("Initialising service...")
	applicationConfig = initialise()
	fmt.Println("Starting service version: ", applicationConfig.Version)

	router := mux.NewRouter()
	corsMw := mux.CORSMethodMiddleware(router)

	// router.HandleFunc("/", homeHandler).Methods("GET", "POST", "OPTIONS")
	router.HandleFunc("/abnlookup", abnLookupHandler).Methods("GET", "POST", "OPTIONS")
	router.HandleFunc("/form", formHandler).Methods("GET", "POST", "OPTIONS")
	router.HandleFunc("/form/test", formTestHandler).Methods("GET", "POST", "OPTIONS")
	router.HandleFunc("/alivez", alivezHandler).Methods("GET", "POST", "OPTIONS")

	router.Use(corsMw)

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})

	var port = ":" + applicationConfig.ListeningPort
	log.Fatal(http.ListenAndServe(port, handlers.CORS(originsOk, headersOk, methodsOk)(router)))
}
