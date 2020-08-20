package globalpayments

import (
	. "github.com/miguel-rivera/go-global/globalpayments"
	"testing"
)

//create auth request. validated struct contains all values. validate correct response
//incorrect hash,
//valid hash

func TestCardStorage_buildRequestHash_validHash(t *testing.T) {

}

func TestCardStorage_Authorize_validRequestBody(t *testing.T) {
	authRequest := &CardStorageRequest{
		Account:       "internet",
		Channel:       "ECOM",
		OrderID:       "TestID_123",
		Amount:        &Amount{Currency: "CAD", Amount: "1001"},
		AutoSettle:    &AutoSettle{Flag: "1"},
		PayerRef:      "03e28f0e-492e-80bd-20ec318e9334",
		PaymentMethod: "3c4af936-3732-a393-f558bec2fb2a",
		PaymentData:   &PaymentData{CVN: CVN{Number: "222"}},
	}

	client, err := NewClient()

	if err != nil {
		t.Errorf("Error initializing new client %v", err)
	}

	response, httpResponse, error := client.CardStorage.Authorize(authRequest)

	// mock time
	// confirm struct for new request contains hard coded fields
	// mock call to client.newRequest

	// mock call to client.do

	t.Errorf("Test %v *** %v *** %v ", response, httpResponse, error)
}
func TestCardStorage_Authorize_validResponseHash(t *testing.T) {
	authRequest := &CardStorageRequest{
		Account:       "internet",
		Channel:       "ECOM",
		OrderID:       "TestID_123",
		Amount:        &Amount{Currency: "CAD", Amount: "1001"},
		AutoSettle:    &AutoSettle{Flag: "1"},
		PayerRef:      "03e28f0e-492e-80bd-20ec318e9334",
		PaymentMethod: "3c4af936-3732-a393-f558bec2fb2a",
		PaymentData:   &PaymentData{CVN: CVN{Number: "222"}},
	}

	client, err := NewClient()

	if err != nil {
		t.Errorf("Error initializing new client %v", err)
	}

	response, httpResponse, error := client.CardStorage.Authorize(authRequest)

	t.Errorf("Test %v *** %v *** %v ", response, httpResponse, error)
}

func TestCardStorage_Authorize_invalidResponseHash(t *testing.T) {

}

func TestCardStorage_Authorize_clientError(t *testing.T) {

}
