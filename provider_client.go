package gophercloud

import (
	"bytes"
	_ "bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"io"
	"io/ioutil"
	"k8s.io/klog/v2"
	"net/http"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
)
type P map[string]interface{}

// DefaultUserAgent is the default User-Agent string set in the request header.
const (
	DefaultUserAgent         = "gophercloud/v1.6.0"
	DefaultMaxBackoffRetries = 60
)

// UserAgent represents a User-Agent header.
type UserAgent struct {
	// prepend is the slice of User-Agent strings to prepend to DefaultUserAgent.
	// All the strings to prepend are accumulated and prepended in the Join method.
	prepend []string
}

type RetryBackoffFunc func(context.Context, *ErrUnexpectedResponseCode, error, uint) error

// RetryFunc is a catch-all function for retrying failed API requests.
// If it returns nil, the request will be retried.  If it returns an error,
// the request method will exit with that error.  failCount is the number of
// times the request has failed (starting at 1).
type RetryFunc func(context context.Context, method, url string, options *RequestOpts, err error, failCount uint) error

// Prepend prepends a user-defined string to the default User-Agent string. Users
// may pass in one or more strings to prepend.
func (ua *UserAgent) Prepend(s ...string) {
	ua.prepend = append(s, ua.prepend...)
}

// Join concatenates all the user-defined User-Agend strings with the default
// Gophercloud User-Agent string.
func (ua *UserAgent) Join() string {
	uaSlice := append(ua.prepend, DefaultUserAgent)
	return strings.Join(uaSlice, " ")
}

// ProviderClient stores details that are required to interact with any
// services within a specific provider's API.
//
// Generally, you acquire a ProviderClient by calling the NewClient method in
// the appropriate provider's child package, providing whatever authentication
// credentials are required.
type ProviderClient struct {
	// IdentityBase is the base URL used for a particular provider's identity
	// service - it will be used when issuing authenticatation requests. It
	// should point to the root resource of the identity service, not a specific
	// identity version.
	IdentityBase string

	// IdentityEndpoint is the identity endpoint. This may be a specific version
	// of the identity service. If this is the case, this endpoint is used rather
	// than querying versions first.
	IdentityEndpoint string

	// TokenID is the ID of the most recently issued valid token.
	// NOTE: Aside from within a custom ReauthFunc, this field shouldn't be set by an application.
	// To safely read or write this value, call `Token` or `SetToken`, respectively
	TokenID string

	// EndpointLocator describes how this provider discovers the endpoints for
	// its constituent services.
	EndpointLocator EndpointLocator

	// HTTPClient allows users to interject arbitrary http, https, or other transit behaviors.
	HTTPClient http.Client

	// UserAgent represents the User-Agent header in the HTTP request.
	UserAgent UserAgent

	// ReauthFunc is the function used to re-authenticate the user if the request
	// fails with a 401 HTTP response code. This a needed because there may be multiple
	// authentication functions for different Identity service versions.
	ReauthFunc func() error

	// Throwaway determines whether if this client is a throw-away client. It's a copy of user's provider client
	// with the token and reauth func zeroed. Such client can be used to perform reauthorization.
	Throwaway bool

	// Context is the context passed to the HTTP request.
	Context context.Context

	// Retry backoff func is called when rate limited.
	RetryBackoffFunc RetryBackoffFunc

	// MaxBackoffRetries set the maximum number of backoffs. When not set, defaults to DefaultMaxBackoffRetries
	MaxBackoffRetries uint

	// A general failed request handler method - this is always called in the end if a request failed. Leave as nil
	// to abort when an error is encountered.
	RetryFunc RetryFunc

	// mut is a mutex for the client. It protects read and write access to client attributes such as getting
	// and setting the TokenID.
	mut *sync.RWMutex

	// reauthmut is a mutex for reauthentication it attempts to ensure that only one reauthentication
	// attempt happens at one time.
	reauthmut *reauthlock

	authResult AuthResult
}

