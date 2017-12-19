package main

import (
	"fmt"
	"net/http"

	"encoding/json"
	"github.com/gorilla/mux"
	negronilogrus "github.com/meatballhat/negroni-logrus"
	"github.com/urfave/negroni"
	"io/ioutil"
	"strconv"
	"strings"
)

const FRONTENDS string = "/traefik/frontends/"
const BACKENDS string = "/traefik/backends/"

var CertPATH = "/traefik/tlsconfiguration/"
var CERT = "/certificate/certfile"
var KEY = "/certificate/keyfile"
var ENTRYPOINT = "/entrypoints"

var AllRoutes []map[string]string

type Certs struct {
	CertFile string
	KeyFile  string
}

func main() {
	r := mux.NewRouter()
	r.StrictSlash(true)
	n := negroni.New() // Includes some default middlewares

	// add jwt authentication
	//  JWTMiddleware := JWTMiddlewareNew()
	//  n.Use(negroni.HandlerFunc(JWTMiddleware.ServeHTTP))

	// add log middleware
	logrusMiddleWare := negronilogrus.NewMiddleware()
	file, err := openAccessLogFile()
	if err == nil {
		logrusMiddleWare.Logger.Out = file
		n.Use(logrusMiddleWare)
		defer file.Close()
	}

	RegisterRequests(r)
	ShowAPI(r)
	/* add middlewares here, since router is the last one */
	n.UseHandler(r)

	http.ListenAndServe(":80", n)
	log.Info("Web server started")
}

func ShowAPI(r *mux.Router) {
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		t, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		m, err := route.GetMethods()
		if err != nil {
			return err
		}
		sm := strings.Join(m, ",")
		SRoute := map[string]string{"method": sm, "path": t}
		AllRoutes = append(AllRoutes, SRoute)
		return nil
	})

}

func RegisterRequests(r *mux.Router) {
	r.HandleFunc("/", ListAPI).Methods("GET")
	r.HandleFunc("/api", ListAPI).Methods("GET")
	r.HandleFunc("/api/v1", ListAPI).Methods("GET")
	r.HandleFunc("/api/v1/services", ListServices).Methods("GET")
	r.HandleFunc("/api/v1/services/{service}", DetailService).Methods("GET")
	r.HandleFunc("/api/v1/services/{service}", CreateService).Methods("POST")
	r.HandleFunc("/api/v1/services/{service}", UpdateService).Methods("PUT")
	r.HandleFunc("/api/v1/services/{service}", DeleteService).Methods("DELETE")
	r.HandleFunc("/api/v1/certs", ListCerts).Methods("GET")
	r.HandleFunc("/api/v1/certs/{service}", CreateCert).Methods("POST")
	r.HandleFunc("/api/v1/certs/{service}", UpdateCert).Methods("PUT")
	r.HandleFunc("/api/v1/certs/{service}", DeleteCert).Methods("DELETE")
	r.HandleFunc("/api/v1/certs/{service}", DetailCert).Methods("GET")
}

func ListAPI(w http.ResponseWriter, r *http.Request) {
	rtjson, _ := json.Marshal(AllRoutes)
	fmt.Fprintf(w, string(rtjson))
	return
}

func ListCerts(w http.ResponseWriter, r *http.Request) {
	services, err := List(CertPATH)
	if err != nil {
		apiError := &APIError{Ecode: OutOfService, EMessage: "Can't get service"}
		retjson, _ := json.Marshal(apiError)
		fmt.Fprintf(w, string(retjson))
		return
	}
	if len(services) == 0 {
		apiError := &APIError{Ecode: ServiceNotFound, EMessage: "No services found"}
		retjson, _ := json.Marshal(apiError)
		fmt.Fprintf(w, string(retjson))
		return
	}

	res, err := json.Marshal(services)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, string(res))
	return
}

