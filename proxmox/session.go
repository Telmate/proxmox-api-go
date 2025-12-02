package proxmox

// inspired by https://github.com/openstack-archive/golang-client/blob/f8471e433432b26dd29aa1c1cd42317a9d79a551/openstack/session.go

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"slices"
)

var Debug = new(bool)

const debugLargeBodyThreshold = 5 * 1024 * 1024

type Session struct {
	httpClient *http.Client
	ApiUrl     string
	AuthTicket string
	CsrfToken  string
	AuthToken  string // Combination of user, realm, token ID and UUID
	Headers    http.Header
}

func NewSession(apiUrl string, hclient *http.Client, proxyString string, tls *tls.Config) (session *Session, err error) {
	if hclient == nil {
		if proxyString == "" {
			tr := &http.Transport{
				TLSClientConfig:    tls,
				DisableCompression: true,
				Proxy:              nil,
			}
			hclient = &http.Client{Transport: tr}
		} else {
			proxyURL, err := url.ParseRequestURI(proxyString)
			if err != nil {
				return nil, err
			}
			if _, _, err := net.SplitHostPort(proxyURL.Host); err != nil {
				return nil, err
			} else {
				// Only build a transport if we're also building the client
				tr := &http.Transport{
					TLSClientConfig:    tls,
					DisableCompression: true,
					Proxy:              http.ProxyURL(proxyURL),
				}
				hclient = &http.Client{Transport: tr}
			}
		}
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

func paramsToBody(params map[string]interface{}) (body []byte) {
	vals := paramsToValuesWithEmpty(params, []string{})
	body = bytes.NewBufferString(vals.Encode()).Bytes()
	return
}

func paramsToValues(params map[string]interface{}) (vals url.Values) {
	vals = paramsToValuesWithEmpty(params, []string{})
	return
}

func paramsToBodyWithEmpty(params map[string]interface{}, allowedEmpty []string) (body []byte) {
	vals := paramsToValuesWithEmpty(params, allowedEmpty)
	body = bytes.NewBufferString(vals.Encode()).Bytes()
	return
}

func paramsToBodyWithAllEmpty(params map[string]interface{}) (body []byte) {
	vals := paramsToValuesWithAllEmpty(params, []string{}, true)
	body = bytes.NewBufferString(vals.Encode()).Bytes()
	return
}

func paramsToValuesWithEmpty(params map[string]interface{}, allowedEmpty []string) (vals url.Values) {
	return paramsToValuesWithAllEmpty(params, allowedEmpty, false)
}

func paramsToValuesWithAllEmpty(params map[string]interface{}, allowedEmpty []string, allowEmpty bool) (vals url.Values) {
	vals = url.Values{}
	for k, intrV := range params {
		var v string
		switch intrV := intrV.(type) {
		// Convert true/false bool to 1/0 string where Proxmox API can understand it.
		case bool:
			if intrV {
				v = "1"
			} else {
				v = "0"
			}
		case []string:
			for _, v := range intrV {
				vals.Add(k, fmt.Sprintf("%v", v))
			}
			continue
		default:
			v = fmt.Sprintf("%v", intrV)
		}
		if allowEmpty {
			vals.Set(k, v)
		} else if v != "" || slices.Contains(allowedEmpty, k) {
			vals.Set(k, v)
		}
	}
	return
}

func decodeResponse(resp *http.Response, v interface{}) error {
	if resp.Body == nil {
		return nil
	}
	rbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response body: %s", err)
	}
	if err = json.Unmarshal(rbody, &v); err != nil {
		return err
	}
	return nil
}

func responseJSON(resp *http.Response) (jbody map[string]interface{}, err error) {
	err = decodeResponse(resp, &jbody)
	return jbody, err
}

func (s *Session) setAPIToken(userID, token string) {
	auth := fmt.Sprintf("%s=%s", userID, token)
	s.AuthToken = auth
}

func (s *Session) setTicket(ticket, csrfPreventionToken string) {
	s.AuthTicket = ticket
	s.CsrfToken = csrfPreventionToken
}

func (s *Session) login(ctx context.Context, username string, password string, otp string) (err error) {
	reqUser := map[string]interface{}{"username": username, "password": password}
	if otp != "" {
		reqUser["otp"] = otp
	}

	loginResp, status, err := s.loginRequest(ctx, reqUser)
	if err != nil && status != http.StatusUnauthorized {
		return err
	}

	// If the first try with OTP was rejected, retry without OTP to discover TFA challenge
	if err != nil && status == http.StatusUnauthorized && otp != "" {
		originalErr := err
		reqUser = map[string]interface{}{"username": username, "password": password}
		loginResp, status, err = s.loginRequest(ctx, reqUser)

		// If the retry also fails, return the original error
		if err != nil {
			return originalErr
		}
	}

	// Two-step TOTP flow when server signals NeedTFA
	if loginResp.needTFA {
		if otp == "" {
			return fmt.Errorf("missing TFA code")
		}
		if loginResp.ticket == "" {
			return fmt.Errorf("two-step login: missing challenge ticket in first response")
		}
		secondReq := map[string]interface{}{
			"username":      username,
			"tfa-challenge": loginResp.ticket,
			"password":      fmt.Sprintf("totp:%s", otp),
		}
		loginResp, status, err = s.loginRequest(ctx, secondReq)
		if err != nil {
			return err
		}
	}

	s.AuthTicket = loginResp.ticket
	s.CsrfToken = loginResp.csrfToken
	return nil
}

type loginResponse struct {
	ticket    string
	csrfToken string
	needTFA   bool
}

func (s *Session) loginRequest(ctx context.Context, body map[string]interface{}) (loginResponse, int, error) {
	reqbody := paramsToBody(body)
	olddebug := *Debug
	*Debug = false // don't share passwords in debug log
	resp, err := s.post(ctx, "/access/ticket", nil, &s.Headers, &reqbody)
	*Debug = olddebug
	if err != nil {
		status := 0
		if resp != nil {
			status = resp.StatusCode
		}
		return loginResponse{}, status, err
	}
	if resp == nil {
		return loginResponse{}, 0, fmt.Errorf("login error reading response")
	}
	status := resp.StatusCode
	dr, _ := httputil.DumpResponse(resp, true)
	jbody, err := responseJSON(resp)
	if err != nil {
		return loginResponse{}, status, err
	}
	if jbody == nil || jbody["data"] == nil {
		return loginResponse{}, status, fmt.Errorf("invalid login response:\n-----\n%s\n-----", dr)
	}
	dat, ok := jbody["data"].(map[string]interface{})
	if !ok {
		return loginResponse{}, status, fmt.Errorf("invalid login response data")
	}

	loginResp := loginResponse{}

	if needTFA, ok := dat["NeedTFA"].(float64); ok && needTFA == 1.0 {
		loginResp.needTFA = true
	}
	if ticket, ok := dat["ticket"].(string); ok {
		loginResp.ticket = ticket
	}
	if csrf, ok := dat["CSRFPreventionToken"].(string); ok {
		loginResp.csrfToken = csrf
	}

	return loginResp, status, nil
}

func (s *Session) NewRequest(ctx context.Context, method, url string, headers *http.Header, body io.Reader) (req *http.Request, err error) {
	req, err = http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	if headers != nil {
		req.Header = *headers
	}
	if s.AuthToken != "" {
		req.Header["Authorization"] = []string{"PVEAPIToken=" + s.AuthToken}
	} else if s.AuthTicket != "" {
		req.Header["Authorization"] = []string{"PVEAuthCookie=" + s.AuthTicket}
		req.Header["CSRFPreventionToken"] = []string{s.CsrfToken}
	}
	return
}

func (s *Session) do(req *http.Request) (*http.Response, error) {
	// Add session headers
	for k, v := range s.Headers {
		req.Header[k] = v
	}

	if *Debug {
		includeBody := req.ContentLength < debugLargeBodyThreshold
		d, _ := httputil.DumpRequestOut(req, includeBody)
		if !includeBody {
			d = append(d, fmt.Sprintf("<request body of %d bytes not shown>\n\n", req.ContentLength)...)
		}
		log.Printf(">>>>>>>>>> REQUEST:\n%v", string(d))
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// The response body reader needs to be closed, but lots of places call
	// session.do, and they might not be able to reliably close it themselves.
	// Therefore, read the body out, close the original, then replace it with
	// a NopCloser over the bytes, which does not need to be closed downsteam.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(respBody))

	if *Debug {
		includeBody := resp.ContentLength < debugLargeBodyThreshold
		dr, _ := httputil.DumpResponse(resp, includeBody)
		if !includeBody {
			dr = append(dr, fmt.Sprintf("<response body of %d bytes not shown>\n\n", resp.ContentLength)...)
		}
		log.Printf("<<<<<<<<<< RESULT:\n%v", string(dr))
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return resp, fmt.Errorf(resp.Status)
	}

	return resp, nil
}

// Perform a simple get to an endpoint
func (s *Session) request(
	ctx context.Context,
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

	// get the body if one is present
	var buf io.Reader
	if body != nil {
		buf = bytes.NewReader(*body)
	}

	req, err := s.NewRequest(ctx, method, url, headers, buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	return s.do(req)
}

// Perform a simple get to an endpoint and unmarshal returned JSON
func (s *Session) requestJSON(
	ctx context.Context,
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

	resp, err = s.request(ctx, method, url, params, headers, &bodyjson)
	if err != nil {
		return resp, err
	}

	// err = util.CheckHTTPResponseStatusCode(resp)
	// if err != nil {
	// 	return nil, err
	// }

	rbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, fmt.Errorf("error reading response body")
	}
	if err = json.Unmarshal(rbody, &responseContainer); err != nil {
		return resp, err
	}

	return resp, nil
}

func (s *Session) delete(
	ctx context.Context,
	url string,
	params *url.Values,
	headers *http.Header,
) (resp *http.Response, err error) {
	return s.request(ctx, "DELETE", url, params, headers, nil)
}

func (s *Session) get(
	ctx context.Context,
	url string,
	params *url.Values,
	headers *http.Header,
) (resp *http.Response, err error) {
	return s.request(ctx, "GET", url, params, headers, nil)
}

func (s *Session) getJSON(
	ctx context.Context,
	url string,
	params *url.Values,
	headers *http.Header,
	responseContainer interface{},
) (resp *http.Response, err error) {
	return s.requestJSON(ctx, "GET", url, params, headers, nil, responseContainer)
}

func (s *Session) post(
	ctx context.Context,
	url string,
	params *url.Values,
	headers *http.Header,
	body *[]byte,
) (resp *http.Response, err error) {
	if headers == nil {
		headers = &http.Header{}
		headers.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	return s.request(ctx, "POST", url, params, headers, body)
}

func (s *Session) postJSON(
	ctx context.Context,
	url string,
	params *url.Values,
	headers *http.Header,
	body interface{},
	responseContainer interface{},
) (resp *http.Response, err error) {
	return s.requestJSON(ctx, "POST", url, params, headers, body, responseContainer)
}

func (s *Session) put(
	ctx context.Context,
	url string,
	params *url.Values,
	headers *http.Header,
	body *[]byte,
) (resp *http.Response, err error) {
	if headers == nil {
		headers = &http.Header{}
		headers.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	return s.request(ctx, "PUT", url, params, headers, body)
}
