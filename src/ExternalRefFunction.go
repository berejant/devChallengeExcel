package main

import (
	"devChallengeExcel/contracts"
	"encoding/json"
	"fmt"
	"github.com/antonmedv/expr"
	"net/http"
	"strconv"
	"time"
)

var httpClient = &http.Client{
	Timeout: time.Second * 4,
}

var fetchExternalRef = func(args ...any) (any, error) {
	url := args[0].(string)
	response, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetchExternalRef url %s: %s", url, response.Status)
	}

	var responsePayload contracts.Cell
	err = json.NewDecoder(response.Body).Decode(&responsePayload)
	if err != nil {
		return nil, err
	}

	return parseString(&responsePayload.Result), nil
}

func parseString(stringValueRef *string) interface{} {
	var floatValue float64
	var intValue int64
	var err error

	if intValue, err = strconv.ParseInt(*stringValueRef, 10, 64); err == nil {
		return intValue
	} else if floatValue, err = strconv.ParseFloat(*stringValueRef, 64); err == nil {
		return floatValue
	}

	return stringValueRef
}

var externalRefFunction = expr.Function("external_ref", fetchExternalRef)
