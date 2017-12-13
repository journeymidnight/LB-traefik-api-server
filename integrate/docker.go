package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
)

var createNet []byte = []byte(`{
  "Name":"isolated_nw",
  "CheckDuplicate":true,
  "Driver":"bridge",
  "EnableIPv6": true,
  "IPAM":{
    "Driver": "default",
    "Config":[
      {
        "Subnet":"172.20.0.0/16",
        "IPRange":"172.20.10.0/24",
        "Gateway":"172.20.10.1"
      },
      {
        "Subnet":"2001:db8:abcd::/64",
        "Gateway":"2001:db8:abcd::1011"
      }
    ],
    "Options": {
      "foo": "bar"
    }
  },
  "Internal":true,
  "Options": {
    "com.docker.network.bridge.default_bridge": "true",
    "com.docker.network.bridge.enable_icc": "true",
    "com.docker.network.bridge.enable_ip_masquerade": "true",
    "com.docker.network.bridge.host_binding_ipv4": "0.0.0.0",
    "com.docker.network.bridge.name": "docker0",
    "com.docker.network.driver.mtu": "1500"
  },
  "Labels": {
    "com.example.some-label": "some-value",
    "com.example.some-other-label": "some-other-value"
  }
}`)

var createCon []byte = []byte(`{
      "Image": "137e18095760",
      "NetworkMode": "isolated_nw",
      "NetworkingConfig": {
          "EndpointsConfig": {
              "isolated_nw" : {
                  "IPAMConfig": {
                      "IPv4Address":"172.20.10.100"
                  }
              }
          }
      }
 }`)

func SocketConn(proto, addr string) (net.Conn, error) {
	conn, e := net.Dial("unix", "/var/run/docker.sock")
	if e != nil {
		fmt.Println("err", e)
	}
	fmt.Println(conn, e)
	return conn, e
}

func createClient() *http.Client {
	tr := &http.Transport{Dial: SocketConn}
	client := &http.Client{Transport: tr}
	return client
}

func createContainer() error {
	cli := createClient()
	resp, e := cli.Post("http://v1.24/containers/create", "application/json", bytes.NewBuffer(createCon))
	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		fmt.Println(e)
		return e
	}
	fmt.Println(string(body))
	return nil
}

func startContainer() error {
	cli := createClient()
	resp, e := cli.Post("http://v1.24/containers/e0cddf1686e5/start", "application/json", nil)
	defer resp.Body.Close()
	_, e = ioutil.ReadAll(resp.Body)
	if e != nil {
		fmt.Print("error")
		fmt.Println("elete err,", e)
		return e
	}
	return nil
}
func deleteContainer() error {
	cli := createClient()
	request, e := http.NewRequest("DELETE", "http://v1.24/containers/70fc8f0d7cf0", nil)
	resp, e := cli.Do(request)
	defer resp.Body.Close()
	_, e = ioutil.ReadAll(resp.Body)
	if e != nil {
		fmt.Print("error")
		fmt.Println("elete err,", e)
		return e
	}
	return nil
}

func createNetwork() error {
	cli := createClient()
	resp, e := cli.Post("http://v1.24/networks/create", "application/json", bytes.NewBuffer(createNet))
	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return e
	}
	fmt.Println(string(body))
	return nil
}

func deleteNetwork() error {
	cli := createClient()
	request, e := http.NewRequest("DELETE", "http://v1.24/networks/isolated_nw", nil)
	resp, e := cli.Do(request)
	defer resp.Body.Close()
	_, e = ioutil.ReadAll(resp.Body)
	if e != nil {
		fmt.Println(e)
		return e
	}
	return nil
}

func main() {
	//	createContainer()
	startContainer()
}
