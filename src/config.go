package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"reflect"
)

var Config *Configuration = LoadConfig()

const CONFIGPATH = "conf.toml"

type Configuration struct {
	Accesslog string
	Logpath   string
	Loglevel  string
	Etcd      *Etcd
}

type Etcd struct {
	Endpoints string
	Https     bool
	Certfile  string
	CertCA    string
	Keyfile   string
}

func DefaultConfiguration() *Configuration {
	etcd := &Etcd{
		Endpoints: "127.0.0.1",
		Https:     false,
	}
	cfg := &Configuration{
		Accesslog: "api-access.log",
		Logpath:   "api.log",
		Loglevel:  "info",
		Etcd:      etcd,
	}
	return cfg
}

func LoadConfig() *Configuration {
	rtConfig := DefaultConfiguration()
	if _, err := os.Stat(CONFIGPATH); err != nil {
		fmt.Fprint(os.Stderr, "config file does exsit,skipped config file")
	} else {
		_, err = toml.DecodeFile("conf.toml", &rtConfig)
		if err != nil {
			fmt.Fprint(os.Stderr, "failed to decode config file,skipped config file")
		}
	}
	mergeConfig(rtConfig, configFromFlag())
	return rtConfig
}

func configFromFlag() *Configuration {
	cfg := &Configuration{Etcd: &Etcd{}}
	flag.StringVar(&cfg.Accesslog, "accesslog", "", "path for access file")
	flag.StringVar(&cfg.Logpath, "logpath", "", "path for the log file")
	flag.StringVar(&cfg.Loglevel, "loglevel", "", "using standard go library")
	flag.StringVar(&cfg.Etcd.Endpoints, "etcd.endpoints", "", "ip/port pairs seperated by comma")
	flag.BoolVar(&cfg.Etcd.Https, "etcd.https", false, "should we connect the  etcd server using https")
	flag.StringVar(&cfg.Etcd.Certfile, "etcd.certfile", "", "certfile file path used for authentication")
	flag.StringVar(&cfg.Etcd.Keyfile, "etcd.keyfile", "", "key file path used for authentication")
	flag.StringVar(&cfg.Etcd.CertCA, "etcd.certca", "", "ca file path used for authentication")
	flag.Parse()
	return cfg
}

func mergeConfig(defaultcfg, filecfg interface{}) {
	v1 := reflect.ValueOf(filecfg).Elem()
	v := reflect.ValueOf(defaultcfg).Elem()
	mergeValue(v, v1)
}

func mergeValue(v, v1 reflect.Value) {
	for i := 0; i < v.NumField(); i++ {
		switch v.Field(i).Kind() {
		case reflect.Ptr:
			if v.Field(i).CanSet() && !v1.Field(i).IsNil() {
				mergeValue(v.Field(i).Elem(), v1.Field(i).Elem())
			} else {
				fmt.Fprint(os.Stderr, "can not set or value is empty")
			}
		case reflect.Bool:
			if v.Field(i).CanSet() {
				v.Field(i).Set(v1.Field(i))
			} else {
				fmt.Fprint(os.Stderr, "can not set or value is empty")
			}
		case reflect.Int:
			if v.Field(i).CanSet() && v1.Field(i).Int() != 0 {
				v.Field(i).Set(v1.Field(i))
			} else {
				fmt.Fprint(os.Stderr, "can not set or value is empty")
			}
		default:
			if v.Field(i).CanSet() && v1.Field(i).Len() != 0 {
				v.Field(i).Set(v1.Field(i))
			} else {
				fmt.Fprint(os.Stderr, "can not set or value is empty")
			}
		}
	}
}