// reauthlock represents a set of attributes used to help in the reauthentication process.
type reauthlock struct {
	sync.RWMutex
	ongoing *reauthFuture
}

// reauthFuture represents future result of the reauthentication process.
// while done channel is not closed, reauthentication is in progress.
// when done channel is closed, err contains the result of reauthentication.
type reauthFuture struct {
	done chan struct{}
	err  error
}

func newReauthFuture() *reauthFuture {
	return &reauthFuture{
		make(chan struct{}),
		nil,
	}
}

func (f *reauthFuture) Set(err error) {
	f.err = err
	close(f.done)
}

func (f *reauthFuture) Get() error {
	<-f.done
	return f.err
}

// AuthenticatedHeaders returns a map of HTTP headers that are common for all
// authenticated service requests. Blocks if Reauthenticate is in progress.
func (client *ProviderClient) AuthenticatedHeaders() (m map[string]string) {
	if client.IsThrowaway() {
		return
	}
	if client.reauthmut != nil {
		// If a Reauthenticate is in progress, wait for it to complete.
		client.reauthmut.Lock()
		ongoing := client.reauthmut.ongoing
		client.reauthmut.Unlock()
		if ongoing != nil {
			_ = ongoing.Get()
		}
	}
	t := client.Token()
	if t == "" {
		return
	}
	// 王玉东 header添加Authorization
	return map[string]string{"X-Auth-Token": t,"Authorization": fmt.Sprintf("Bearer %s", string(t))}
}

// UseTokenLock creates a mutex that is used to allow safe concurrent access to the auth token.
// If the application's ProviderClient is not used concurrently, this doesn't need to be called.
func (client *ProviderClient) UseTokenLock() {
	client.mut = new(sync.RWMutex)
	client.reauthmut = new(reauthlock)
}

// GetAuthResult returns the result from the request that was used to obtain a
// provider client's Keystone token.
//
// The result is nil when authentication has not yet taken place, when the token
// was set manually with SetToken(), or when a ReauthFunc was used that does not
// record the AuthResult.
func (client *ProviderClient) GetAuthResult() AuthResult {
	if client.mut != nil {
		client.mut.RLock()
		defer client.mut.RUnlock()
	}
	return client.authResult
}

// Token safely reads the value of the auth token from the ProviderClient. Applications should
// call this method to access the token instead of the TokenID field
func (client *ProviderClient) Token() string {
	if client.mut != nil {
		client.mut.RLock()
		defer client.mut.RUnlock()
	}
	return client.TokenID
}

// SetToken safely sets the value of the auth token in the ProviderClient. Applications may
// use this method in a custom ReauthFunc.
//
// WARNING: This function is deprecated. Use SetTokenAndAuthResult() instead.
func (client *ProviderClient) SetToken(t string) {
	if client.mut != nil {
		client.mut.Lock()
		defer client.mut.Unlock()
	}
	client.TokenID = t
	client.authResult = nil
}

// SetTokenAndAuthResult safely sets the value of the auth token in the
// ProviderClient and also records the AuthResult that was returned from the
// token creation request. Applications may call this in a custom ReauthFunc.
func (client *ProviderClient) SetTokenAndAuthResult(r AuthResult) error {
	tokenID := ""
	var err error
	//klog.V(5).Infof("## SetTokenAndAuthResult,r: %+v", r)
	klog.Infof("## SetTokenAndAuthResult,r: %+v", r)
	if r != nil {
		tokenID, err = r.ExtractTokenID()
		klog.Infof("auth result ID: %s", tokenID)
		if err != nil {
			return err
		}
	}

	if client.mut != nil {
		client.mut.Lock()
		defer client.mut.Unlock()
	}
	klog.V(5).Infof("auth result ID: %s", tokenID)
	client.TokenID = tokenID
	client.authResult = r
	return nil
}

