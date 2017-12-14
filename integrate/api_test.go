package integrate

import (
	//	"encoding/json"
	"fmt"
	gocheck "gopkg.in/check.v1"
	"io/ioutil"
	"net/http"
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
}

func (s *APISuite) TearDownSuite(c *gocheck.C) {
	var err error
	err = deleteNetwork(s.netid)
	c.Assert(err, gocheck.Equals, nil)
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
	//	rt := struct{ Ecode string }{}
	//	fmt.Println(rt)
	time.Sleep(2 * time.Second)
	resp, err := http.Get(fmt.Sprintf("http://172.20.10.200/api/v1/services"))
	//	c.Assert(resp.StatusCode, gocheck.Equals, 200)
	c.Assert(err, gocheck.Equals, nil)
	body, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, gocheck.Equals, nil)
	//err = json.Unmarshal(body, rt)
	c.Assert(err, gocheck.Equals, nil)
	//fmt.Println(rt)
	//	c.Assert(rt.Ecode, gocheck.Equals, "")
	fmt.Println(string(body))
}
