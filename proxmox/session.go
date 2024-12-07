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
	"strings"
	"sync"
	"time"
)

var Debug = new(bool)

const DebugLargeBodyThreshold = 5 * 1024 * 1024

type Response struct {
	Resp *http.Response
	Body []byte
}

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

func ParamsToBody(params map[string]interface{}) (body []byte) {
	vals := ParamsToValuesWithEmpty(params, []string{})
	body = bytes.NewBufferString(vals.Encode()).Bytes()
	return
}

func ParamsToValues(params map[string]interface{}) (vals url.Values) {
	vals = ParamsToValuesWithEmpty(params, []string{})
	return
}

func ParamsToBodyWithEmpty(params map[string]interface{}, allowedEmpty []string) (body []byte) {
	vals := ParamsToValuesWithEmpty(params, allowedEmpty)
	body = bytes.NewBufferString(vals.Encode()).Bytes()
	return
}

func ParamsToBodyWithAllEmpty(params map[string]interface{}) (body []byte) {
	vals := ParamsToValuesWithAllEmpty(params, []string{}, true)
	body = bytes.NewBufferString(vals.Encode()).Bytes()
	return
}

func ParamsToValuesWithEmpty(params map[string]interface{}, allowedEmpty []string) (vals url.Values) {
	return ParamsToValuesWithAllEmpty(params, allowedEmpty, false)
}

