package globalpayments

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

type request struct {
	XMLName    xml.Name `xml:"request"`
	Type       string   `xml:"type,attr"`
	Timestamp  string   `xml:"timestamp,attr"`
	MerchantId string   `xml:"merchantid"`
	Account    string   `xml:"account"`
}

func TestClient_NewClient_DefaultValues(t *testing.T) {

	client, err := NewClient()

	if err != nil {
		t.Errorf("Error initializing new client %v", err)
	}

	if got, want := client.BaseURL.String(), DefaultBaseURL; got != want {
		t.Errorf("NewClient BaseURL is %v, want %v", got, want)
	}

	if got, want := client.HashSecret, DefaultHashSecret; got != want {
		t.Errorf("NewClient HashSecret is %v, want %v", got, want)
	}

	if got, want := client.MerchantID, DefaultMerchantID; got != want {
		t.Errorf("NewClient MerchantID is %v, want %v", got, want)
	}

	clientTwo, _ := NewClient()
	if client.HTTPClient == clientTwo.HTTPClient {
		t.Error("NewClient returned same http.Clients, but they should differ")
	}
}

func TestClient_NewClient_VariadicFunctionsConfiguration(t *testing.T) {

	baseUrl := func(client *Client) {
		url, _ := url.Parse("https://testing.go")
		client.BaseURL = url
	}

	hashSecret := func(client *Client) {
		client.HashSecret = "testing"
	}

	merchantId := func(client *Client) {
		client.MerchantID = "testMerchant"
	}

	httpClient := &http.Client{}

	setHttpClient := func(client *Client) {
		client.HTTPClient = httpClient
	}

	client, _ := NewClient(baseUrl, hashSecret, merchantId, setHttpClient)

	if got, want := client.BaseURL.String(), "https://testing.go"; got != want {
		t.Errorf("NewClient BaseURL is %v, want %v", got, want)
	}
	if got, want := client.HashSecret, "testing"; got != want {
		t.Errorf("NewClient HashSecret is %v, want %v", got, want)
	}

	if got, want := client.MerchantID, "testMerchant"; got != want {
		t.Errorf("NewClient MerchantID is %v, want %v", got, want)
	}

	if client.HTTPClient != httpClient {
		t.Error("NewClient returned different httpClient")
	}
}

func TestClient_NewRequest(t *testing.T) {
	client, _ := NewClient()
	method := "POST"
	inURL, outURL := "/test", DefaultBaseURL+"/test"

	requestBody := &request{Type: "auth", Timestamp: "20180613141207", MerchantId: "testMerchant",
		Account: "internet"}

	requestXMLBody := `<request type="auth" timestamp="20180613141207"><merchantid>testMerchant</merchantid><account>internet</account></request>`

	req, _ := client.NewRequest(method, inURL, requestBody)

	if got, want := req.Method, method; got != want {
		t.Errorf("Request Method is %v, want %v", got, want)
	}

	if got, want := req.URL.String(), outURL; got != want {
		t.Errorf("Request URL is %v, want %v", got, want)
	}

	body, _ := ioutil.ReadAll(req.Body)

	if got, want := string(body), requestXMLBody; got != want {
		t.Errorf("Request Body is %v, want %v", got, want)
	}
}

func TestClient_Do(t *testing.T) {

	requestBody := &request{XMLName: xml.Name{Local: "request"}, Type: "auth", Timestamp: "20180613141207",
		MerchantId: "testMerchant",
		Account:    "internet"}

	requestXMLBody := `<request type="auth" timestamp="20180613141207"><merchantid>testMerchant</merchantid><account>internet</account></request>`

	handler := func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, "POST"; got != want {
			t.Errorf("Request method: %v, want %v", got, want)
		}

		fmt.Fprint(w, requestXMLBody)
	}

	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/test", handler)
	server := httptest.NewServer(serverMux)
	defer server.Close()

	baseUrl := func(client *Client) {
		url, _ := url.Parse(server.URL)
		client.BaseURL = url
	}

	client, _ := NewClient(baseUrl)

	req, _ := client.NewRequest("POST", "/test", requestBody)
	body := &request{}

	client.Do(req, body)
	if !reflect.DeepEqual(body, requestBody) {
		t.Errorf("Response Body = %v, want %v", body, requestBody)
	}
}

