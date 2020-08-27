package globalpayments

import (
	"bytes"
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"
)

//create auth request. validated struct contains all values. validate correct response
//incorrect hash,
//valid hash
type MockTimeFormatter struct {
	counter int
	layout  string
}

func (mock *MockTimeFormatter) Format(layout string) string {
	mock.layout = layout
	mock.counter++
	return "20180614095601"
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

func Test_formatTimer(t *testing.T) {
	mock := &MockTimeFormatter{}

	returnedString := formatTime(mock, "20060102150405")

	if got, want := mock.counter, 1; got != want {
		t.Errorf("formatTime called  %v time(s), want %v time(s)", got, want)
	}

	if got, want := returnedString, "20180614095601"; got != want {
		t.Errorf("CardStorage formatTime is %v, want %v", got, want)
	}
}

func Test_hashAndEncode(t *testing.T) {
	mock := &MockMarshaller{}

	hashedValue, err := hashAndEncode(mock, message)

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

func TestCardStorageService_buildRequestHash(t *testing.T) {
	request := &CardStorageRequest{sharedSecret: "test", elementsToHash: []string{"elem1", "elem2"}}

	service := &CardStorageService{}

	service.buildRequestHash(request)

	//sha1 hashed elements deliminator by "." and salted with shared secret
	if got, want := request.Sha1Hash, "dbd4aebd6ead0f3c2e56017aef55135c4efd3aba"; got != want {
		t.Errorf("Request Sha1Hash is: %v want: %v", got, want)
	}
}

func TestCardStorageService_validateResponseHash_valid(t *testing.T) {
	response := &CardStorageResponse{
		Timestamp:  "20200204155942",
		MerchantID: "Merchant ID",
		OrderID:    "N6qsk4kYRZihmPrTXWYS6g",
		Result:     "00",
		Message:    "[ test system ] Authorised",
		PasRef:     "14631546336115597",
		AuthCode:   "12345",
		SHA1Hash:   "a4fd14b21b1e4061b94902dabff63287690c4f0c"}

	service := &CardStorageService{client: &Client{HashSecret: "Po8lRRT67a"}}

	err := service.validateResponseHash(&http.Response{}, response)

	if err != nil {
		t.Errorf("Validation Error thrown on valid response %v", err)
	}
}

func TestCardStorageService_validateResponseHash_invalid(t *testing.T) {
	response := &CardStorageResponse{
		Timestamp:  "20200204155942",
		MerchantID: "Merchant ID",
		OrderID:    "N6qsk4kYRZihmPrTXWYS6g",
		Result:     "00",
		Message:    "[ test system ] Authorised",
		PasRef:     "14631546336115597",
		AuthCode:   "12345",
		SHA1Hash:   "invalid"}

	service := &CardStorageService{client: &Client{HashSecret: "Po8lRRT67a"}}

	err := service.validateResponseHash(&http.Response{StatusCode: 200, Request: &http.Request{Method: "POST", URL: &url.URL{Path: "/test"}}}, response)

	if got, want := err.Error(), "Validation Hash Error: method: POST, path: /test, status code:200"; got != want {
		t.Errorf("Incorrect Validation Error thrown got: %v, want: %v", got, want)
	}
}

func TestCardStorageService_addCommonValuesToRequest(t *testing.T) {
	Now = func() time.Time { return time.Unix(1528969800, 0) }

	request := &CardStorageRequest{Account: "internet", OrderID: "test", elementsToHash: []string{"internet", "test"}, sharedSecret: "Po8lRRT67a"}

	service := &CardStorageService{client: &Client{MerchantID: "test"}}

	service.addCommonValuesToRequest(request)

	if got, want := request.Timestamp, "20180614055000"; got != want {
		t.Errorf("Request Timestamp is: %v want: %v", got, want)
	}

	if got, want := request.MerchantID, "test"; got != want {
		t.Errorf("Request Sha1Hash is: %v want: %v", got, want)
	}

	if got, want := request.Sha1Hash, "5a9888a4222a91e8032c34b6493e409dd17288d6"; got != want {
		t.Errorf("Request Sha1Hash is: %v want: %v", got, want)
	}
}
