package globalpayments

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
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

// Global Payment Default test environment constants
const (
	DefaultBaseURL    = "https://test.realexpayments.com"
	DefaultMerchantID = "realexsandbox"
	DefaultHashSecret = "Po8lRRT67a"
	DefaultRebateHash = "Po8lRRT67a"
	DefaultPath       = "/epage-remote.cgi"
)

// Global Payment Error values

type ValidationError struct {
	Response *http.Response
}

func (err *ValidationError) Error() string {
	return fmt.Sprintf("Validation Hash Error: method: %v, path: %v, status code:%d", err.Response.Request.Method,
		err.Response.Request.URL.Path, err.Response.StatusCode)
}

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

	client.CardStorage = &CardStorageService{service: service{client: client, Path: DefaultPath}}

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

type serviceAuthenticator struct {
	elementsToHash []string
	sharedSecret   string
}

type Marshaller interface {
	io.Writer
	Sum(b []byte) []byte
}

type Authenticator interface {
	hashAndEncode(m Marshaller, str string) (hashAndEncodedString string, err error)
	buildSignature() (signature string, err error)
}

func (authenticator *serviceAuthenticator) buildSignature() (signature string, err error) {
	hashedElementsString, err := authenticator.hashAndEncode(sha1.New(), strings.Join(authenticator.elementsToHash, "."))
	if err != nil {
		return "", err
	}

	signature, err = authenticator.hashAndEncode(sha1.New(), hashedElementsString+"."+authenticator.sharedSecret)
	if err != nil {
		return "", err
	}

	return signature, nil
}

func (authenticator *serviceAuthenticator) hashAndEncode(m Marshaller, str string) (hashAndEncodedString string, err error) {

	_, err = io.WriteString(m, str)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(m.Sum(nil)), nil
}

type ServiceResponse struct {
	XMLName             xml.Name    `xml:"response"`
	Timestamp           string      `xml:"timestamp,attr"`
	MerchantID          string      `xml:"merchantid"`
	Account             string      `xml:"account"`
	OrderID             string      `xml:"orderid"`
	AuthCode            string      `xml:"authcode"`
	Result              string      `xml:"result"`
	CVNResult           string      `xml:"cvnresult"`
	AVSPostcodeResponse string      `xml:"avspostcoderesponse"`
	AVSAddressResponse  string      `xml:"avsaddressresponse"`
	BatchId             string      `xml:"batchid"`
	Message             string      `xml:"message"`
	PasRef              string      `xml:"pasref"`
	TimeTaken           string      `xml:"timetaken"`
	AuthTimeTaken       string      `xml:"authtimetaken"`
	CardIssuer          *CardIssuer `xml:"cardissuer"`
	Sha1Hash            string      `xml:"sha1hash"`
	serviceAuthenticator
}

type CardIssuer struct {
	Bank        string `xml:"bank"`
	Country     string `xml:"country"`
	CountryCode string `xml:"countrycode"`
	Region      string `xml:"region"`
}

type ResponseAuthenticator interface {
	Authenticator
	validateSignature(httpResponse *http.Response) (err error)
}

func (authenticator *ServiceResponse) validateResponseHash(httpResponse *http.Response) (err error) {
	signature, err := authenticator.buildSignature()
	if err != nil {
		return err
	}
	if signature == authenticator.Sha1Hash {
		return nil
	}
	return &ValidationError{httpResponse}
}

type Transmitter interface {
	transmitRequest(interface{}) (response *ServiceResponse, httpResponse *http.Response,
		err error)
}

func (transmitter *service) transmitRequest(request interface{}) (response *ServiceResponse, httpResponse *http.Response,
	err error) {

	response = &ServiceResponse{}
	if err != nil {
		return nil, nil, err
	}

	httpRequest, err := transmitter.client.NewRequest("POST", transmitter.Path, request)

	if err != nil {
		return nil, nil, err
	}

	httpResponse, err = transmitter.client.Do(httpRequest, response)
	if err != nil {
		return nil, httpResponse, err
	}

	response.elementsToHash = []string{response.Timestamp, response.MerchantID, response.OrderID, response.Result, response.Message, response.PasRef, response.AuthCode}
	response.sharedSecret = transmitter.client.HashSecret
	err = response.validateResponseHash(httpResponse)
	if err != nil {
		return nil, httpResponse, err
	}

	return response, httpResponse, nil
}