// CopyTokenFrom safely copies the token from another ProviderClient into the
// this one.
func (client *ProviderClient) CopyTokenFrom(other *ProviderClient) {
	if client.mut != nil {
		client.mut.Lock()
		defer client.mut.Unlock()
	}
	if other.mut != nil && other.mut != client.mut {
		other.mut.RLock()
		defer other.mut.RUnlock()
	}
	client.TokenID = other.TokenID
	client.authResult = other.authResult
}

// IsThrowaway safely reads the value of the client Throwaway field.
func (client *ProviderClient) IsThrowaway() bool {
	if client.reauthmut != nil {
		client.reauthmut.RLock()
		defer client.reauthmut.RUnlock()
	}
	return client.Throwaway
}

// SetThrowaway safely sets the value of the client Throwaway field.
func (client *ProviderClient) SetThrowaway(v bool) {
	if client.reauthmut != nil {
		client.reauthmut.Lock()
		defer client.reauthmut.Unlock()
	}
	client.Throwaway = v
}

// Reauthenticate calls client.ReauthFunc in a thread-safe way. If this is
// called because of a 401 response, the caller may pass the previous token. In
// this case, the reauthentication can be skipped if another thread has already
// reauthenticated in the meantime. If no previous token is known, an empty
// string should be passed instead to force unconditional reauthentication.
func (client *ProviderClient) Reauthenticate(previousToken string) error {
	if client.ReauthFunc == nil {
		return nil
	}

	if client.reauthmut == nil {
		return client.ReauthFunc()
	}

	future := newReauthFuture()

	// Check if a Reauthenticate is in progress, or start one if not.
	client.reauthmut.Lock()
	ongoing := client.reauthmut.ongoing
	if ongoing == nil {
		client.reauthmut.ongoing = future
	}
	client.reauthmut.Unlock()

	// If Reauthenticate is running elsewhere, wait for its result.
	if ongoing != nil {
		return ongoing.Get()
	}

	// Perform the actual reauthentication.
	var err error
	if previousToken == "" || client.TokenID == previousToken {
		err = client.ReauthFunc()
	} else {
		err = nil
	}

	// Mark Reauthenticate as finished.
	client.reauthmut.Lock()
	client.reauthmut.ongoing.Set(err)
	client.reauthmut.ongoing = nil
	client.reauthmut.Unlock()

	return err
}

// RequestOpts customizes the behavior of the provider.Request() method.
type RequestOpts struct {
	// JSONBody, if provided, will be encoded as JSON and used as the body of the HTTP request. The
	// content type of the request will default to "application/json" unless overridden by MoreHeaders.
	// It's an error to specify both a JSONBody and a RawBody.
	JSONBody interface{}
	// RawBody contains an io.Reader that will be consumed by the request directly. No content-type
	// will be set unless one is provided explicitly by MoreHeaders.
	RawBody io.Reader
	// JSONResponse, if provided, will be populated with the contents of the response body parsed as
	// JSON.
	JSONResponse interface{}
	// OkCodes contains a list of numeric HTTP status codes that should be interpreted as success. If
	// the response has a different code, an error will be returned.
	OkCodes []int
	// MoreHeaders specifies additional HTTP headers to be provided on the request.
	// MoreHeaders will be overridden by OmitHeaders
	MoreHeaders map[string]string
	// OmitHeaders specifies the HTTP headers which should be omitted.
	// OmitHeaders will override MoreHeaders
	OmitHeaders []string
	// ErrorContext specifies the resource error type to return if an error is encountered.
	// This lets resources override default error messages based on the response status code.
	ErrorContext error
	// KeepResponseBody specifies whether to keep the HTTP response body. Usually used, when the HTTP
	// response body is considered for further use. Valid when JSONResponse is nil.
	KeepResponseBody bool
	// iam or others
	Type string
}

// requestState contains temporary state for a single ProviderClient.Request() call.
type requestState struct {
	// This flag indicates if we have reauthenticated during this request because of a 401 response.
	// It ensures that we don't reauthenticate multiple times for a single request. If we
	// reauthenticate, but keep getting 401 responses with the fresh token, reauthenticating some more
	// will just get us into an infinite loop.
	hasReauthenticated bool
	// Retry-After backoff counter, increments during each backoff call
	retries uint
}

