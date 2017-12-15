package main

import (
	//	"fmt"
	"encoding/json"
	"regexp"
	"strings"
)

type ErrorCode int

const (
	RequiredFieldMissed = iota + 1000
	ServiceAlreadyExist
	ServiceDoesNotExist
	JSONIsNil
	ParseJSONFailed
	ServiceNameInvalid
	ServiceFieldInvalid
	InternalError
	OutOfService
	ServiceNotFound
	InvalidFile
	ErrorEmptyFile
)

const URLPattern string = "^http://[A-Za-z0-9.]+:[0-9]+$"

var EntryPointsValue []string = []string{"http", "https", "http,https", "https,http"}

type APIError struct {
	Ecode    ErrorCode
	EMessage string
}

func NewAPIError(ecode ErrorCode, emessage string) []byte {
	apiError := &APIError{Ecode: ecode, EMessage: emessage}
	errJSON, _ := json.Marshal(apiError)
	return errJSON
}

func NewSuccess(message string) []byte {
	rtdict := map[string]string{"message": message}
	rtJSON, _ := json.Marshal(rtdict)
	return rtJSON
}

func checkNilJSON(body []byte) []byte {
	apiError := &APIError{Ecode: JSONIsNil, EMessage: "the json content can't be nil"}
	errJSON, _ := json.Marshal(apiError)
	if len(body) == 0 {
		return errJSON
	}
	return nil
}

func checkRequiredField(svc *Service) []byte {
	apiError := &APIError{Ecode: RequiredFieldMissed, EMessage: "ther entrypoints and server url can not be empty"}
	errJSON, _ := json.Marshal(apiError)
	if len(svc.EntryPoints) == 0 || len(svc.Servers) == 0 {
		return errJSON
	}
	for _, server := range svc.Servers {
		if server.Url == "" {
			return errJSON
		}
	}
	return nil
}

func checkServiceName(svcname string) []byte {
	apiError := &APIError{Ecode: ServiceNameInvalid, EMessage: "Invalid service name"}
	errJSON, _ := json.Marshal(apiError)
	if m, _ := regexp.MatchString("^[0-9a-zA-z.*]+$", svcname); !m {
		return errJSON
	}
	return nil
}

func checkServiceField(svc *Service) []byte {
	apiError := &APIError{Ecode: ServiceFieldInvalid, EMessage: "failed to pass syntax check"}
	errJSON, _ := json.Marshal(apiError)
	if !sliceContainString(EntryPointsValue, strings.Join(svc.EntryPoints, ",")) {
		log.Warning("the request failed to pass entrypoint syntax check")
		return errJSON
	}
	for _, server := range svc.Servers {
		if m, _ := regexp.MatchString(URLPattern, server.Url); !m {
			log.Warning("the request failed to pass server url syntax check")
			return errJSON
		}
		if server.Weight != "" {
			if m1, _ := regexp.MatchString("^[0-9]+$", server.Weight); !m1 {
				log.Warning("the request failed to pass server weight syntax check")
				return errJSON
			}
		}
	}
	return nil
}

func checkServiceAlreadyExist(svcname string) []byte {
	apiError := &APIError{Ecode: ServiceAlreadyExist, EMessage: "the service already exist,can not create it again"}
	errJSON, _ := json.Marshal(apiError)
	rtsvcs, _ := List(FRONTENDS)
	for _, svc := range rtsvcs {
		if svc == svcname {
			return errJSON
		}
	}
	return nil
}

func checkServiceDoesNotExist(svcname string) []byte {
	apiError := &APIError{Ecode: ServiceDoesNotExist, EMessage: "the service does not exist,please create first"}
	errJSON, _ := json.Marshal(apiError)
	rtsvcs, _ := List(FRONTENDS)
	for _, svc := range rtsvcs {
		if svc == svcname {
			return nil
		}
	}
	return errJSON
}

func sliceContainString(slc []string, str string) bool {
	for _, v := range slc {
		if v == str {
			return true
		}
	}
	return false
}
