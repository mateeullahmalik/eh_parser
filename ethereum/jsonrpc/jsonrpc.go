package jsonrpc

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"time"

	"encoding/json"
)

const (
	jsonrpcVersion = "2.0"
	timeout        = 30 * time.Second
)

// RPCClient sends JSON-RPC requests over HTTP to the provided JSON-RPC backend.
type RPCClient interface {
	CallWithContext(ctx context.Context, method string, params ...interface{}) (*RPCResponse, error)
	CallForWithContext(ctx context.Context, out interface{}, method string, params ...interface{}) error
}

type RPCRequest struct {
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
	ID      int         `json:"id"`
	JSONRPC string      `json:"jsonrpc"`
}

// NewRequest returns a new RPCRequest that can be created using the same convenient parameter syntax as Call()
func NewRequest(method string, params ...interface{}) *RPCRequest {
	request := &RPCRequest{
		Method:  method,
		Params:  Params(params...),
		ID:      rand.Int(),
		JSONRPC: jsonrpcVersion,
	}

	return request
}

type RPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
	ID      int         `json:"id"`
}

// RPCError represents a JSON-RPC error object if an RPC error occurred.
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Error function is provided to be used as error object.
func (e *RPCError) Error() string {
	return strconv.Itoa(e.Code) + ":" + e.Message
}

// HTTPError represents a error that occurred on HTTP level.
type HTTPError struct {
	Code int
	err  error
}

// Error function is provided to be used as error object.
func (e *HTTPError) Error() string {
	return e.err.Error()
}

type rpcClient struct {
	endpoint      string
	httpClient    *http.Client
	customHeaders map[string]string
}

// RPCClientOpts can be provided to NewClientWithOpts() to change configuration of RPCClient.
type RPCClientOpts struct {
	HTTPClient    *http.Client
	CustomHeaders map[string]string
}

// RPCResponses is of type []*RPCResponse.
type RPCResponses []*RPCResponse

// RPCRequests is of type []*RPCRequest.
// This type is used to provide helper functions on the request list
type RPCRequests []*RPCRequest

// NewClient returns a new RPCClient instance with default configuration.
func NewClient(endpoint string) RPCClient {
	return NewClientWithOpts(endpoint, nil)
}

// NewClientWithOpts returns a new RPCClient instance with custom configuration.
func NewClientWithOpts(endpoint string, opts *RPCClientOpts) RPCClient {
	rpcClient := &rpcClient{
		endpoint:      endpoint,
		httpClient:    &http.Client{},
		customHeaders: make(map[string]string),
	}

	if opts == nil {
		return rpcClient
	}

	if opts.HTTPClient != nil {
		rpcClient.httpClient = opts.HTTPClient
	}

	if opts.CustomHeaders != nil {
		for k, v := range opts.CustomHeaders {
			rpcClient.customHeaders[k] = v
		}
	}

	return rpcClient
}

func (client *rpcClient) CallWithContext(ctx context.Context, method string, params ...interface{}) (*RPCResponse, error) {
	request := &RPCRequest{
		Method:  method,
		Params:  Params(params...),
		JSONRPC: jsonrpcVersion,
	}

	return client.doCall(ctx, request)
}

func (client *rpcClient) CallForWithContext(ctx context.Context, out interface{}, method string, params ...interface{}) error {
	rpcResponse, err := client.CallWithContext(ctx, method, params...)

	if err != nil {
		return err
	}

	if rpcResponse.Error != nil {
		return fmt.Errorf("code: %d, message: %s", rpcResponse.Error.Code, rpcResponse.Error.Message)
	}

	err = rpcResponse.GetObject(out)
	// log.WithContext(ctx).Debugf("-----CONVERTED RESPONSE FOR CALL: %+v\n", out)
	return err
}