var applicationJSON = "application/json"

// Request performs an HTTP request using the ProviderClient's current HTTPClient. An authentication
// header will automatically be provided.
func (client *ProviderClient) Request(method, url string, options *RequestOpts) (*http.Response, error) {
	if options.Type == "iamAuthentication" {
		return client.doRequestIam(method, url, options, &requestState{
			hasReauthenticated: false,
		})
	}
	return client.doRequest(method, url, options, &requestState{
		hasReauthenticated: false,
	})
}
// 将结构体转换为URL编码的表单数据
func StructToURLValues(data map[string]interface{}) (url.Values, error) {
	vs := url.Values{}
	if data == nil {
		return vs, nil
	}
	for k, v := range data{
		key := ToString(k)
		if IsMapArray(v) {
			vs.Set(key, JSONEncode(v))
		} else if IsArray(v) {
			a, _ := v.([]interface{})
			for i, iv := range a {
				if i == 0 {
					vs.Set(key, ToString(iv))
				} else {
					vs.Add(key, ToString(iv))
				}
			}
		} else {
			vs.Set(key, ToString(v))
		}
	}
	return vs, nil
}
func (client *ProviderClient) doRequestIam(method, url string, options *RequestOpts, state *requestState) (*http.Response, error) {
	var body io.Reader
	var contentType *string

	klog.Infof("doRequestIam-->method: %+v,url: %+v,options: %+v,state: %+v", method, url, options, state)

	// Derive the content body by either encoding an arbitrary object as JSON, or by taking a provided
	// io.ReadSeeker as-is. Default the content-type to application/json.
	if options.JSONBody != nil {
		if options.RawBody != nil {
			return nil, errors.New("please provide only one of JSONBody or RawBody to gophercloud.Request()")
		}

		// 将结构体转换为URL编码的表单数据
		rendered := options.JSONBody.(map[string]interface{})
		formData, err := StructToURLValues(rendered)
		if err != nil {
			fmt.Println("Error converting struct to URL values:", err)
			return nil, errors.New("cannot StructToURLValues")
		}

		// 将表单数据写入请求体
		klog.Infof("doRequest-->formData.Encode(): %+v", formData.Encode())
		body = strings.NewReader(formData.Encode())
		contentType = &applicationJSON
	}

	// Return an error, when "KeepResponseBody" is true and "JSONResponse" is not nil
	if options.KeepResponseBody && options.JSONResponse != nil {
		return nil, errors.New("cannot use KeepResponseBody when JSONResponse is not nil")
	}

	if options.RawBody != nil {
		//body = options.RawBody
	}
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if client.Context != nil {
		req = req.WithContext(client.Context)
	}

	// Populate the request headers.
	// Apply options.MoreHeaders and options.OmitHeaders, to give the caller the chance to
	// modify or omit any header.
	if contentType != nil {
		req.Header.Set("Content-Type", *contentType)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", applicationJSON)

	// Set the User-Agent header
	req.Header.Set("User-Agent", client.UserAgent.Join())

	if options.MoreHeaders != nil {
		for k, v := range options.MoreHeaders {
			req.Header.Set(k, v)
		}
	}

	for _, v := range options.OmitHeaders {
		req.Header.Del(v)
	}

	// get latest token from client
	for k, v := range client.AuthenticatedHeaders() {
		req.Header.Set(k, v)
	}

	prereqtok := req.Header.Get("X-Auth-Token")

	client.HTTPClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// Issue the request.
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		if client.RetryFunc != nil {
			var e error
			state.retries = state.retries + 1
			e = client.RetryFunc(client.Context, method, url, options, err, state.retries)
			if e != nil {
				return nil, e
			}

			return client.doRequestIam(method, url, options, state)
		}
		return nil, err
	}

	// Allow default OkCodes if none explicitly set
	okc := options.OkCodes
	if okc == nil {
		okc = defaultOkCodes(method)
	}

	// Validate the HTTP response status.
	var ok bool
	for _, code := range okc {
		if resp.StatusCode == code {
			ok = true
			break
		}
	}

	if !ok {
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		respErr := ErrUnexpectedResponseCode{
			URL:            url,
			Method:         method,
			Expected:       okc,
			Actual:         resp.StatusCode,
			Body:           body,
			ResponseHeader: resp.Header,
		}
		errType := options.ErrorContext
		switch resp.StatusCode {
		case http.StatusBadRequest:
			err = ErrDefault400{respErr}
			if error400er, ok := errType.(Err400er); ok {
				err = error400er.Error400(respErr)
			}
		case http.StatusUnauthorized:
			if client.ReauthFunc != nil && !state.hasReauthenticated {
				err = client.Reauthenticate(prereqtok)
				if err != nil {
					e := &ErrUnableToReauthenticate{}
					e.ErrOriginal = respErr
					e.ErrReauth = err
					return nil, e
				}
				if options.RawBody != nil {
					if seeker, ok := options.RawBody.(io.Seeker); ok {
						seeker.Seek(0, 0)
					}
				}
				state.hasReauthenticated = true
				resp, err = client.doRequestIam(method, url, options, state)
				if err != nil {
					switch err.(type) {
					case *ErrUnexpectedResponseCode:
						e := &ErrErrorAfterReauthentication{}
						e.ErrOriginal = err.(*ErrUnexpectedResponseCode)
						return nil, e
					default:
						e := &ErrErrorAfterReauthentication{}
						e.ErrOriginal = err
						return nil, e
					}
				}
				return resp, nil
			}
			err = ErrDefault401{respErr}
			if error401er, ok := errType.(Err401er); ok {
				err = error401er.Error401(respErr)
			}
		case http.StatusForbidden:
			err = ErrDefault403{respErr}
			if error403er, ok := errType.(Err403er); ok {
				err = error403er.Error403(respErr)
			}
		case http.StatusNotFound:
			err = ErrDefault404{respErr}
			if error404er, ok := errType.(Err404er); ok {
				err = error404er.Error404(respErr)
			}
		case http.StatusMethodNotAllowed:
			err = ErrDefault405{respErr}
			if error405er, ok := errType.(Err405er); ok {
				err = error405er.Error405(respErr)
			}
		case http.StatusRequestTimeout:
			err = ErrDefault408{respErr}
			if error408er, ok := errType.(Err408er); ok {
				err = error408er.Error408(respErr)
			}
		case http.StatusConflict:
			err = ErrDefault409{respErr}
			if error409er, ok := errType.(Err409er); ok {
				err = error409er.Error409(respErr)
			}
		case http.StatusTooManyRequests, 498:
			err = ErrDefault429{respErr}
			if error429er, ok := errType.(Err429er); ok {
				err = error429er.Error429(respErr)
			}

			maxTries := client.MaxBackoffRetries
			if maxTries == 0 {
				maxTries = DefaultMaxBackoffRetries
			}

			if f := client.RetryBackoffFunc; f != nil && state.retries < maxTries {
				var e error

				state.retries = state.retries + 1
				e = f(client.Context, &respErr, err, state.retries)

				if e != nil {
					return resp, e
				}

				return client.doRequestIam(method, url, options, state)
			}
		case http.StatusInternalServerError:
			err = ErrDefault500{respErr}
			if error500er, ok := errType.(Err500er); ok {
				err = error500er.Error500(respErr)
			}
		case http.StatusBadGateway:
			err = ErrDefault502{respErr}
			if error502er, ok := errType.(Err502er); ok {
				err = error502er.Error502(respErr)
			}
		case http.StatusServiceUnavailable:
			err = ErrDefault503{respErr}
			if error503er, ok := errType.(Err503er); ok {
				err = error503er.Error503(respErr)
			}
		case http.StatusGatewayTimeout:
			err = ErrDefault504{respErr}
			if error504er, ok := errType.(Err504er); ok {
				err = error504er.Error504(respErr)
			}
		}

		if err == nil {
			err = respErr
		}

		if err != nil && client.RetryFunc != nil {
			var e error
			state.retries = state.retries + 1
			e = client.RetryFunc(client.Context, method, url, options, err, state.retries)
			if e != nil {
				return resp, e
			}

			return client.doRequestIam(method, url, options, state)
		}

		return resp, err
	}

	// Parse the response body as JSON, if requested to do so.
	if options.JSONResponse != nil {
		defer resp.Body.Close()
		// Don't decode JSON when there is no content
		if resp.StatusCode == http.StatusNoContent {
			// read till EOF, otherwise the connection will be closed and cannot be reused
			_, err = io.Copy(ioutil.Discard, resp.Body)
			return resp, err
		}
		if err := json.NewDecoder(resp.Body).Decode(options.JSONResponse); err != nil {
			if client.RetryFunc != nil {
				var e error
				state.retries = state.retries + 1
				e = client.RetryFunc(client.Context, method, url, options, err, state.retries)
				if e != nil {
					return resp, e
				}

				return client.doRequestIam(method, url, options, state)
			}
			return nil, err
		}
	}

	// Close unused body to allow the HTTP connection to be reused
	if !options.KeepResponseBody && options.JSONResponse == nil {
		defer resp.Body.Close()
		// read till EOF, otherwise the connection will be closed and cannot be reused
		if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
			return nil, err
		}
	}

	return resp, nil
}


