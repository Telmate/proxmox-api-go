package proxmox

// inspired by https://github.com/openstack-archive/golang-client/blob/f8471e433432b26dd29aa1c1cd42317a9d79a551/openstack/session.go

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"slices"
	"strings"
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
		case bool: // TODO we shouldn't to this here, but when we do the initial serialization
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

func (s *Session) setAPIToken(token ApiTokenID, secret ApiTokenSecret) {
	s.AuthToken = token.String() + "=" + secret.String()
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
	reqbody := paramsToBody(reqUser)
	olddebug := *Debug
	*Debug = false // don't share passwords in debug log
	resp, _, err := s.post(ctx, "/access/ticket", nil, &s.Headers, &reqbody)
	*Debug = olddebug
	if err != nil {
		return err
	}
	if resp == nil {
		return fmt.Errorf("login error reading response")
	}
	dr, _ := httputil.DumpResponse(resp, true)
	jbody, err := responseJSON(resp)
	if err != nil {
		return err
	}
	if jbody == nil || jbody["data"] == nil {
		return fmt.Errorf("invalid login response:\n-----\n%s\n-----", dr)
	}
	dat := jbody["data"].(map[string]interface{})
	//Check if the 2FA was required
	if dat["NeedTFA"] == 1.0 {
		return fmt.Errorf("missing TFA code")
	}
	s.AuthTicket = dat["ticket"].(string)
	s.CsrfToken = dat["CSRFPreventionToken"].(string)
	return nil
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

func (s *Session) do(req *http.Request) (resp *http.Response, retry bool, err error) {
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

	resp, err = s.httpClient.Do(req)
	if err != nil {
		return nil, true, &errorWrap{
			err:     err,
			message: "error performing http request",
		}
	}

	// The response body reader needs to be closed, but lots of places call
	// session.do, and they might not be able to reliably close it themselves.
	// Therefore, read the body out, close the original, then replace it with
	// a NopCloser over the bytes, which does not need to be closed downsteam.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, &errorWrap{
			err:     err,
			message: "error reading response body",
		}
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
		if len(respBody) > 1 { // we need at least {} to be valid JSON, sometimes a single `\n` is returned.
			var bodyObj map[string]any
			err = json.Unmarshal(respBody, &bodyObj)
			if err == nil {
				apiErr := ApiError{
					Code: resp.Status[0:3],
				}
				if v, ok := bodyObj["errors"]; ok {
					apiErr.Errors = v.(map[string]any)
				}
				if v, ok := bodyObj["message"]; ok {
					apiErr.Message = strings.TrimRight(v.(string), "\n")
				}
				return resp, false, &apiErr
			}
		}
		return resp, true, errors.New(resp.Status)
	}

	return resp, false, nil
}

// Perform a simple get to an endpoint
func (s *Session) request(
	ctx context.Context,
	method string,
	url string,
	params *url.Values,
	headers *http.Header,
	body *[]byte,
) (resp *http.Response, retry bool, err error) {
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
		return nil, true, err
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
	body *[]byte,
	responseContainer interface{},
) (resp *http.Response, retry bool, err error) {
	// if headers == nil {
	// 	headers = &http.Header{}
	// 	headers.Add("Content-Type", "application/json")
	// }

	resp, retry, err = s.request(ctx, method, url, params, headers, body)
	if err != nil {
		return resp, retry, err
	}

	// err = util.CheckHTTPResponseStatusCode(resp)
	// if err != nil {
	// 	return nil, err
	// }

	rbody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, true, fmt.Errorf("error reading response body")
	}
	if err = json.Unmarshal(rbody, &responseContainer); err != nil {
		return resp, true, err
	}

	return resp, false, nil
}

func (s *Session) delete(
	ctx context.Context,
	url string,
	params *url.Values,
	headers *http.Header,
) (resp *http.Response, retry bool, err error) {
	return s.request(ctx, "DELETE", url, params, headers, nil)
}

func (s *Session) get(
	ctx context.Context,
	url string,
	params *url.Values,
	headers *http.Header,
) (resp *http.Response, retry bool, err error) {
	return s.request(ctx, "GET", url, params, headers, nil)
}

func (s *Session) getJSON(
	ctx context.Context,
	url string,
	params *url.Values,
	headers *http.Header,
	responseContainer interface{},
) (resp *http.Response, retry bool, err error) {
	return s.requestJSON(ctx, "GET", url, params, headers, nil, responseContainer)
}

func (s *Session) post(
	ctx context.Context,
	url string,
	params *url.Values,
	headers *http.Header,
	body *[]byte,
) (resp *http.Response, retry bool, err error) {
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
	body *[]byte,
	responseContainer interface{},
) (resp *http.Response, retry bool, err error) {
	return s.requestJSON(ctx, "POST", url, params, headers, body, responseContainer)
}

func (s *Session) put(
	ctx context.Context,
	url string,
	params *url.Values,
	headers *http.Header,
	body *[]byte,
) (resp *http.Response, retry bool, err error) {
	if headers == nil {
		headers = &http.Header{}
		headers.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	return s.request(ctx, "PUT", url, params, headers, body)
}
