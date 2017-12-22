package main

import (
	"fmt"
	"reflect"
	"testing"
)

func TestMergeConfig(t *testing.T) {
	conf1 := &Configuration{
		Accesslog: "conf1.log",
		Logpath:   "conf1",
		Etcd:      &Etcd{Endpoints: "1.1.1.1:2379"},
	}
	conf2 := &Configuration{
		Etcd: &Etcd{Https: true},
	}
	expected := &Configuration{
		Accesslog: "conf1.log",
		Logpath:   "conf1",
		Etcd:      &Etcd{Endpoints: "1.1.1.1:2379", Https: true},
	}
	mergeConfig(conf1, conf2)
	if reflect.DeepEqual(conf1, expected) {
		fmt.Println("ok")
	} else {
		t.Errorf("test for mergeConfig failed")
	}
}
