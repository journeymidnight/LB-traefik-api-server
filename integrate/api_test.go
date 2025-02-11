package integrate

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	gocheck "gopkg.in/check.v1"
	"io/ioutil"
	"net/http"
	"regexp"
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
	cookie1id string
	cookie2id string
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

	s.cookie1id, err = createContainer("testcookie1.tmpl")
	c.Assert(err, gocheck.Equals, nil)
	err = startContainer(s.cookie1id)
	c.Assert(err, gocheck.Equals, nil)

	s.cookie2id, err = createContainer("testcookie2.tmpl")
	c.Assert(err, gocheck.Equals, nil)
	err = startContainer(s.cookie2id)
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
	err = deleteContainer(s.cookie1id)
	c.Assert(err, gocheck.Equals, nil)
	err = deleteContainer(s.cookie2id)
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

	data, e = ioutil.ReadFile("jsons/single112.json")
	c.Assert(e, gocheck.Equals, nil)
	resp, e = http.Post("http://172.20.10.200/api/v1/services/single112.com", "application/json", bytes.NewBuffer(data))
	time.Sleep(1 * time.Second)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	rt = &struct{ Ecode int64 }{}
	body, err = ioutil.ReadAll(resp.Body)
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

func (s *APISuite) TestUpdateServices(c *gocheck.C) {
	time.Sleep(1 * time.Second)
	data, e := ioutil.ReadFile("jsons/single111.json")
	c.Assert(e, gocheck.Equals, nil)
	resp, e := http.Post("http://172.20.10.200/api/v1/services/testupdate.com", "application/json", bytes.NewBuffer(data))
	time.Sleep(1 * time.Second)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	rt := &struct{ Ecode int64 }{}
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, gocheck.Equals, nil)
	err = json.Unmarshal(body, rt)
	c.Assert(err, gocheck.Equals, nil)
	c.Assert(rt.Ecode, gocheck.Equals, int64(0))
	client := &http.Client{}
	request, e := http.NewRequest("Get", "http://172.20.10.101", nil)
	c.Assert(e, gocheck.Equals, nil)
	request.Host = "testupdate.com"
	resp, e = client.Do(request)
	c.Assert(e, gocheck.Equals, nil)
	body, e = ioutil.ReadAll(resp.Body)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	c.Assert(strings.Contains(string(body), "172.20.10.111"), gocheck.Equals, true)

	data, e = ioutil.ReadFile("jsons/single112.json")
	c.Assert(e, gocheck.Equals, nil)
	client = &http.Client{}
	request, e = http.NewRequest("PUT", "http://172.20.10.200/api/v1/services/testupdate.com", bytes.NewBuffer(data))
	c.Assert(e, gocheck.Equals, nil)
	resp, e = client.Do(request)
	time.Sleep(1 * time.Second)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	rt = &struct{ Ecode int64 }{}
	body, err = ioutil.ReadAll(resp.Body)
	c.Assert(err, gocheck.Equals, nil)
	err = json.Unmarshal(body, rt)
	c.Assert(err, gocheck.Equals, nil)
	c.Assert(rt.Ecode, gocheck.Equals, int64(0))
	request, e = http.NewRequest("Get", "http://172.20.10.101", nil)
	c.Assert(e, gocheck.Equals, nil)
	request.Host = "testupdate.com"
	resp, e = client.Do(request)
	c.Assert(e, gocheck.Equals, nil)
	body, e = ioutil.ReadAll(resp.Body)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	c.Assert(strings.Contains(string(body), "172.20.10.112"), gocheck.Equals, true)
}

