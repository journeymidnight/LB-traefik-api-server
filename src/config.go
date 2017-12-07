package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"os"
	"reflect"
)

var Config *Configuration

const CONFIGPATH = "conf.toml"

func init() {
	Config, _ = LoadConfig()
	fmt.Println(Config)
}

type Configuration struct {
	Logpath  string
	Loglevel string
	Etcd     *Etcd
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
		Logpath:  "api.log",
		Loglevel: "Info",
		Etcd:     etcd,
	}
	return cfg
}

func LoadConfig() (*Configuration, error) {
	rtConfig := DefaultConfiguration()
	if _, err := os.Stat(CONFIGPATH); err != nil {
		fmt.Println("file does exsit")
	} else {
		_, err = toml.DecodeFile("conf.toml", &rtConfig)
		if err != nil {
			fmt.Println("err decode toml")
		}
	}
	mergeConfig(rtConfig, configFromFlag())
	return rtConfig, nil
}

func configFromFlag() *Configuration {
	cfg := &Configuration{Etcd: &Etcd{}}
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
				fmt.Println(v.Field(i), "can not set or value is empty")
			}
		case reflect.Bool:
			if v.Field(i).CanSet() {
				fmt.Println("ok")
				v.Field(i).Set(v1.Field(i))
			} else {
				fmt.Println(v.Field(i).CanSet())
				fmt.Println(v.Field(i), "can not set or value is empty")
			}
		case reflect.Int:
			if v.Field(i).CanSet() && v1.Field(i).Int() != 0 {
				v.Field(i).Set(v1.Field(i))
			} else {
				fmt.Println(v.Field(i).CanSet())
				fmt.Println(v.Field(i), "can not set or value is empty")
			}
		default:
			if v.Field(i).CanSet() && v1.Field(i).Len() != 0 {
				v.Field(i).Set(v1.Field(i))
			} else {
				fmt.Println(v.Field(i), "can not set or value is empty")
			}
		}
	}
}