func CreateCert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["service"]
	if err := checkServiceDoesNotExist(serviceName); err != nil {
		fmt.Fprintf(w, string(err))
		return
	}

	var file Certs
	body, _ := ioutil.ReadAll(r.Body)
	if err := checkNilJSON(body); err != nil {
		fmt.Fprintf(w, string(err))
		return
	}

	if err := json.Unmarshal(body, &file); err != nil {
		apiError := &APIError{Ecode: ParseJSONFailed, EMessage: "fail to parse the content"}
		retjson, _ := json.Marshal(apiError)
		fmt.Fprintf(w, string(retjson))
		return
	}

	if _, cn, err := parseCert(file.CertFile, file.KeyFile); err != nil {
		apiError := &APIError{Ecode: InvalidFile, EMessage: "cert file not valid"}
		retjson, _ := json.Marshal(apiError)
		fmt.Fprintf(w, string(retjson))
		return
	} else if cn == "" || cn != serviceName {
		apiError := &APIError{Ecode: InvalidFile, EMessage: "Common name not correct"}
		retjson, _ := json.Marshal(apiError)
		fmt.Fprintf(w, string(retjson))
		return
	}

	if err := Put(CertPATH+serviceName+CERT, file.CertFile); err != nil {
		apiError := &APIError{Ecode: InternalError, EMessage: "error while adding cert file"}
		retjson, _ := json.Marshal(apiError)
		fmt.Fprintf(w, string(retjson))
		return
	}
	if err := Put(CertPATH+serviceName+KEY, file.KeyFile); err != nil {
		apiError := &APIError{Ecode: InternalError, EMessage: "error while adding key file"}
		retjson, _ := json.Marshal(apiError)
		fmt.Fprintf(w, string(retjson))
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "create successfully\n")
	return
}

func DetailCert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	servicename := vars["service"]

	var file Certs
	var err error
	if file.CertFile, err = Get(CertPATH + servicename + CERT); err != nil {
		apiError := &APIError{Ecode: InternalError, EMessage: "error while getting key file"}
		retjson, _ := json.Marshal(apiError)
		fmt.Fprintf(w, string(retjson))
		return
	}
	if file.KeyFile, err = Get(CertPATH + servicename + KEY); err != nil {
		apiError := &APIError{Ecode: InternalError, EMessage: "error while getting key file"}
		retjson, _ := json.Marshal(apiError)
		fmt.Fprintf(w, string(retjson))
		return
	}

	if file.CertFile == "" || file.KeyFile == "" {
		apiError := &APIError{Ecode: ErrorEmptyFile, EMessage: "file empty"}
		retjson, _ := json.Marshal(apiError)
		fmt.Fprintf(w, string(retjson))
		return
	}

	res, _ := json.Marshal(file)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, string(res))
	return
}

func UpdateCert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["service"]

	if err := checkServiceAlreadyExist(serviceName); err != nil {
		fmt.Fprintf(w, string(err))
		return
	}

	body, _ := ioutil.ReadAll(r.Body)
	if err := checkNilJSON(body); err != nil {
		fmt.Fprintf(w, string(err))
		return
	}

	var file Certs
	if err := json.Unmarshal(body, &file); err != nil {
		apiError := &APIError{Ecode: ParseJSONFailed, EMessage: "fail to parse the content"}
		retjson, _ := json.Marshal(apiError)
		fmt.Fprintf(w, string(retjson))
		return
	}

	if _, cn, err := parseCert(file.CertFile, file.KeyFile); err != nil {
		apiError := &APIError{Ecode: InvalidFile, EMessage: "cert file not valid"}
		retjson, _ := json.Marshal(apiError)
		fmt.Fprintf(w, string(retjson))
		return
	} else if cn == "" || cn != serviceName {
		apiError := &APIError{Ecode: InvalidFile, EMessage: "Common name not correct"}
		retjson, _ := json.Marshal(apiError)
		fmt.Fprintf(w, string(retjson))
		return
	}

	if err := Put(CertPATH+serviceName+CERT, file.CertFile); err != nil {
		apiError := &APIError{Ecode: InternalError, EMessage: "error while updating cert file"}
		retjson, _ := json.Marshal(apiError)
		fmt.Fprintf(w, string(retjson))
		return
	}
	if err := Put(CertPATH+serviceName+KEY, file.KeyFile); err != nil {
		apiError := &APIError{Ecode: InternalError, EMessage: "error while updating key file"}
		retjson, _ := json.Marshal(apiError)
		fmt.Fprintf(w, string(retjson))
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "update successfully\n")
	return
}

func DeleteCert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceName := vars["service"]

	if err := DeleteWithPrefix(CertPATH + serviceName); err != nil {
		apiError := &APIError{Ecode: InternalError, EMessage: "error while deleting files"}
		retjson, _ := json.Marshal(apiError)
		fmt.Fprintf(w, string(retjson))
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "delete successfully\n")
	return
}