func (s *APISuite) TestDeleteServices(c *gocheck.C) {
	time.Sleep(1 * time.Second)
	data, e := ioutil.ReadFile("jsons/single111.json")
	c.Assert(e, gocheck.Equals, nil)
	resp, e := http.Post("http://172.20.10.200/api/v1/services/testdelete.com", "application/json", bytes.NewBuffer(data))
	time.Sleep(1 * time.Second)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	rt := &struct{ Ecode int64 }{}
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, gocheck.Equals, nil)
	err = json.Unmarshal(body, rt)
	c.Assert(err, gocheck.Equals, nil)
	c.Assert(rt.Ecode, gocheck.Equals, int64(0))
	client := &http.Client{}
	request, e := http.NewRequest("Get", "http://172.20.10.101", nil)
	c.Assert(e, gocheck.Equals, nil)
	request.Host = "testdelete.com"
	resp, e = client.Do(request)
	c.Assert(e, gocheck.Equals, nil)
	body, e = ioutil.ReadAll(resp.Body)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	c.Assert(strings.Contains(string(body), "172.20.10.111"), gocheck.Equals, true)

	client = &http.Client{}
	request, e = http.NewRequest("DELETE", "http://172.20.10.200/api/v1/services/testdelete.com", nil)
	c.Assert(e, gocheck.Equals, nil)
	resp, e = client.Do(request)
	time.Sleep(1 * time.Second)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	rt = &struct{ Ecode int64 }{}
	body, err = ioutil.ReadAll(resp.Body)
	c.Assert(err, gocheck.Equals, nil)
	err = json.Unmarshal(body, rt)
	c.Assert(err, gocheck.Equals, nil)
	c.Assert(rt.Ecode, gocheck.Equals, int64(0))
	request, e = http.NewRequest("Get", "http://172.20.10.101", nil)
	c.Assert(e, gocheck.Equals, nil)
	request.Host = "testdelete.com"
	resp, e = client.Do(request)
	c.Assert(e, gocheck.Equals, nil)
	body, e = ioutil.ReadAll(resp.Body)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 404)
}

func (s *APISuite) TestHealthCheck(c *gocheck.C) {
	time.Sleep(1 * time.Second)
	data, e := ioutil.ReadFile("jsons/health.json")
	c.Assert(e, gocheck.Equals, nil)
	resp, e := http.Post("http://172.20.10.200/api/v1/services/testhealth.com", "application/json", bytes.NewBuffer(data))
	time.Sleep(1 * time.Second)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	rt := &struct{ Ecode int64 }{}
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, gocheck.Equals, nil)
	err = json.Unmarshal(body, rt)
	c.Assert(err, gocheck.Equals, nil)
	c.Assert(rt.Ecode, gocheck.Equals, int64(0))
	client := &http.Client{}
	request, e := http.NewRequest("Get", "http://172.20.10.101", nil)
	c.Assert(e, gocheck.Equals, nil)
	request.Host = "testhealth.com"
	resp, e = client.Do(request)
	c.Assert(e, gocheck.Equals, nil)
	body, e = ioutil.ReadAll(resp.Body)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)

	e = stopContainer(s.whoami1id)
	c.Assert(e, gocheck.Equals, nil)
	time.Sleep(20 * time.Second)
	for i := 0; i < 5; i++ {
		request, e = http.NewRequest("Get", "http://172.20.10.101", nil)
		c.Assert(e, gocheck.Equals, nil)
		request.Host = "testhealth.com"
		resp, e = client.Do(request)
		c.Assert(e, gocheck.Equals, nil)
		body, e = ioutil.ReadAll(resp.Body)
		c.Assert(e, gocheck.Equals, nil)
		c.Assert(resp.StatusCode, gocheck.Equals, 200)
		c.Assert(strings.Contains(string(body), "172.20.10.112"), gocheck.Equals, true)
		resp.Body.Close()
	}
}

func (s *APISuite) TestStickiness(c *gocheck.C) {
	time.Sleep(1 * time.Second)
	data, e := ioutil.ReadFile("jsons/stickiness.json")
	c.Assert(e, gocheck.Equals, nil)
	resp, e := http.Post("http://172.20.10.200/api/v1/services/teststick.com", "application/json", bytes.NewBuffer(data))
	time.Sleep(1 * time.Second)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	rt := &struct{ Ecode int64 }{}
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, gocheck.Equals, nil)
	err = json.Unmarshal(body, rt)
	c.Assert(err, gocheck.Equals, nil)
	c.Assert(rt.Ecode, gocheck.Equals, int64(0))

	time.Sleep(10 * time.Second)
	client := &http.Client{}
	request, e := http.NewRequest("Get", "http://172.20.10.101", nil)
	c.Assert(e, gocheck.Equals, nil)
	request.Host = "teststick.com"
	resp, e = client.Do(request)
	c.Assert(e, gocheck.Equals, nil)
	body, e = ioutil.ReadAll(resp.Body)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	rtcook := resp.Cookies()[0]

	r, e := regexp.Compile("172.20.10.11[3-4]")
	c.Assert(e, gocheck.Equals, nil)
	ip := r.FindString(string(body))
	for i := 0; i < 5; i++ {
		request, e = http.NewRequest("Get", "http://172.20.10.101/display", nil)
		request.AddCookie(rtcook)
		c.Assert(e, gocheck.Equals, nil)
		request.Host = "teststick.com"
		resp, e = client.Do(request)
		c.Assert(e, gocheck.Equals, nil)
		body, e = ioutil.ReadAll(resp.Body)
		c.Assert(e, gocheck.Equals, nil)
		c.Assert(resp.StatusCode, gocheck.Equals, 200)
		c.Assert(strings.Contains(string(body), ip), gocheck.Equals, true)
	}
}

