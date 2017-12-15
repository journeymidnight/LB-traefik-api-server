package integrate

import (
	"bytes"
	"encoding/json"
	"fmt"
	gocheck "gopkg.in/check.v1"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

func Test(t *testing.T) {
	gocheck.TestingT(t)
}

type APISuite struct {
	netid     string
	etcdid    string
	traefikid string
	whoami1id string
	whoami2id string
	apiid     string
}

var _ = gocheck.Suite(&APISuite{})

func (s *APISuite) SetUpSuite(c *gocheck.C) {
	var err error
	s.netid, err = createNetwork()
	c.Assert(err, gocheck.Equals, nil)
}

func (s *APISuite) SetUpTest(c *gocheck.C) {
	var err error
	s.etcdid, err = createContainer("etcd.tmpl")
	c.Assert(err, gocheck.Equals, nil)
	err = startContainer(s.etcdid)
	c.Assert(err, gocheck.Equals, nil)

	s.traefikid, err = createContainer("traefik.tmpl")
	c.Assert(err, gocheck.Equals, nil)
	err = startContainer(s.traefikid)
	c.Assert(err, gocheck.Equals, nil)

	s.whoami1id, err = createContainer("whoami1.tmpl")
	c.Assert(err, gocheck.Equals, nil)
	err = startContainer(s.whoami1id)
	c.Assert(err, gocheck.Equals, nil)

	s.whoami2id, err = createContainer("whoami2.tmpl")
	c.Assert(err, gocheck.Equals, nil)
	err = startContainer(s.whoami2id)
	c.Assert(err, gocheck.Equals, nil)

	s.apiid, err = createContainer("api.tmpl")
	c.Assert(err, gocheck.Equals, nil)
	err = startContainer(s.apiid)
	c.Assert(err, gocheck.Equals, nil)
	time.Sleep(2 * time.Second)
}

func (s *APISuite) TearDownSuite(c *gocheck.C) {
	var err error
	err = deleteNetwork(s.netid)
	c.Assert(err, gocheck.Equals, nil)
	time.Sleep(1 * time.Second)
}

func (s *APISuite) TearDownTest(c *gocheck.C) {
	var err error
	err = deleteContainer(s.etcdid)
	c.Assert(err, gocheck.Equals, nil)
	err = deleteContainer(s.traefikid)
	c.Assert(err, gocheck.Equals, nil)
	err = deleteContainer(s.whoami1id)
	c.Assert(err, gocheck.Equals, nil)
	err = deleteContainer(s.whoami2id)
	c.Assert(err, gocheck.Equals, nil)
	err = deleteContainer(s.apiid)
	c.Assert(err, gocheck.Equals, nil)
}
func (s *APISuite) TestListService(c *gocheck.C) {
	rt := &struct{ Ecode string }{}
	resp, err := http.Get(fmt.Sprintf("http://172.20.10.200/api/v1/services"))
	time.Sleep(1 * time.Second)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	c.Assert(err, gocheck.Equals, nil)
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, gocheck.Equals, nil)
	err = json.Unmarshal(body, rt)
	c.Assert(err, gocheck.Equals, nil)
	c.Assert(rt.Ecode, gocheck.Equals, "")
}

func (s *APISuite) TestListAPI(c *gocheck.C) {
	resp, err := http.Get(fmt.Sprintf("http://172.20.10.200"))
	time.Sleep(1 * time.Second)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	c.Assert(err, gocheck.Equals, nil)
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, gocheck.Equals, nil)
	c.Assert(strings.Contains(string(body), "api/v1"), gocheck.Equals, true)
}

func (s *APISuite) TestCreateService(c *gocheck.C) {
	time.Sleep(1 * time.Second)
	data, e := ioutil.ReadFile("jsons/basic.json")
	c.Assert(e, gocheck.Equals, nil)
	resp, e := http.Post("http://172.20.10.200/api/v1/services/basic.com", "application/json", bytes.NewBuffer(data))
	time.Sleep(1 * time.Second)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	rt := &struct{ Ecode int64 }{}
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body), "aaa")
	c.Assert(err, gocheck.Equals, nil)
	err = json.Unmarshal(body, rt)
	c.Assert(err, gocheck.Equals, nil)
	c.Assert(rt.Ecode, gocheck.Equals, int64(0))
	client := &http.Client{}
	request, e := http.NewRequest("Get", "http://172.20.10.101", nil)
	c.Assert(e, gocheck.Equals, nil)
	request.Host = "basic.com"
	resp, e = client.Do(request)
	c.Assert(e, gocheck.Equals, nil)
	body, e = ioutil.ReadAll(resp.Body)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	fmt.Println("body:", string(body))
}

func (s *APISuite) TestAddService(c *gocheck.C) {
	time.Sleep(1 * time.Second)
	data, e := ioutil.ReadFile("jsons/single111.json")
	c.Assert(e, gocheck.Equals, nil)
	resp, e := http.Post("http://172.20.10.200/api/v1/services/single111.com", "application/json", bytes.NewBuffer(data))
	time.Sleep(1 * time.Second)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	rt := &struct{ Ecode int64 }{}
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body), "aaa")
	c.Assert(err, gocheck.Equals, nil)
	err = json.Unmarshal(body, rt)
	c.Assert(err, gocheck.Equals, nil)
	c.Assert(rt.Ecode, gocheck.Equals, int64(0))
	client := &http.Client{}
	request, e := http.NewRequest("Get", "http://172.20.10.101", nil)
	c.Assert(e, gocheck.Equals, nil)
	request.Host = "single111.com"
	resp, e = client.Do(request)
	c.Assert(e, gocheck.Equals, nil)
	body, e = ioutil.ReadAll(resp.Body)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	fmt.Println("body:", string(body))

	data, e = ioutil.ReadFile("jsons/single112.json")
	c.Assert(e, gocheck.Equals, nil)
	resp, e = http.Post("http://172.20.10.200/api/v1/services/single112.com", "application/json", bytes.NewBuffer(data))
	time.Sleep(1 * time.Second)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	rt = &struct{ Ecode int64 }{}
	body, err = ioutil.ReadAll(resp.Body)
	fmt.Println(string(body), "aaa")
	c.Assert(err, gocheck.Equals, nil)
	err = json.Unmarshal(body, rt)
	c.Assert(err, gocheck.Equals, nil)
	c.Assert(rt.Ecode, gocheck.Equals, int64(0))
	client = &http.Client{}
	request, e = http.NewRequest("Get", "http://172.20.10.101", nil)
	c.Assert(e, gocheck.Equals, nil)
	request.Host = "single112.com"
	resp, e = client.Do(request)
	c.Assert(e, gocheck.Equals, nil)
	body, e = ioutil.ReadAll(resp.Body)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	c.Assert(strings.Contains(string(body), "172.20.10.112"), gocheck.Equals, true)
}
