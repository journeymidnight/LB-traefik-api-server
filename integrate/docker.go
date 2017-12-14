package integrate

import (
	"bytes"
	"encoding/json"
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

func SocketConn(proto, addr string) (net.Conn, error) {
	conn, e := net.Dial("unix", "/var/run/docker.sock")
	if e != nil {
	}
	return conn, e
}

func createClient() *http.Client {
	tr := &http.Transport{Dial: SocketConn}
	client := &http.Client{Transport: tr}
	return client
}

type Result struct {
	Id      string
	Message string
}

func createContainer(filename string) (string, error) {
	cli := createClient()
	createCon, _ := ioutil.ReadFile(filename)
	resp, e := cli.Post("http://v1.24/containers/create", "application/json", bytes.NewBuffer(createCon))
	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return "", e
	}
	rt := Result{}
	err := json.Unmarshal(body, &rt)
	if err != nil {
		return "", err
	}
	return rt.Id, nil
}

func startContainer(containerid string) error {
	cli := createClient()
	resp, e := cli.Post(fmt.Sprintf("http://v1.24/containers/%s/start", containerid), "application/json", nil)
	defer resp.Body.Close()
	_, e = ioutil.ReadAll(resp.Body)
	if e != nil {
		return e
	}
	if resp.StatusCode != 204 {
		return fmt.Errorf("return code is %d,not 204", resp.StatusCode)
	}
	return nil
}
func deleteContainer(containerid string) error {
	cli := createClient()
	request, e := http.NewRequest("DELETE", fmt.Sprintf("http://v1.24/containers/%s", containerid), nil)
	resp, e := cli.Do(request)
	defer resp.Body.Close()
	_, e = ioutil.ReadAll(resp.Body)
	if e != nil {
		return e
	}
	return nil
}

func createNetwork() (string, error) {
	cli := createClient()
	resp, e := cli.Post("http://v1.24/networks/create", "application/json", bytes.NewBuffer(createNet))
	defer resp.Body.Close()
	body, e := ioutil.ReadAll(resp.Body)
	if e != nil {
		return "", e
	}
	rt := Result{}
	err := json.Unmarshal(body, &rt)
	if err != nil {
		return "", err
	}
	return rt.Id, nil
}

func deleteNetwork(netid string) error {
	cli := createClient()
	request, e := http.NewRequest("DELETE", fmt.Sprintf("http://v1.24/networks/%s", netid), nil)
	resp, e := cli.Do(request)
	defer resp.Body.Close()
	_, e = ioutil.ReadAll(resp.Body)
	if e != nil {
		return e
	}
	return nil
}

/*
func main() {
	createContainer()
	//startContainer()
}

*/