type Certs struct {
	CertFile string
	KeyFile  string
}

func (s *APISuite) TestCreateAndDeleteCert(c *gocheck.C) {
	time.Sleep(1 * time.Second)
	data, e := ioutil.ReadFile("jsons/basic.json")
	c.Assert(e, gocheck.Equals, nil)
	resp, e := http.Post("http://172.20.10.200/api/v1/services/www.testapi.org", "application/json", bytes.NewBuffer(data))
	time.Sleep(1 * time.Second)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	rt := &struct{ Ecode int64 }{}
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, gocheck.Equals, nil)
	err = json.Unmarshal(body, rt)
	c.Assert(err, gocheck.Equals, nil)
	c.Assert(rt.Ecode, gocheck.Equals, int64(0))

	orgcert, err := ioutil.ReadFile("certs/testapi.org.crt")
	c.Assert(err, gocheck.Equals, nil)
	orgkey, err := ioutil.ReadFile("certs/testapi.org")
	c.Assert(err, gocheck.Equals, nil)
	scert := Certs{CertFile: string(orgcert), KeyFile: string(orgkey)}
	sjson, err := json.Marshal(scert)
	c.Assert(err, gocheck.Equals, nil)
	resp, e = http.Post("http://172.20.10.200/api/v1/certs/www.testapi.org", "application/json", bytes.NewBuffer(sjson))
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	body, err = ioutil.ReadAll(resp.Body)
	c.Assert(err, gocheck.Equals, nil)
	time.Sleep(time.Second * 1)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true, ServerName: "www.testapi.org"},
	}
	client := &http.Client{Transport: tr}
	request, e := http.NewRequest("Get", "https://172.20.10.101", nil)
	c.Assert(e, gocheck.Equals, nil)
	request.Host = "www.testapi.org"
	request.Header.Set("Host", "www.testapi.org")
	request.Header.Set("Accept", "*/*")
	resp, e = client.Do(request)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	body, err = ioutil.ReadAll(resp.Body)
	c.Assert(err, gocheck.Equals, nil)
	c.Assert(resp.TLS.PeerCertificates[0].Subject.CommonName, gocheck.Equals, "www.testapi.org")

	//test delete cert

	time.Sleep(2 * time.Second)
	request, e = http.NewRequest("DELETE", "http://172.20.10.200/api/v1/certs/www.testapi.org", nil)
	c.Assert(e, gocheck.Equals, nil)
	simplecli := &http.Client{}
	resp, e = simplecli.Do(request)
	c.Assert(e, gocheck.Equals, nil)
	body, e = ioutil.ReadAll(resp.Body)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	time.Sleep(2 * time.Second)

	req, e := http.NewRequest("Get", "https://172.20.10.101", nil)
	c.Assert(e, gocheck.Equals, nil)
	req.Host = "www.testapi.org"
	req.Header.Set("Host", "www.testapi.org")
	req.Header.Set("Accept", "*/*")
	tr1 := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true, ServerName: "www.testapi.org"},
	}
	cli := &http.Client{Transport: tr1}
	res, e := cli.Do(req)
	c.Assert(e, gocheck.Equals, nil)
	c.Assert(res.StatusCode, gocheck.Equals, 200)
	body, err = ioutil.ReadAll(res.Body)
	c.Assert(err, gocheck.Equals, nil)
	c.Assert(res.TLS.PeerCertificates[0].Subject.CommonName, gocheck.Equals, "TRAEFIK DEFAULT CERT")
}