func ListServices(w http.ResponseWriter, r *http.Request) {
	rtservers, err := List(FRONTENDS)
	if err != nil {
		rtjson := NewAPIError(InternalError, "internal error,please contact the administrator")
		fmt.Fprint(w, string(rtjson))
		return
	}
	if len(rtservers) == 0 {
		rtjson := NewSuccess("there is no service now")
		fmt.Fprint(w, string(rtjson))
		return
	} else {
		rtjson, _ := json.Marshal(rtservers)
		fmt.Fprint(w, string(rtjson))
	}

}

func DetailService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	svcname := vars["service"]
	if rtjson := checkServiceDoesNotExist(svcname); rtjson != nil {
		fmt.Fprint(w, string(rtjson))
		return
	}
	svc := &Service{}
	if e := svc.syncFromEtcd(svcname); e != nil {
		rtjson := NewAPIError(InternalError, "internal error,please contact the administrator")
		fmt.Fprint(w, string(rtjson))
		return
	} else {
		rtjson, _ := json.Marshal(svc)
		fmt.Fprint(w, string(rtjson))
	}
}

func CreateService(w http.ResponseWriter, r *http.Request) {
	var rtjson []byte
	vars := mux.Vars(r)
	svcname := vars["service"]
	body, _ := ioutil.ReadAll(r.Body)
	if rtjson = createService(svcname, body); rtjson != nil {
		fmt.Fprint(w, string(rtjson))
		return
	} else {
		rtjson = NewSuccess("create service successfully")
		fmt.Fprint(w, string(rtjson))
	}
}

func createService(svcname string, body []byte) []byte {
	var s *Service = &Service{}
	var rtjson []byte
	if rtjson = checkNilJSON(body); rtjson != nil {
		return rtjson
	}
	e := json.Unmarshal(body, s)
	if e != nil {
		apiError := APIError{Ecode: ParseJSONFailed, EMessage: "failed to parse json,please check the json format"}
		rtjson, _ = json.Marshal(apiError)
		return rtjson
	}

	//check the service name is valid
	if rtjson = checkServiceName(svcname); rtjson != nil {
		return rtjson
	}
	// checke if the service has already exists.
	if rtjson = checkServiceAlreadyExist(svcname); rtjson != nil {
		return rtjson
	}

	// check required fields
	if rtjson = checkRequiredField(s); rtjson != nil {
		return rtjson
	}

	//check the field in Service object
	if rtjson = checkServiceField(s); rtjson != nil {
		return rtjson
	}

	if e := s.syncToEtcd(svcname); e != nil {
		rtjson := NewAPIError(InternalError, "internal error,please contact the administrator")
		return rtjson
	} else {
		return nil
	}
}

func UpdateService(w http.ResponseWriter, r *http.Request) {
	var s *Service = &Service{}
	var rtjson []byte
	var e error
	body, _ := ioutil.ReadAll(r.Body)
	if rtjson = checkNilJSON(body); rtjson != nil {
		fmt.Fprint(w, string(rtjson))
		return
	}
	e = json.Unmarshal(body, s)
	if e != nil {
		apiError := APIError{Ecode: ParseJSONFailed, EMessage: "failed to parse json,please check the json format"}
		rtjson, _ = json.Marshal(apiError)
		fmt.Fprint(w, string(rtjson))
		return
	}

	// checke if the service has already exists.
	vars := mux.Vars(r)
	svcname := vars["service"]
	if rtjson = checkServiceDoesNotExist(svcname); rtjson != nil {
		fmt.Fprint(w, string(rtjson))
		return
	}

	// check required fields
	if rtjson = checkRequiredField(s); rtjson != nil {
		fmt.Fprint(w, string(rtjson))
		return
	}

	//check the field in Service object
	if rtjson = checkServiceField(s); rtjson != nil {
		fmt.Fprint(w, string(rtjson))
		return
	}
	if rtjson = deleteService(svcname); rtjson != nil {
		fmt.Fprint(w, string(rtjson))
		return
	}

	if rtjson = createService(svcname, body); rtjson != nil {
		fmt.Fprint(w, string(rtjson))
		return
	}
	rtjson = NewSuccess("update service successfully")
	fmt.Fprint(w, string(rtjson))
}

