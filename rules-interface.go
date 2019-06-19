package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type RuleInputs struct {
	firstName string
	lastName  string
	abn       string
}

// RuleResults returns the success fail of each rule, and an aggregated message string.
type RuleResults struct {
	validFirstName bool
	validLastName  bool
	abnStatus      bool
	message        string
}

type DMResult struct {
	result  bool
	message string
}

var errorList map[string]string

// #####################################################################

func validateRules(ruleInputs RuleInputs) (RuleResults, error) {
	var results RuleResults
	var err error
	ruleReq := buildRuleRequest(ruleInputs)
	reqBody := []byte(ruleReq)

	m := make(map[string]string)
	errorList = m

	var URL string
	if ruleInputs.abn != "" {
		URL = applicationConfig.ABNRuleServerURL
		ruleOkay, _ := callDecisionManager(URL, reqBody)
		results.abnStatus = ruleOkay
	}
	if ruleInputs.firstName != "" {
		URL = applicationConfig.NameRuleServerURL
		ruleOkay, _ := callDecisionManager(URL, reqBody)
		results.validFirstName = ruleOkay
	}
	if ruleInputs.lastName != "" {
		URL = applicationConfig.LNameRuleServerURL
		ruleOkay, _ := callDecisionManager(URL, reqBody)
		results.validLastName = ruleOkay
	}

	return results, err
}

// validateABN calls the ABN Validate Rules Engine. Returns true if rule was okay
func callDecisionManager(URL string, reqBody []byte) (bool, error) {
	var err error
	// var results RuleResults
	var ruleOkay bool

	// Build the HTTP request.
	client := &http.Client{}

	//jsonStr, _x := json.Marshal(ruleReq)
	req, err := http.NewRequest("POST", URL, bytes.NewBuffer(reqBody))
	req.SetBasicAuth(applicationConfig.Username, applicationConfig.Password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/xml")

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	bodyText, err := ioutil.ReadAll(resp.Body)
	s := string(bodyText)
	fmt.Println("INFO: Request Message:\n", s)

	var result map[string]interface{}
	err = json.Unmarshal(bodyText, &result)
	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR: Cannot unmarshall DM response.")
	}

	l := len(errorList)
	parseMap(result)
	if l == len(errorList) {
		ruleOkay = true // No rule violations
	} else {
		ruleOkay = false // Rule violations
	}

	return ruleOkay, err
}

func parseMap(aMap map[string]interface{}) {
	for key, val := range aMap {
		switch concreteVal := val.(type) {
		case map[string]interface{}:
			// fmt.Println(key)
			parseMap(val.(map[string]interface{}))
		case []interface{}:
			// fmt.Println(key)
			parseArray(val.([]interface{}))
		default:
			//fmt.Print("M> ")
			//fmt.Println(key, ":", concreteVal)
			if key == "cause" {
				// fmt.Println("VALIDATION ERROR: ", concreteVal)
				msg := fmt.Sprintf("%s", concreteVal)
				errorList[msg] = "error"
			}
		}
	}
}

func parseArray(anArray []interface{}) {
	for _, val := range anArray {
		switch val.(type) {
		case map[string]interface{}:
			// fmt.Println("Index:", i)
			parseMap(val.(map[string]interface{}))
		case []interface{}:
			// fmt.Println("Index:", i)
			parseArray(val.([]interface{}))
		default:
			// Do nothing
			// fmt.Print("A> ")
			//fmt.Println("Index", i, ":", concreteVal)

		}
	}
}

// buildRuleRequest
func buildRuleRequest(ruleInputs RuleInputs) string {
	jsonReq := []byte(`{
		"lookup": "statelessSession",
		"commands": [
		  {
			"set-global": {
			  "identifier": "service",
			  "object": {
				"com.redhat.demo.abnclient.Client": {}
			  }
			}
		  },
		  {
			"insert": {
			  "object": {
				"com.myspace.datavalidation.Entity": {
					*****
				}
			  }
			}
		  },
		  {
			"fire-all-rules": ""
		  },
		  {
			"query": {
			  "out-identifier": "error-results",
			  "name": "get_validation_error"
			}
		  }
		]
	  }`)

	var query string
	if ruleInputs.firstName != "" {
		query = "\"name\" : \"" + ruleInputs.firstName + "\""
	}

	if ruleInputs.lastName != "" {
		if len(query) > 0 && query[len(query)-1:] == "\"" {
			query = query + ", "
		}
		query = query + "\"lastName\" : \"" + ruleInputs.lastName + "\""
	}

	if ruleInputs.abn != "" {
		if len(query) > 0 && query[len(query)-1:] == "\"" {
			query = query + ", "
		}
		query = query + "\"abn\" : \"" + ruleInputs.abn + "\""
	}

	// BIGGEST HACK EVER!
	s := string(jsonReq)
	s = strings.Replace(s, "*****", query, 1)
	fmt.Fprintln(os.Stdout, "INFO: Rules query: \n", s)

	return s
}
