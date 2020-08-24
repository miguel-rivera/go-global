package globalpayments

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client manages communication with Global Payments API
type Client struct {
	HTTPClient       *http.Client
	BaseURL          *url.URL
	HashSecret       string
	RebateHashSecret string
	MerchantID       string
	APIPath          string
	// Services used for communicating different actions of Global Payments API
	CardStorage *CardStorageService
}

type service struct {
	client *Client
	Path   string
}

type ClientAPI interface {
	NewClient(options ...func(*Client)) (*Client, error)
	NewRequest(method, urlStr string, body interface{}) (*http.Request, error)
	Do(req *http.Request, v interface{}) (*http.Response, error)
}

// Global Payment Default test environment constants
const (
	DefaultBaseURL    = "https://test.realexpayments.com"
	DefaultMerchantID = "realexsandbox"
	DefaultHashSecret = "Po8lRRT67a"
	DefaultRebateHash = "Po8lRRT67a"
	DefaultPath       = "/epage-remote.cgi"
)

// Global Payment Error values
type DependencyError struct {
	ResponseCode string
	Message      string
	Response     *http.Response
}

func (err *DependencyError) Error() string {
	return fmt.Sprintf("Dependency Error: %v, %v : %d response code: %v, message: %v ", err.Response.Request.Method,
		err.Response.Request.URL, err.Response.StatusCode, err.ResponseCode, err.Message)
}

type RequestFormatError struct {
	ResponseCode string
	Message      string
	Response     *http.Response
}

func (err *RequestFormatError) Error() string {
	return fmt.Sprintf("Request Format Error: %v, %v : %d response code: %v, message: %v ", err.Response.Request.Method,
		err.Response.Request.URL, err.Response.StatusCode, err.ResponseCode, err.Message)
}

type InvalidAccountError struct {
	ResponseCode string
	Message      string
	Response     *http.Response
}

func (err *InvalidAccountError) Error() string {
	return fmt.Sprintf("Invalid Account Error: %v, %v : %d response code: %v, message: %v ",
		err.Response.Request.Method,
		err.Response.Request.URL, err.Response.StatusCode, err.ResponseCode, err.Message)
}

// NewClient returns a Global Payments API Client. If no functional options are provided, Default values will be used to initiate the client.
// Note: Default Values initiate requests to Global payments test environment.
// The services of a client divide the API into logical chunks and correspond to the structure of the Global Payments documentation at https://developer.globalpay.com/api/getting-started.
func NewClient(options ...func(*Client)) (*Client, error) {

	httpClient := &http.Client{}

	baseURL, err := url.Parse(DefaultBaseURL)

	if err != nil {
		return nil, err
	}

	client := &Client{HTTPClient: httpClient, BaseURL: baseURL, HashSecret: DefaultHashSecret,
		MerchantID: DefaultMerchantID, RebateHashSecret: DefaultRebateHash}

	client.CardStorage = &CardStorageService{client: client, Path: DefaultPath}

	for _, option := range options {
		option(client)
	}

	if strings.HasSuffix(client.BaseURL.Path, "/") {
		return nil, fmt.Errorf("baseURL %q contains a trailing slash", client.BaseURL)
	}

	return client, nil
}

// NewRequest creates API Request. Relative URLs should be specified with preceding slash.
// If specified, the value pointed to by body is XML encoded and included within the request body.
func (client *Client) NewRequest(method, urlStr string, body interface{}) (*http.Request, error) {

	rel, err := client.BaseURL.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	var buffer io.ReadWriter
	if body != nil {
		buffer = &bytes.Buffer{}
		encoder := xml.NewEncoder(buffer)
		err := encoder.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, rel.String(), buffer)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/xml")
	return req, nil
}

// Do sends an API request and returns an API Response.
// The API response is XML decoded and stored in value pointed to by v.
func (client *Client) Do(req *http.Request, v interface{}) (*http.Response, error) {

	resp, err := client.HTTPClient.Do(req)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	decoder := xml.NewDecoder(resp.Body)
	err = decoder.Decode(v)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