func DeleteService(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	svcname := vars["service"]
	if rtjson := checkServiceDoesNotExist(svcname); rtjson != nil {
		fmt.Fprint(w, string(rtjson))
		return
	}
	if rtjson := deleteService(svcname); rtjson != nil {
		fmt.Fprint(w, string(rtjson))
		return
	} else {
		rtjson := NewSuccess("delete service successfully")
		fmt.Fprint(w, string(rtjson))
	}
}

func deleteService(svcname string) []byte {
	prefixs := []string{FRONTENDS + svcname, BACKENDS + svcname}
	/*	certs, _ := List("/traefik/tlsconfiguration")
		if sliceContainString(certs,svcname) {
			prefixs = append(prefixs,"/traefik/tlsconfiguration/"+svcname)
		}
	*/
	if e := DeleteWithPrefixInList(prefixs); e != nil {
		rtjson := NewAPIError(InternalError, "internal error,please contact the administrator")
		return rtjson
	} else {
		return nil
	}
}

type Service struct {
	EntryPoints []string
	Servers     []*Server
	Health      *HealthCheck
	Stickiness  string
}

func (svc *Service) syncFromEtcd(svcname string) error {
	entrypoints, _ := Get(FRONTENDS + svcname + "/entrypoints")
	servers := GetServers(svcname)
	hpath, _ := Get(BACKENDS + svcname + "/healthcheck/path")
	hinterval, _ := Get(BACKENDS + svcname + "/healthcheck/interval")
	stickiness, _ := Get(BACKENDS + svcname + "/loadbalancer/stickiness")
	myhealth := &HealthCheck{Path: hpath, Interval: hinterval}

	svc.EntryPoints = strings.Split(entrypoints, ",")
	svc.Servers = servers
	svc.Health = myhealth
	svc.Stickiness = stickiness
	return nil
}

func (svc *Service) syncToEtcd(domainName string) error {
	/*
		Put(FRONTENDS+domainName+"/entrypoints",strings.Join(svc.EntryPoints,","))
		Put(FRONTENDS+domainName+"/backend",domainName)
		for k,server  := range svc.Servers {
			Put(BACKENDS+domainName+"/servers/server"+strconv.Itoa(k)+"/url",server.Url)
			Put(BACKENDS+domainName+"/servers/server"+strconv.Itoa(k)+"/weight",server.Weight)
		}
		Put(BACKENDS+domainName+"/healthcheck/path",svc.Health.Path)
		Put(BACKENDS+domainName+"/healthcheck/interval",svc.Health.Interval)
		Put(BACKENDS+domainName+"/loadbalancer/stickiness",svc.Stickness) */
	srcdict := map[string]string{
		FRONTENDS + domainName + "/entrypoints":       strings.Join(svc.EntryPoints, ","),
		FRONTENDS + domainName + "/backend":           domainName,
		FRONTENDS + domainName + "/routes/route/rule": "Host:" + domainName,
	}
	for k, server := range svc.Servers {
		srcdict[BACKENDS+domainName+"/servers/server"+strconv.Itoa(k)+"/url"] = server.Url
		if server.Weight != "" {
			srcdict[BACKENDS+domainName+"/servers/server"+strconv.Itoa(k)+"/weight"] = server.Weight
		}
	}
	if svc.Stickiness != "" {
		srcdict[BACKENDS+domainName+"/loadbalancer/stickiness"] = svc.Stickiness
		srcdict[BACKENDS+domainName+"/loadbalancer/method"] = "wrr"
		srcdict[BACKENDS+domainName+"/loadbalancer/cookiename"] = domainName
	}
	if svc.Health != nil && svc.Health.Path != "" {
		srcdict[BACKENDS+domainName+"/healthcheck/path"] = svc.Health.Path
	}
	if svc.Health != nil && svc.Health.Interval != "" {
		srcdict[BACKENDS+domainName+"/healthcheck/interval"] = svc.Health.Interval
	}
	if err := PutMap(srcdict); err != nil {
		return err
	}
	return nil
}

type Server struct {
	Url    string
	Weight string
}

type HealthCheck struct {
	Path     string
	Interval string
}

func GetServers(servicename string) []*Server {
	serverdir, _ := List(BACKENDS + servicename + "/servers")
	result := []*Server{}
	for _, server := range serverdir {
		tserver := &Server{}
		tserver.Url, _ = Get(BACKENDS + servicename + "/servers/" + server + "/url")
		tserver.Weight, _ = Get(BACKENDS + servicename + "/servers/" + server + "/weight")
		result = append(result, tserver)
	}
	return result
}