func (client *ProviderClient) doRequest(method, url string, options *RequestOpts, state *requestState) (*http.Response, error) {
	var body io.Reader
	var contentType *string

	// Derive the content body by either encoding an arbitrary object as JSON, or by taking a provided
	// io.ReadSeeker as-is. Default the content-type to application/json.
	if options.JSONBody != nil {
		if options.RawBody != nil {
			return nil, errors.New("please provide only one of JSONBody or RawBody to gophercloud.Request()")
		}

		rendered, err := json.Marshal(options.JSONBody)
		if err != nil {
			return nil, err
		}

		body = bytes.NewReader(rendered)
		contentType = &applicationJSON
	}

	// Return an error, when "KeepResponseBody" is true and "JSONResponse" is not nil
	if options.KeepResponseBody && options.JSONResponse != nil {
		return nil, errors.New("cannot use KeepResponseBody when JSONResponse is not nil")
	}

	if options.RawBody != nil {
		body = options.RawBody
	}

	// Construct the http.Request.
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if client.Context != nil {
		req = req.WithContext(client.Context)
	}

	// Populate the request headers.
	// Apply options.MoreHeaders and options.OmitHeaders, to give the caller the chance to
	// modify or omit any header.
	if contentType != nil {
		req.Header.Set("Content-Type", *contentType)
	}
	req.Header.Set("Accept", applicationJSON)

	// Set the User-Agent header
	req.Header.Set("User-Agent", client.UserAgent.Join())

	if options.MoreHeaders != nil {
		for k, v := range options.MoreHeaders {
			req.Header.Set(k, v)
		}
	}

	for _, v := range options.OmitHeaders {
		req.Header.Del(v)
	}

	// get latest token from client
	for k, v := range client.AuthenticatedHeaders() {
		req.Header.Set(k, v)
	}

	prereqtok := req.Header.Get("X-Auth-Token")

	// Issue the request.
	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		if client.RetryFunc != nil {
			var e error
			state.retries = state.retries + 1
			e = client.RetryFunc(client.Context, method, url, options, err, state.retries)
			if e != nil {
				return nil, e
			}

			return client.doRequest(method, url, options, state)
		}
		return nil, err
	}

	// Allow default OkCodes if none explicitly set
	okc := options.OkCodes
	if okc == nil {
		okc = defaultOkCodes(method)
	}

	// Validate the HTTP response status.
	var ok bool
	for _, code := range okc {
		if resp.StatusCode == code {
			ok = true
			break
		}
	}

	if !ok {
		body, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		respErr := ErrUnexpectedResponseCode{
			URL:            url,
			Method:         method,
			Expected:       okc,
			Actual:         resp.StatusCode,
			Body:           body,
			ResponseHeader: resp.Header,
		}

		errType := options.ErrorContext
		switch resp.StatusCode {
		case http.StatusBadRequest:
			err = ErrDefault400{respErr}
			if error400er, ok := errType.(Err400er); ok {
				err = error400er.Error400(respErr)
			}
		case http.StatusUnauthorized:
			if client.ReauthFunc != nil && !state.hasReauthenticated {
				err = client.Reauthenticate(prereqtok)
				if err != nil {
					e := &ErrUnableToReauthenticate{}
					e.ErrOriginal = respErr
					e.ErrReauth = err
					return nil, e
				}
				if options.RawBody != nil {
					if seeker, ok := options.RawBody.(io.Seeker); ok {
						seeker.Seek(0, 0)
					}
				}
				state.hasReauthenticated = true
				resp, err = client.doRequest(method, url, options, state)
				if err != nil {
					switch err.(type) {
					case *ErrUnexpectedResponseCode:
						e := &ErrErrorAfterReauthentication{}
						e.ErrOriginal = err.(*ErrUnexpectedResponseCode)
						return nil, e
					default:
						e := &ErrErrorAfterReauthentication{}
						e.ErrOriginal = err
						return nil, e
					}
				}
				return resp, nil
			}
			err = ErrDefault401{respErr}
			if error401er, ok := errType.(Err401er); ok {
				err = error401er.Error401(respErr)
			}
		case http.StatusForbidden:
			err = ErrDefault403{respErr}
			if error403er, ok := errType.(Err403er); ok {
				err = error403er.Error403(respErr)
			}
		case http.StatusNotFound:
			err = ErrDefault404{respErr}
			if error404er, ok := errType.(Err404er); ok {
				err = error404er.Error404(respErr)
			}
		case http.StatusMethodNotAllowed:
			err = ErrDefault405{respErr}
			if error405er, ok := errType.(Err405er); ok {
				err = error405er.Error405(respErr)
			}
		case http.StatusRequestTimeout:
			err = ErrDefault408{respErr}
			if error408er, ok := errType.(Err408er); ok {
				err = error408er.Error408(respErr)
			}
		case http.StatusConflict:
			err = ErrDefault409{respErr}
			if error409er, ok := errType.(Err409er); ok {
				err = error409er.Error409(respErr)
			}
		case http.StatusTooManyRequests, 498:
			err = ErrDefault429{respErr}
			if error429er, ok := errType.(Err429er); ok {
				err = error429er.Error429(respErr)
			}

			maxTries := client.MaxBackoffRetries
			if maxTries == 0 {
				maxTries = DefaultMaxBackoffRetries
			}

			if f := client.RetryBackoffFunc; f != nil && state.retries < maxTries {
				var e error

				state.retries = state.retries + 1
				e = f(client.Context, &respErr, err, state.retries)

				if e != nil {
					return resp, e
				}

				return client.doRequest(method, url, options, state)
			}
		case http.StatusInternalServerError:
			err = ErrDefault500{respErr}
			if error500er, ok := errType.(Err500er); ok {
				err = error500er.Error500(respErr)
			}
		case http.StatusBadGateway:
			err = ErrDefault502{respErr}
			if error502er, ok := errType.(Err502er); ok {
				err = error502er.Error502(respErr)
			}
		case http.StatusServiceUnavailable:
			err = ErrDefault503{respErr}
			if error503er, ok := errType.(Err503er); ok {
				err = error503er.Error503(respErr)
			}
		case http.StatusGatewayTimeout:
			err = ErrDefault504{respErr}
			if error504er, ok := errType.(Err504er); ok {
				err = error504er.Error504(respErr)
			}
		}

		if err == nil {
			err = respErr
		}

		if err != nil && client.RetryFunc != nil {
			var e error
			state.retries = state.retries + 1
			e = client.RetryFunc(client.Context, method, url, options, err, state.retries)
			if e != nil {
				return resp, e
			}

			return client.doRequest(method, url, options, state)
		}

		return resp, err
	}

	// Parse the response body as JSON, if requested to do so.
	if options.JSONResponse != nil {
		defer resp.Body.Close()
		// Don't decode JSON when there is no content
		if resp.StatusCode == http.StatusNoContent {
			// read till EOF, otherwise the connection will be closed and cannot be reused
			_, err = io.Copy(ioutil.Discard, resp.Body)
			return resp, err
		}
		if err := json.NewDecoder(resp.Body).Decode(options.JSONResponse); err != nil {
			if client.RetryFunc != nil {
				var e error
				state.retries = state.retries + 1
				e = client.RetryFunc(client.Context, method, url, options, err, state.retries)
				if e != nil {
					return resp, e
				}

				return client.doRequest(method, url, options, state)
			}
			return nil, err
		}
	}

	// Close unused body to allow the HTTP connection to be reused
	if !options.KeepResponseBody && options.JSONResponse == nil {
		defer resp.Body.Close()
		// read till EOF, otherwise the connection will be closed and cannot be reused
		if _, err := io.Copy(ioutil.Discard, resp.Body); err != nil {
			return nil, err
		}
	}

	return resp, nil
}