func ParamsToValuesWithAllEmpty(params map[string]interface{}, allowedEmpty []string, allowEmpty bool) (vals url.Values) {
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
		} else if v != "" || inArray(allowedEmpty, k) {
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

func ResponseJSON(resp *http.Response) (jbody map[string]interface{}, err error) {
	err = decodeResponse(resp, &jbody)
	return jbody, err
}

func taskResponse(resp *http.Response) (Task, error) {
	var jbody map[string]interface{}
	var err error
	if err = decodeResponse(resp, &jbody); err != nil {
		return nil, err
	}
	if v, isSet := jbody["errors"]; isSet {
		errJSON, _ := json.MarshalIndent(v, "", "  ")
		return nil, fmt.Errorf("error: %s", errJSON)
	}
	if v, isSet := jbody["data"]; isSet {
		task := &task{}
		task.mapToSDK_Unsafe(v.(string))
		return task, nil
	}
	return nil, nil
}

// Is this needed?
func TypedResponse(resp *http.Response, v interface{}) error {
	var intermediate struct {
		Data struct {
			Result json.RawMessage `json:"result"`
		} `json:"data"`
	}
	err := decodeResponse(resp, &intermediate)
	if err != nil {
		return fmt.Errorf("error reading response envelope: %v", err)
	}
	if err = json.Unmarshal(intermediate.Data.Result, v); err != nil {
		return fmt.Errorf("error unmarshalling result %v", err)
	}
	return nil
}

func (s *Session) SetAPIToken(userID, token string) {
	auth := fmt.Sprintf("%s=%s", userID, token)
	s.AuthToken = auth
}

func (s *Session) setTicket(ticket, csrfPreventionToken string) {
	s.AuthTicket = ticket
	s.CsrfToken = csrfPreventionToken
}

func (s *Session) Login(ctx context.Context, username string, password string, otp string) (err error) {
	reqUser := map[string]interface{}{"username": username, "password": password}
	if otp != "" {
		reqUser["otp"] = otp
	}
	reqbody := ParamsToBody(reqUser)
	olddebug := *Debug
	*Debug = false // don't share passwords in debug log
	resp, err := s.Post(ctx, "/access/ticket", nil, &s.Headers, &reqbody)
	*Debug = olddebug
	if err != nil {
		return err
	}
	if resp == nil {
		return fmt.Errorf("Login error reading response")
	}
	dr, _ := httputil.DumpResponse(resp, true)
	jbody, err := ResponseJSON(resp)
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

func (s *Session) Do(req *http.Request) (*http.Response, error) {
	// Add session headers
	for k, v := range s.Headers {
		req.Header[k] = v
	}

	if *Debug {
		includeBody := req.ContentLength < DebugLargeBodyThreshold
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
	// session.Do, and they might not be able to reliably close it themselves.
	// Therefore, read the body out, close the original, then replace it with
	// a NopCloser over the bytes, which does not need to be closed downsteam.
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(respBody))

	if *Debug {
		includeBody := resp.ContentLength < DebugLargeBodyThreshold
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
func (s *Session) Request(
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

	// Get the body if one is present
	var buf io.Reader
	if body != nil {
		buf = bytes.NewReader(*body)
	}

	req, err := s.NewRequest(ctx, method, url, headers, buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	return s.Do(req)
}

// Perform a simple get to an endpoint and unmarshal returned JSON
func (s *Session) RequestJSON(
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

	resp, err = s.Request(ctx, method, url, params, headers, &bodyjson)
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

func (s *Session) Delete(
	ctx context.Context,
	url string,
	params *url.Values,
	headers *http.Header,
) (resp *http.Response, err error) {
	return s.Request(ctx, "DELETE", url, params, headers, nil)
}

func (s *Session) Get(
	ctx context.Context,
	url string,
	params *url.Values,
	headers *http.Header,
) (resp *http.Response, err error) {
	return s.Request(ctx, "GET", url, params, headers, nil)
}

func (s *Session) GetJSON(
	ctx context.Context,
	url string,
	params *url.Values,
	headers *http.Header,
	responseContainer interface{},
) (resp *http.Response, err error) {
	return s.RequestJSON(ctx, "GET", url, params, headers, nil, responseContainer)
}

func (s *Session) Head(
	ctx context.Context,
	url string,
	params *url.Values,
	headers *http.Header,
) (resp *http.Response, err error) {
	return s.Request(ctx, "HEAD", url, params, headers, nil)
}

func (s *Session) Post(
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
	return s.Request(ctx, "POST", url, params, headers, body)
}

func (s *Session) PostJSON(
	ctx context.Context,
	url string,
	params *url.Values,
	headers *http.Header,
	body interface{},
	responseContainer interface{},
) (resp *http.Response, err error) {
	return s.RequestJSON(ctx, "POST", url, params, headers, body, responseContainer)
}

func (s *Session) Put(
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
	return s.Request(ctx, "PUT", url, params, headers, body)
}

type Task interface {
	EndTime() time.Time
	ExitStatus() string
	ID() string
	Node() string
	OperationType() string
	ProcessID() uint
	StartTime() time.Time
	Status() string
	User() UserID
	WaitForCompletion(context.Context, *Client) error
}

type task struct {
	id            string
	node          string
	operationType string
	status        map[string]interface{}
	statusMutex   sync.Mutex
	user          UserID
}

const (
	taskApiKeyEndTime    = "endtime"
	taskApiKeyExitStatus = "exitstatus"
	taskApiKeyProcessID  = "pid"
	taskApiKeyStartTime  = "starttime"
	taskApiKeyStatus     = "status"
)

// Returns the time the task ended. If the task has not ended, the zero time is returned.
func (t *task) EndTime() time.Time {
	if t == nil {
		return time.Time{}
	}
	t.statusMutex.Lock()
	defer t.statusMutex.Unlock()
	if v, isSet := t.status[taskApiKeyEndTime]; isSet {
		return time.Unix(int64(v.(float64)), 0)
	}
	return time.Time{}
}

// Returns the exit status of the task. If the task has not ended, an empty string is returned.
func (t *task) ExitStatus() string {
	if t == nil {
		return ""
	}
	t.statusMutex.Lock()
	defer t.statusMutex.Unlock()
	if v, isSet := t.status[taskApiKeyExitStatus]; isSet {
		return v.(string)
	}
	return ""
}

// Returns the ID of the task.
func (t *task) ID() string {
	if t == nil {
		return ""
	}
	return t.id
}

// Returns the node the task was executed on.
func (t *task) Node() string {
	if t == nil {
		return ""
	}
	return t.node
}

// Returns the operation type of the task.
func (t *task) OperationType() string {
	if t == nil {
		return ""
	}
	return t.operationType
}

// Returns the process ID of the task.
func (t *task) ProcessID() uint {
	if t == nil {
		return 0
	}
	t.statusMutex.Lock()
	defer t.statusMutex.Unlock()
	if v, isSet := t.status[taskApiKeyProcessID]; isSet {
		return uint(v.(float64))
	}
	return 0
}

// Returns the time the task started.
func (t *task) StartTime() time.Time {
	if t == nil {
		return time.Time{}
	}
	t.statusMutex.Lock()
	defer t.statusMutex.Unlock()
	if v, isSet := t.status[taskApiKeyStartTime]; isSet {
		return time.Unix(int64(v.(float64)), 0)
	}
	return time.Time{}
}

// Returns the status of the task.
func (t *task) Status() string {
	if t == nil {
		return ""
	}
	t.statusMutex.Lock()
	defer t.statusMutex.Unlock()
	if v, isSet := t.status[taskApiKeyStatus]; isSet {
		return v.(string)
	}
	return ""
}

// Returns the user that started the task.
func (t *task) User() UserID {
	if t == nil {
		return UserID{}
	}
	return t.user
}

// Poll the API for task completion
func (t *task) WaitForCompletion(ctx context.Context, c *Client) error {
	if t == nil {
		return nil
	}
	var err error
	var waited int
	for waited < c.TaskTimeout {
		err = t.getTaskStatus_Unsafe(ctx, c.session)
		if err != nil && err != io.ErrUnexpectedEOF { // don't give up on ErrUnexpectedEOF
			return err
		}
		time.Sleep(TaskStatusCheckInterval * time.Second)
		if err = ctx.Err(); err != nil {
			return err
		}
		waited += TaskStatusCheckInterval
	}
	return fmt.Errorf("Wait timeout for:" + t.id)
}

// UPID:pve-test:002860A9:051E01C1:67536165:qmmove:102:root@pam:
// Requires the caller to ensure (t *task) is not nil.
func (t *task) mapToSDK_Unsafe(upID string) {
	t.id = upID
	indexA := strings.Index(upID[5:], ":") + 5
	t.node = upID[5:indexA]
	indexB := strings.Index(upID[indexA+28:], ":") + indexA + 28
	t.operationType = upID[indexA+28 : indexB]
	indexA = strings.Index(upID[indexB+1:], ":") + indexB + 1 + 1 // +1 because we are skipping a field
	t.user = UserID{}.mapToStruct(upID[indexA : strings.Index(upID[indexA:], ":")+indexA])
}

// Requires the caller to ensure (t *task) is not nil.
func (t *task) getTaskStatus_Unsafe(ctx context.Context, session *Session) (err error) {
	var data map[string]interface{}
	_, err = session.GetJSON(ctx, "/nodes/"+t.node+"/tasks/"+t.id+"/status", nil, nil, &data)
	if err != nil {
		return
	}
	status := data["data"].(map[string]interface{})
	t.statusMutex.Lock()
	t.status = status
	t.statusMutex.Unlock()
	if v, isSet := status[taskApiKeyExitStatus]; isSet {
		exitStatus := v.(string)
		if !(strings.HasPrefix(exitStatus, "OK") || strings.HasPrefix(exitStatus, "WARNINGS")) {
			return fmt.Errorf(exitStatus)
		}
	}
	return
}
