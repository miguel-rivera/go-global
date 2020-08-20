package globalpayments

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	. "github.com/miguel-rivera/go-global/globalpayments"
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