func defaultOkCodes(method string) []int {
	switch method {
	case "GET", "HEAD":
		return []int{200}
	case "POST":
		return []int{201, 202}
	case "PUT":
		return []int{201, 202}
	case "PATCH":
		return []int{200, 202, 204}
	case "DELETE":
		return []int{202, 204}
	}

	return []int{}
}

func ToString(v interface{}, def ...string) string {
	if v != nil {
		switch v.(type) {
		case bson.ObjectId:
			return v.(bson.ObjectId).Hex()
		case []byte:
			return string(v.([]byte))
		case *P, P:
			var p P
			switch v.(type) {
			case *P:
				if v.(*P) != nil {
					p = *v.(*P)
				}
			case P:
				p = v.(P)
			}
			var keys []string
			for k := range p {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			r := "P{"
			for _, k := range keys {
				r = JoinStr(r, k, ":", p[k], " ")
			}
			r = JoinStr(r, "}")
			return r
		case map[string]interface{}, []P, []interface{}:
			return JSONEncode(v)
		case int64:
			return strconv.FormatInt(v.(int64), 10)
		case []string:
			s := ""
			for _, j := range v.([]string) {
				s = JoinStr(s, ",", j)
			}
			if len(s) > 0 {
				s = s[1:]
			}
			return s
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	if len(def) > 0 {
		return def[0]
	} else {
		return ""
	}
}
func JoinStr(val ...interface{}) (r string) {
	for _, v := range val {
		r += ToString(v)
	}
    return
}
func JSONEncode(value interface{}) string {
	_bytes, err := json.Marshal(value)
	if err != nil {
        return ""
    }
    return string(_bytes)
}
func IsMapArray(v interface{}) bool {
	a, b := v.([]interface{})
	if b {
		for _, m := range a {
			switch m.(type) {
			case map[string]interface{}:
				return true
			default:
				return false
			}
		}
	}
	return false
}
func IsArray(v interface{}) bool {
	if IsEmpty(v) {
		return false
	}
	switch reflect.TypeOf(v).Kind() {
	case reflect.Array, reflect.Slice:
		return true
	default:
		return false
	}
}

func IsEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	switch v.(type) {
	case P:
		return len(v.(P)) == 0
	case []interface{}:
		return len(v.([]interface{})) == 0
	case []P:
		return len(v.([]P)) == 0
	case *[]P:
		return len(*v.(*[]P)) == 0
	}
	return ToString(v) == ""
}