type MockMarshaller struct {
	calls []string
	p     []byte
}

const write = "write"
const sum = "sum"
const message = "hello world"

func (mock *MockMarshaller) Write(p []byte) (n int, err error) {
	mock.calls = append(mock.calls, write)
	mock.p = p
	return 1, nil
}

func (mock *MockMarshaller) Sum(b []byte) []byte {
	mock.calls = append(mock.calls, sum)
	return []byte("Hello")
}

func Test_Authenticator_hashAndEncode(t *testing.T) {
	mock := &MockMarshaller{}
	authenticator := &serviceAuthenticator{}
	hashedValue, err := authenticator.hashAndEncode(mock, message)

	if err != nil {
		t.Errorf("Error calling hashAndEncode %v", err)
	}

	if got, want := hashedValue, "48656c6c6f"; got != want {
		t.Errorf("Hashed value returned is: %v, want %v", got, want)
	}

	if got, want := mock.p, []byte(message); !bytes.Equal(got, want) {
		t.Errorf("Hashed byte slice passed to marsheller returned: %s, want %s", got, want)
	}

	if got, want := mock.calls, []string{write, sum}; !reflect.DeepEqual(got, want) {
		t.Errorf("Order of operations is incorrect: %s, want %s", got, want)
	}
}

func Test_Authenticator_buildSignature(t *testing.T) {
	request := &CardStorageRequest{serviceAuthenticator: serviceAuthenticator{sharedSecret: "test", elementsToHash: []string{"elem1", "elem2"}}}

	request.Sha1Hash, _ = request.buildSignature()

	//sha1 hashed elements deliminator by "." and salted with shared secret
	if got, want := request.Sha1Hash, "dbd4aebd6ead0f3c2e56017aef55135c4efd3aba"; got != want {
		t.Errorf("Request Sha1Hash is: %v want: %v", got, want)
	}
}

func Test_ResponseAuthenticator_validateResponseHash_valid(t *testing.T) {
	response := &ServiceResponse{
		Timestamp:            "20200204155942",
		MerchantID:           "Merchant ID",
		OrderID:              "N6qsk4kYRZihmPrTXWYS6g",
		Result:               "00",
		Message:              "[ test system ] Authorised",
		PasRef:               "14631546336115597",
		AuthCode:             "12345",
		Sha1Hash:             "a4fd14b21b1e4061b94902dabff63287690c4f0c",
		serviceAuthenticator: serviceAuthenticator{sharedSecret: "Po8lRRT67a"}}

	response.elementsToHash = []string{response.Timestamp, response.MerchantID, response.OrderID, response.Result, response.Message, response.PasRef, response.AuthCode}
	err := response.validateResponseHash(&http.Response{})

	if err != nil {
		t.Errorf("Validation Error thrown on valid response %v", err)
	}
}

func Test_ResponseAuthenticator_validateResponseHash_invalid(t *testing.T) {
	response := &ServiceResponse{
		Timestamp:            "20200204155942",
		MerchantID:           "Merchant ID",
		OrderID:              "N6qsk4kYRZihmPrTXWYS6g",
		Result:               "00",
		Message:              "[ test system ] Authorised",
		PasRef:               "14631546336115597",
		AuthCode:             "12345",
		Sha1Hash:             "invalid",
		serviceAuthenticator: serviceAuthenticator{sharedSecret: "Po8lRRT67a"}}

	response.elementsToHash = []string{response.Timestamp, response.MerchantID, response.OrderID, response.Result, response.Message, response.PasRef, response.AuthCode}
	err := response.validateResponseHash(&http.Response{StatusCode: 200, Request: &http.Request{Method: "POST", URL: &url.URL{Path: "/test"}}})

	if got, want := err.Error(), "Validation Hash Error: method: POST, path: /test, status code:200"; got != want {
		t.Errorf("Incorrect Validation Error thrown got: %v, want: %v", got, want)
	}
}
