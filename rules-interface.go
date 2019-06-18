package main

import (
	"bytes"
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

// #####################################################################

type RuleRequest struct {
	Lookup   string     `json:"lookup,omitempty"`
	Commands []Commands `json:"commands,omitempty"`
}
type ComRedhatDemoAbnclientClient struct {
}
type ClientObject struct {
	ComRedhatDemoAbnclientClient ComRedhatDemoAbnclientClient `json:"com.redhat.demo.abnclient.Client,omitempty"`
}
type SetGlobal struct {
	Identifier string       `json:"identifier,omitempty"`
	Object     ClientObject `json:"object,omitempty"`
}
type ComMyspaceDatavalidationEntity struct {
	Name     string `json:"name,omitempty"`
	LastName string `json:"lastname,omitempty"`
	Abn      string `json:"abn,omitempty"`
}
type EntityObject struct {
	ComMyspaceDatavalidationEntity ComMyspaceDatavalidationEntity `json:"com.myspace.datavalidation.Entity,omitempty"`
}
type Insert struct {
	OutIdentifier string       `json:"out-identifier,omitempty"`
	ReturnObject  string       `json:"return-object,omitempty"`
	Object        EntityObject `json:"object,omitempty"`
}
type Query struct {
	OutIdentifier string `json:"out-identifier,omitempty"`
	Name          string `json:"name,omitempty,omitempty"`
}
type Commands struct {
	SetGlobal    SetGlobal `json:"set-global,omitempty"`
	Insert       Insert    `json:"insert,omitempty"`
	FireAllRules string    `json:"fire-all-rules,omitempty"`
	Query        Query     `json:"query,omitempty"`
}

// validateABN calls the ABN Validate Rules Engine.
func validateRules(ruleInputs RuleInputs) (RuleResults, error) {
	var err error
	var results RuleResults

	ruleReq := buildRuleRequest(ruleInputs)
	reqBody := []byte(ruleReq)

	// Build the HTTP request.
	client := &http.Client{}
	URL := applicationConfig.RuleServerURL
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
	println(s)

	return results, err
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
			  "out-identifier": "entity",
			  "return-object": "true",
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
			  "out-identifier": "error",
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