func (client *rpcClient) newRequest(ctx context.Context, req interface{}) (*http.Request, error) {

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequestWithContext(ctx, "POST", client.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	// set default headers first, so that even content type and accept can be overwritten
	for k, v := range client.customHeaders {
		request.Header.Set(k, v)
	}

	return request, nil
}

func (client *rpcClient) doCall(cctx context.Context, RPCRequest *RPCRequest) (*RPCResponse, error) {
	ctx, cancel := context.WithTimeout(cctx, timeout)
	defer cancel()

	httpRequest, err := client.newRequest(ctx, RPCRequest)
	if err != nil {
		return nil, fmt.Errorf("rpc call %v() on %v: %v", RPCRequest.Method, client.endpoint, err.Error())
	}
	httpRequest.Close = true
	httpResponse, err := client.httpClient.Do(httpRequest)
	if err != nil {
		return nil, fmt.Errorf("rpc call %v() on %v: %v", RPCRequest.Method, httpRequest.URL.String(), err.Error())
	}
	defer httpResponse.Body.Close()

	var rpcResponse *RPCResponse
	decoder := json.NewDecoder(httpResponse.Body)
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	err = decoder.Decode(&rpcResponse)

	// parsing error
	if err != nil {
		// if we have some http error, return it
		if httpResponse.StatusCode >= 400 {
			return nil, &HTTPError{
				Code: httpResponse.StatusCode,
				err:  fmt.Errorf("rpc call %v() on %v status code: %v. could not decode body to rpc response: %v", RPCRequest.Method, httpRequest.URL.String(), httpResponse.StatusCode, err.Error()),
			}
		}
		return nil, fmt.Errorf("rpc call %v() on %v status code: %v. could not decode body to rpc response: %v", RPCRequest.Method, httpRequest.URL.String(), httpResponse.StatusCode, err.Error())
	}

	// response body empty
	if rpcResponse == nil {
		// if we have some http error, return it
		if httpResponse.StatusCode >= 400 {
			return nil, &HTTPError{
				Code: httpResponse.StatusCode,
				err:  fmt.Errorf("rpc call %v() on %v status code: %v. rpc response missing", RPCRequest.Method, httpRequest.URL.String(), httpResponse.StatusCode),
			}
		}
		return nil, fmt.Errorf("rpc call %v() on %v status code: %v. rpc response missing", RPCRequest.Method, httpRequest.URL.String(), httpResponse.StatusCode)
	}

	return rpcResponse, nil
}

// Params is a helper function that uses the same parameter syntax as Call().
func Params(params ...interface{}) interface{} {
	var finalParams interface{}

	// if params was nil skip this and p stays nil
	if params != nil {
		switch len(params) {
		case 0: // no parameters were provided, do nothing so finalParam is nil and will be omitted
		case 1: // one param was provided, use it directly as is, or wrap primitive types in array
			if params[0] != nil {
				var typeOf reflect.Type

				// traverse until nil or not a pointer type
				for typeOf = reflect.TypeOf(params[0]); typeOf != nil && typeOf.Kind() == reflect.Ptr; typeOf = typeOf.Elem() {
					continue
				}

				if typeOf != nil {
					// now check if we can directly marshal the type or if it must be wrapped in an array
					switch typeOf.Kind() {
					// for these types we just do nothing, since value of p is already unwrapped from the array params
					case reflect.Struct:
						finalParams = params[0]
					case reflect.Array:
						finalParams = params[0]
					case reflect.Slice:
						finalParams = params[0]
					case reflect.Interface:
						finalParams = params[0]
					case reflect.Map:
						finalParams = params[0]
					default: // everything else must stay in an array (int, string, etc)
						finalParams = params
					}
				}
			} else {
				finalParams = params
			}
		default: // if more than one parameter was provided it should be treated as an array
			finalParams = params
		}
	}

	return finalParams
}

// GetInt converts the rpc response to an int64 and returns it.
//
// If result was not an integer an error is returned.
func (RPCResponse *RPCResponse) GetInt() (int64, error) {
	val, ok := RPCResponse.Result.(json.Number)
	if !ok {
		return 0, fmt.Errorf("could not parse int64 from %s", RPCResponse.Result)
	}

	i, err := val.Int64()
	if err != nil {
		return 0, err
	}

	return i, nil
}

// GetObject converts the rpc response to an arbitrary type.
func (RPCResponse *RPCResponse) GetObject(toType interface{}) error {
	js, err := json.Marshal(RPCResponse.Result)
	if err != nil {
		return err
	}

	err = json.Unmarshal(js, toType)
	if err != nil {
		return err
	}

	return nil
}
