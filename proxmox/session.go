package proxmox

// inspired by https://github.com/openstack/golang-client/blob/master/openstack/session.go

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var Debug = new(bool)

type Response struct {
	Resp *http.Response
	Body []byte
}

type Session struct {
	httpClient *http.Client
	ApiUrl     string
	AuthTicket string
	CsrfToken  string
	Headers    http.Header
}

func NewSession(apiUrl string, hclient *http.Client, tls *tls.Config) (session *Session, err error) {
	if hclient == nil {
		// Only build a transport if we're also building the client
		tr := &http.Transport{
			TLSClientConfig:    tls,
			DisableCompression: true,
		}
		hclient = &http.Client{Transport: tr}
	}
	session = &Session{
		httpClient: hclient,
		ApiUrl:     apiUrl,
		AuthTicket: "",
		CsrfToken:  "",
		Headers:    http.Header{},
	}
	return session, nil
}

func (s *Session) Login(username string, password string) (err error) {
	v := url.Values{}
	v.Set("username", username)
	v.Set("password", password)
	reqh := http.Header{}
	reqh.Add("Content-Type", "application/x-www-form-urlencoded")
	reqbody := bytes.NewBufferString(v.Encode()).Bytes()
	olddebug := *Debug
	*Debug = false // don't share passwords in debug log
	resp, err := s.Request("POST", "/access/ticket", nil, &reqh, &reqbody)
	*Debug = olddebug
	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New("error reading response body")
	}
	var jbody map[string]interface{}
	if err = json.Unmarshal(rbody, &jbody); err != nil {
		return
	}

	dat := jbody["data"].(map[string]interface{})
	s.AuthTicket = dat["ticket"].(string)
	s.CsrfToken = dat["CSRFPreventionToken"].(string)
	return nil
}

func (s *Session) NewRequest(method, url string, headers *http.Header, body io.Reader) (req *http.Request, err error) {
	req, err = http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if headers != nil {
		req.Header = *headers
	}
	if s.AuthTicket != "" {
		req.Header.Add("Cookie", "PVEAuthCookie="+s.AuthTicket)
		req.Header.Add("CSRFPreventionToken", s.CsrfToken)
	}
	return
}

func (s *Session) Do(req *http.Request) (*http.Response, error) {
	// Add session headers
	for k := range s.Headers {
		req.Header.Set(k, s.Headers.Get(k))
	}

	if *Debug {
		d, _ := httputil.DumpRequestOut(req, true)
		log.Printf(">>>>>>>>>> REQUEST:\n", string(d))
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if *Debug {
		dr, _ := httputil.DumpResponse(resp, true)
		log.Printf("<<<<<<<<<< RESULT:\n", string(dr))
	}

	return resp, nil
}

// Perform a simple get to an endpoint
func (s *Session) Request(
	method string,
	url string,
	params *url.Values,
	headers *http.Header,
	body *[]byte,
) (resp *http.Response, err error) {
	// add params to url here
	url = s.ApiUrl + url
	if params != nil {
		url = url + "?" + params.Encode()
	}

	// Get the body if one is present
	var buf io.Reader
	if body != nil {
		buf = bytes.NewReader(*body)
	}

	req, err := s.NewRequest(method, url, headers, buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err = s.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Perform a simple get to an endpoint and unmarshall returned JSON
func (s *Session) RequestJSON(
	method string,
	url string,
	params *url.Values,
	headers *http.Header,
	body interface{},
	responseContainer interface{},
) (resp *http.Response, err error) {
	var bodyjson []byte
	if body != nil {
		bodyjson, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	// if headers == nil {
	// 	headers = &http.Header{}
	// 	headers.Add("Content-Type", "application/json")
	// }

	resp, err = s.Request(method, url, params, headers, &bodyjson)
	if err != nil {
		return nil, err
	}

	// err = util.CheckHTTPResponseStatusCode(resp)
	// if err != nil {
	// 	return nil, err
	// }

	rbody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("error reading response body")
	}
	if err = json.Unmarshal(rbody, &responseContainer); err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *Session) Delete(
	url string,
	params *url.Values,
	headers *http.Header,
) (resp *http.Response, err error) {
	return s.Request("DELETE", url, params, headers, nil)
}

func (s *Session) Get(
	url string,
	params *url.Values,
	headers *http.Header,
) (resp *http.Response, err error) {
	return s.Request("GET", url, params, headers, nil)
}

func (s *Session) GetJSON(
	url string,
	params *url.Values,
	headers *http.Header,
	responseContainer interface{},
) (resp *http.Response, err error) {
	return s.RequestJSON("GET", url, params, headers, nil, responseContainer)
}

func (s *Session) Head(
	url string,
	params *url.Values,
	headers *http.Header,
) (resp *http.Response, err error) {
	return s.Request("HEAD", url, params, headers, nil)
}

func (s *Session) Post(
	url string,
	params *url.Values,
	headers *http.Header,
	body *[]byte,
) (resp *http.Response, err error) {
	return s.Request("POST", url, params, headers, body)
}

func (s *Session) PostJSON(
	url string,
	params *url.Values,
	headers *http.Header,
	body interface{},
	responseContainer interface{},
) (resp *http.Response, err error) {
	return s.RequestJSON("POST", url, params, headers, body, responseContainer)
}

func (s *Session) Put(
	url string,
	params *url.Values,
	headers *http.Header,
	body *[]byte,
) (resp *http.Response, err error) {
	return s.Request("PUT", url, params, headers, body)
}
