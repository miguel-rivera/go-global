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
	"time"
)

type MockTimeFormatter struct {
	counter int
	layout  string
}

func (mock *MockTimeFormatter) Format(layout string) string {
	mock.layout = layout
	mock.counter++
	return "20180614095601"
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

const ( //url path for global payments api
	apiURLPath = "/epage-remote.cgi"
)

//Setup new client with mux to be used to add additional handlers for test.
func setup() (client *Client, mux *http.ServeMux, serverURL string, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	client, _ = NewClient()
	url, _ := url.Parse(server.URL)
	client.BaseURL = url

	Now = func() time.Time { return time.Unix(1528969800, 0) }

	return client, mux, server.URL, server.Close
}

func TestCardStorageService_Authorize(t *testing.T) {

	authRequest := &CardStorageRequest{
		Account:       "internet",
		Channel:       "ECOM",
		OrderID:       "AiCibJ5UR7utURy_slxhJw",
		PayerRef:      "03e28f0e-492e-80bd-20ec318e9334",
		PaymentMethod: "3c4af936-483e-a393-f558bec2fb2a",
		Amount: &Amount{
			Amount:   "10000",
			Currency: "CAD",
		},
		AutoSettle:  &AutoSettle{Flag: "1"},
		PaymentData: &PaymentData{CVN: CVN{Number: "123"}},
	}

	client, mux, _, teardown := setup()
	defer teardown()
	mux.HandleFunc(apiURLPath, func(w http.ResponseWriter, r *http.Request) {
		requestXMLBody := `<request type="receipt-in" timestamp="20180614095000"><merchantid>realexsandbox</merchantid><account>internet</account><channel>ECOM</channel><orderid>AiCibJ5UR7utURy_slxhJw</orderid><payerref>03e28f0e-492e-80bd-20ec318e9334</payerref><paymentmethod>3c4af936-483e-a393-f558bec2fb2a</paymentmethod><sha1hash>59a88d763f26bdcbbf4dd65d3b0aec0b1dd5f6f6</sha1hash><amount currency="CAD">10000</amount><autosettle flag="1"></autosettle><paymentdata><cvn><number>123</number></cvn></paymentdata></request>`
		responseXMLBody := `<response timestamp="20180731090859">
							   <merchantid>MerchantId</merchantid>
							   <account>internet</account>
							   <orderid>N6qsk4kYRZihmPrTXWYS6g</orderid>
							   <authcode>12345</authcode>
							   <result>00</result>
							   <cvnresult>M</cvnresult>
							   <avspostcoderesponse>M</avspostcoderesponse>
							   <avsaddressresponse>M</avsaddressresponse>
							   <batchid>319623</batchid>
							   <message>[ test system ] AUTHORISED</message>
							   <pasref>14610544313177922</pasref>
							   <timetaken>1</timetaken>
							   <authtimetaken>0</authtimetaken>
							   <srd>MMC0F00YE4000000715</srd>
							   <cardissuer>
								  <bank>AIB BANK</bank>
								  <country>IRELAND</country>
								  <countrycode>IE</countrycode>
								  <region>EUR</region>
							   </cardissuer>
							   <sha1hash>77ac77956e57156f47142a5723835badf767e272</sha1hash>
							</response>`
		if got, want := r.Method, "POST"; got != want {
			t.Errorf("Request method: %v, want %v", got, want)
		}
		body, _ := ioutil.ReadAll(r.Body)
		if !reflect.DeepEqual(string(body), requestXMLBody) {
			t.Errorf("Request Body = %v, want %v", string(body), requestXMLBody)
		}
		fmt.Fprint(w, responseXMLBody)
	})

	response, _, err := client.CardStorage.Authorize(authRequest)
	if err != nil {
		t.Errorf("Error performing Client.Do: %v", err)
	}

	expectedResponse := &ServiceResponse{
		XMLName:             xml.Name{Local: "response"},
		Timestamp:           "20180731090859",
		MerchantID:          "MerchantId",
		Account:             "internet",
		OrderID:             "N6qsk4kYRZihmPrTXWYS6g",
		Result:              "00",
		Message:             "[ test system ] AUTHORISED",
		PasRef:              "14610544313177922",
		AuthCode:            "12345",
		CVNResult:           "M",
		AVSPostcodeResponse: "M",
		AVSAddressResponse:  "M",
		BatchID:             "319623",
		TimeTaken:           "1",
		AuthTimeTaken:       "0",
		CardIssuer: &CardIssuer{
			Bank:        "AIB BANK",
			Country:     "IRELAND",
			CountryCode: "IE",
			Region:      "EUR",
		},
		Sha1Hash:             "77ac77956e57156f47142a5723835badf767e272",
		serviceAuthenticator: serviceAuthenticator{elementsToHash: []string{"20180731090859", "MerchantId", "N6qsk4kYRZihmPrTXWYS6g", "00", "[ test system ] AUTHORISED", "14610544313177922", "12345"}, sharedSecret: "Po8lRRT67a"}}
	if !reflect.DeepEqual(response, expectedResponse) {
		t.Errorf("Request Body = %v, want %v", response, expectedResponse)
	}
}

func TestCardStorageService_Validate(t *testing.T) {
	validateRequest := &CardStorageRequest{
		Account:       "internet",
		OrderID:       "AiCibJ5UR7utURy_slxhJw",
		PayerRef:      "03e28f0e-492e-80bd-20ec318e9334",
		PaymentMethod: "3c4af936-483e-a393-f558bec2fb2a",
		PaymentData:   &PaymentData{CVN: CVN{Number: "123"}},
	}

	client, mux, _, teardown := setup()
	defer teardown()
	mux.HandleFunc(apiURLPath, func(w http.ResponseWriter, r *http.Request) {
		requestXMLBody := `<request type="receipt-in-otb" timestamp="20180614095000"><merchantid>realexsandbox</merchantid><account>internet</account><orderid>AiCibJ5UR7utURy_slxhJw</orderid><payerref>03e28f0e-492e-80bd-20ec318e9334</payerref><paymentmethod>3c4af936-483e-a393-f558bec2fb2a</paymentmethod><sha1hash>0fc774cd46731deb27883ed3a019fe348bb8c206</sha1hash><paymentdata><cvn><number>123</number></cvn></paymentdata></request>`
		responseXMLBody := `<response timestamp="20180731090859">
							   <merchantid>MerchantId</merchantid>
							   <account>internet</account>
							   <orderid>N6qsk4kYRZihmPrTXWYS6g</orderid>
							   <authcode>12345</authcode>
							   <result>00</result>
							   <cvnresult>M</cvnresult>
							   <avspostcoderesponse>M</avspostcoderesponse>
							   <avsaddressresponse>M</avsaddressresponse>
							   <batchid>319623</batchid>
							   <message>[ test system ] AUTHORISED</message>
							   <pasref>14610544313177922</pasref>
							   <timetaken>1</timetaken>
							   <authtimetaken>0</authtimetaken>
							   <srd>MMC0F00YE4000000715</srd>
							   <cardissuer>
								  <bank>AIB BANK</bank>
								  <country>IRELAND</country>
								  <countrycode>IE</countrycode>
								  <region>EUR</region>
							   </cardissuer>
							   <sha1hash>77ac77956e57156f47142a5723835badf767e272</sha1hash>
							</response>`
		if got, want := r.Method, "POST"; got != want {
			t.Errorf("Request method: %v, want %v", got, want)
		}
		body, _ := ioutil.ReadAll(r.Body)
		if !reflect.DeepEqual(string(body), requestXMLBody) {
			t.Errorf("Request Body = %v, want %v", string(body), requestXMLBody)
		}
		fmt.Fprint(w, responseXMLBody)
	})

	response, _, err := client.CardStorage.Validate(validateRequest)
	if err != nil {
		t.Errorf("Error performing Client.Do: %v", err)
	}

	expectedResponse := &ServiceResponse{
		XMLName:             xml.Name{Local: "response"},
		Timestamp:           "20180731090859",
		MerchantID:          "MerchantId",
		Account:             "internet",
		OrderID:             "N6qsk4kYRZihmPrTXWYS6g",
		Result:              "00",
		Message:             "[ test system ] AUTHORISED",
		PasRef:              "14610544313177922",
		AuthCode:            "12345",
		CVNResult:           "M",
		AVSPostcodeResponse: "M",
		AVSAddressResponse:  "M",
		BatchID:             "319623",
		TimeTaken:           "1",
		AuthTimeTaken:       "0",
		CardIssuer: &CardIssuer{
			Bank:        "AIB BANK",
			Country:     "IRELAND",
			CountryCode: "IE",
			Region:      "EUR",
		},
		Sha1Hash:             "77ac77956e57156f47142a5723835badf767e272",
		serviceAuthenticator: serviceAuthenticator{elementsToHash: []string{"20180731090859", "MerchantId", "N6qsk4kYRZihmPrTXWYS6g", "00", "[ test system ] AUTHORISED", "14610544313177922", "12345"}, sharedSecret: "Po8lRRT67a"}}
	if !reflect.DeepEqual(response, expectedResponse) {
		t.Errorf("Request Body = %v, want %v", response, expectedResponse)
	}
}

func TestCardStorageService_Credit(t *testing.T) {
	creditRequest := &CardStorageRequest{
		Account:       "internet",
		Channel:       "ECOM",
		OrderID:       "AiCibJ5UR7utURy_slxhJw",
		PayerRef:      "03e28f0e-492e-80bd-20ec318e9334",
		PaymentMethod: "3c4af936-483e-a393-f558bec2fb2a",
		Amount: &Amount{
			Amount:   "10000",
			Currency: "CAD",
		},
	}

	client, mux, _, teardown := setup()
	defer teardown()
	mux.HandleFunc(apiURLPath, func(w http.ResponseWriter, r *http.Request) {
		requestXMLBody := `<request type="payment-out" timestamp="20180614095000"><merchantid>realexsandbox</merchantid><account>internet</account><channel>ECOM</channel><orderid>AiCibJ5UR7utURy_slxhJw</orderid><payerref>03e28f0e-492e-80bd-20ec318e9334</payerref><paymentmethod>3c4af936-483e-a393-f558bec2fb2a</paymentmethod><sha1hash>59a88d763f26bdcbbf4dd65d3b0aec0b1dd5f6f6</sha1hash><amount currency="CAD">10000</amount></request>`
		responseXMLBody := `<response timestamp="20180731090859">
							   <merchantid>MerchantId</merchantid>
							   <account>internet</account>
							   <orderid>N6qsk4kYRZihmPrTXWYS6g</orderid>
							   <authcode>12345</authcode>
							   <result>00</result>
							   <cvnresult>M</cvnresult>
							   <avspostcoderesponse>M</avspostcoderesponse>
							   <avsaddressresponse>M</avsaddressresponse>
							   <batchid>319623</batchid>
							   <message>[ test system ] AUTHORISED</message>
							   <pasref>14610544313177922</pasref>
							   <timetaken>1</timetaken>
							   <authtimetaken>0</authtimetaken>
							   <srd>MMC0F00YE4000000715</srd>
							   <cardissuer>
								  <bank>AIB BANK</bank>
								  <country>IRELAND</country>
								  <countrycode>IE</countrycode>
								  <region>EUR</region>
							   </cardissuer>
							   <sha1hash>77ac77956e57156f47142a5723835badf767e272</sha1hash>
							</response>`
		if got, want := r.Method, "POST"; got != want {
			t.Errorf("Request method: %v, want %v", got, want)
		}
		body, _ := ioutil.ReadAll(r.Body)
		if !reflect.DeepEqual(string(body), requestXMLBody) {
			t.Errorf("Request Body = %v, want %v", string(body), requestXMLBody)
		}
		fmt.Fprint(w, responseXMLBody)
	})

	response, _, err := client.CardStorage.Credit(creditRequest)
	if err != nil {
		t.Errorf("Error performing Client.Do: %v", err)
	}

	expectedResponse := &ServiceResponse{
		XMLName:             xml.Name{Local: "response"},
		Timestamp:           "20180731090859",
		MerchantID:          "MerchantId",
		Account:             "internet",
		OrderID:             "N6qsk4kYRZihmPrTXWYS6g",
		Result:              "00",
		Message:             "[ test system ] AUTHORISED",
		PasRef:              "14610544313177922",
		AuthCode:            "12345",
		CVNResult:           "M",
		AVSPostcodeResponse: "M",
		AVSAddressResponse:  "M",
		BatchID:             "319623",
		TimeTaken:           "1",
		AuthTimeTaken:       "0",
		CardIssuer: &CardIssuer{
			Bank:        "AIB BANK",
			Country:     "IRELAND",
			CountryCode: "IE",
			Region:      "EUR",
		},
		Sha1Hash:             "77ac77956e57156f47142a5723835badf767e272",
		serviceAuthenticator: serviceAuthenticator{elementsToHash: []string{"20180731090859", "MerchantId", "N6qsk4kYRZihmPrTXWYS6g", "00", "[ test system ] AUTHORISED", "14610544313177922", "12345"}, sharedSecret: "Po8lRRT67a"}}
	if !reflect.DeepEqual(response, expectedResponse) {
		t.Errorf("Request Body = %v, want %v", response, expectedResponse)
	}
}

func TestCardStorageService_CreateCustomer(t *testing.T) {

	creditRequest := &CardStorageRequest{
		Account: "internet",
		Channel: "ECOM",
		OrderID: "AiCibJ5UR7utURy_slxhJw",
		Payer: &Payer{
			Ref:               "03e28f0e-492e-80bd-20ec318e9334",
			PayerType:         "Retail",
			Title:             "Mr.",
			FirstName:         "James",
			Surname:           "Mason",
			Company:           "Global Payments",
			Email:             "text@mail.com",
			DateOfBirth:       "19851222",
			State:             "yorkshire",
			PassPhrase:        "montgomery",
			VatNumber:         "gb 1234",
			VariableReference: "Car Part",
			CustomerNumber:    "E8953893489",
			Address: &Address{
				Line1:    "Flat 123",
				Line2:    "House 123",
				Line3:    "Cul-de-sac",
				City:     "Halifax",
				County:   "West Yorkshire",
				PostCode: "W6 9HR",
				Country:  &Country{Code: "GB"},
			},
			PhoneNumbers: &PhoneNumbers{
				Home:   "+35312345678",
				Work:   "+3531987654321",
				Fax:    "+3531987654321",
				Mobile: "+3531987654321",
			},
		},
		PaymentMethod: "3c4af936-483e-a393-f558bec2fb2a",
		Amount: &Amount{
			Amount:   "10000",
			Currency: "CAD",
		},
		AutoSettle:  &AutoSettle{Flag: "1"},
		PaymentData: &PaymentData{CVN: CVN{Number: "123"}},
	}

	client, mux, _, teardown := setup()
	defer teardown()
	mux.HandleFunc(apiURLPath, func(w http.ResponseWriter, r *http.Request) {
		requestXMLBody := `<request type="payer-new" timestamp="20180614095000"><merchantid>realexsandbox</merchantid><account>internet</account><channel>ECOM</channel><orderid>AiCibJ5UR7utURy_slxhJw</orderid><payerref></payerref><paymentmethod>3c4af936-483e-a393-f558bec2fb2a</paymentmethod><sha1hash>59a88d763f26bdcbbf4dd65d3b0aec0b1dd5f6f6</sha1hash><amount currency="CAD">10000</amount><autosettle flag="1"></autosettle><paymentdata><cvn><number>123</number></cvn></paymentdata><payer ref="03e28f0e-492e-80bd-20ec318e9334" type="Retail" title="Mr."><firstname>James</firstname><surname>Mason</surname><company>Global Payments</company><email>text@mail.com</email><dateofbirth>19851222</dateofbirth><state>yorkshire</state><passphrase>montgomery</passphrase><vatnumber>gb 1234</vatnumber><varref>Car Part</varref><custnum>E8953893489</custnum><address><line1>Flat 123</line1><line2>House 123</line2><line3>Cul-de-sac</line3><city>Halifax</city><county>West Yorkshire</county><postcode>W6 9HR</postcode><country ref="GB"></country></address><phonenumbers><home>+35312345678</home><work>+3531987654321</work><fax>+3531987654321</fax><mobile>+3531987654321</mobile></phonenumbers></payer></request>`
		responseXMLBody := `<response timestamp="20180731090859">
							   <merchantid>MerchantId</merchantid>
							   <account>internet</account>
							   <orderid>N6qsk4kYRZihmPrTXWYS6g</orderid>
							   <authcode>12345</authcode>
							   <result>00</result>
							   <cvnresult>M</cvnresult>
							   <avspostcoderesponse>M</avspostcoderesponse>
							   <avsaddressresponse>M</avsaddressresponse>
							   <batchid>319623</batchid>
							   <message>[ test system ] AUTHORISED</message>
							   <pasref>14610544313177922</pasref>
							   <timetaken>1</timetaken>
							   <authtimetaken>0</authtimetaken>
							   <srd>MMC0F00YE4000000715</srd>
							   <cardissuer>
								  <bank>AIB BANK</bank>
								  <country>IRELAND</country>
								  <countrycode>IE</countrycode>
								  <region>EUR</region>
							   </cardissuer>
							   <sha1hash>77ac77956e57156f47142a5723835badf767e272</sha1hash>
							</response>`
		if got, want := r.Method, "POST"; got != want {
			t.Errorf("Request method: %v, want %v", got, want)
		}
		body, _ := ioutil.ReadAll(r.Body)
		if !reflect.DeepEqual(string(body), requestXMLBody) {
			t.Errorf("Request Body = %v, want %v", string(body), requestXMLBody)
		}
		fmt.Fprint(w, responseXMLBody)
	})

	response, _, err := client.CardStorage.CreateCustomer(creditRequest)
	if err != nil {
		t.Errorf("Error performing Client.Do: %v", err)
	}

	expectedResponse := &ServiceResponse{
		XMLName:             xml.Name{Local: "response"},
		Timestamp:           "20180731090859",
		MerchantID:          "MerchantId",
		Account:             "internet",
		OrderID:             "N6qsk4kYRZihmPrTXWYS6g",
		Result:              "00",
		Message:             "[ test system ] AUTHORISED",
		PasRef:              "14610544313177922",
		AuthCode:            "12345",
		CVNResult:           "M",
		AVSPostcodeResponse: "M",
		AVSAddressResponse:  "M",
		BatchID:             "319623",
		TimeTaken:           "1",
		AuthTimeTaken:       "0",
		CardIssuer: &CardIssuer{
			Bank:        "AIB BANK",
			Country:     "IRELAND",
			CountryCode: "IE",
			Region:      "EUR",
		},
		Sha1Hash:             "77ac77956e57156f47142a5723835badf767e272",
		serviceAuthenticator: serviceAuthenticator{elementsToHash: []string{"20180731090859", "MerchantId", "N6qsk4kYRZihmPrTXWYS6g", "00", "[ test system ] AUTHORISED", "14610544313177922", "12345"}, sharedSecret: "Po8lRRT67a"}}
	if !reflect.DeepEqual(response, expectedResponse) {
		t.Errorf("Request Body = %v, want %v", response, expectedResponse)
	}
}

func TestCardStorageService_EditCustomer(t *testing.T) {
	editCustomer := &CardStorageRequest{
		Account: "internet",
		Channel: "ECOM",
		OrderID: "AiCibJ5UR7utURy_slxhJw",
		Payer: &Payer{
			Ref:               "03e28f0e-492e-80bd-20ec318e9334",
			PayerType:         "Retail",
			Title:             "Mr.",
			FirstName:         "James",
			Surname:           "Mason",
			Company:           "Global Payments",
			Email:             "text@mail.com",
			DateOfBirth:       "19851222",
			State:             "yorkshire",
			PassPhrase:        "montgomery",
			VatNumber:         "gb 1234",
			VariableReference: "Car Part",
			CustomerNumber:    "E8953893489",
			Address: &Address{
				Line1:    "Flat 123",
				Line2:    "House 123",
				Line3:    "Cul-de-sac",
				City:     "Halifax",
				County:   "West Yorkshire",
				PostCode: "W6 9HR",
				Country:  &Country{Code: "GB"},
			},
			PhoneNumbers: &PhoneNumbers{
				Home:   "+35312345678",
				Work:   "+3531987654321",
				Fax:    "+3531987654321",
				Mobile: "+3531987654321",
			},
		},
		PaymentMethod: "3c4af936-483e-a393-f558bec2fb2a",
		Amount: &Amount{
			Amount:   "10000",
			Currency: "CAD",
		},
		AutoSettle:  &AutoSettle{Flag: "1"},
		PaymentData: &PaymentData{CVN: CVN{Number: "123"}},
	}

	client, mux, _, teardown := setup()
	defer teardown()
	mux.HandleFunc(apiURLPath, func(w http.ResponseWriter, r *http.Request) {
		requestXMLBody := `<request type="payer-edit" timestamp="20180614095000"><merchantid>realexsandbox</merchantid><account>internet</account><channel>ECOM</channel><orderid>AiCibJ5UR7utURy_slxhJw</orderid><payerref></payerref><paymentmethod>3c4af936-483e-a393-f558bec2fb2a</paymentmethod><sha1hash>9b1687d318b226fd1954f423e95507de6f481b30</sha1hash><amount currency="CAD">10000</amount><autosettle flag="1"></autosettle><paymentdata><cvn><number>123</number></cvn></paymentdata><payer ref="03e28f0e-492e-80bd-20ec318e9334" type="Retail" title="Mr."><firstname>James</firstname><surname>Mason</surname><company>Global Payments</company><email>text@mail.com</email><dateofbirth>19851222</dateofbirth><state>yorkshire</state><passphrase>montgomery</passphrase><vatnumber>gb 1234</vatnumber><varref>Car Part</varref><custnum>E8953893489</custnum><address><line1>Flat 123</line1><line2>House 123</line2><line3>Cul-de-sac</line3><city>Halifax</city><county>West Yorkshire</county><postcode>W6 9HR</postcode><country ref="GB"></country></address><phonenumbers><home>+35312345678</home><work>+3531987654321</work><fax>+3531987654321</fax><mobile>+3531987654321</mobile></phonenumbers></payer></request>`
		responseXMLBody := `<response timestamp="20180731090859">
							   <merchantid>MerchantId</merchantid>
							   <account>internet</account>
							   <orderid>N6qsk4kYRZihmPrTXWYS6g</orderid>
							   <authcode>12345</authcode>
							   <result>00</result>
							   <cvnresult>M</cvnresult>
							   <avspostcoderesponse>M</avspostcoderesponse>
							   <avsaddressresponse>M</avsaddressresponse>
							   <batchid>319623</batchid>
							   <message>[ test system ] AUTHORISED</message>
							   <pasref>14610544313177922</pasref>
							   <timetaken>1</timetaken>
							   <authtimetaken>0</authtimetaken>
							   <srd>MMC0F00YE4000000715</srd>
							   <cardissuer>
								  <bank>AIB BANK</bank>
								  <country>IRELAND</country>
								  <countrycode>IE</countrycode>
								  <region>EUR</region>
							   </cardissuer>
							   <sha1hash>77ac77956e57156f47142a5723835badf767e272</sha1hash>
							</response>`
		if got, want := r.Method, "POST"; got != want {
			t.Errorf("Request method: %v, want %v", got, want)
		}
		body, _ := ioutil.ReadAll(r.Body)
		if !reflect.DeepEqual(string(body), requestXMLBody) {
			t.Errorf("Request Body = %v, want %v", string(body), requestXMLBody)
		}
		fmt.Fprint(w, responseXMLBody)
	})

	response, _, err := client.CardStorage.EditCustomer(editCustomer)
	if err != nil {
		t.Errorf("Error performing Client.Do: %v", err)
	}

	expectedResponse := &ServiceResponse{
		XMLName:             xml.Name{Local: "response"},
		Timestamp:           "20180731090859",
		MerchantID:          "MerchantId",
		Account:             "internet",
		OrderID:             "N6qsk4kYRZihmPrTXWYS6g",
		Result:              "00",
		Message:             "[ test system ] AUTHORISED",
		PasRef:              "14610544313177922",
		AuthCode:            "12345",
		CVNResult:           "M",
		AVSPostcodeResponse: "M",
		AVSAddressResponse:  "M",
		BatchID:             "319623",
		TimeTaken:           "1",
		AuthTimeTaken:       "0",
		CardIssuer: &CardIssuer{
			Bank:        "AIB BANK",
			Country:     "IRELAND",
			CountryCode: "IE",
			Region:      "EUR",
		},
		Sha1Hash:             "77ac77956e57156f47142a5723835badf767e272",
		serviceAuthenticator: serviceAuthenticator{elementsToHash: []string{"20180731090859", "MerchantId", "N6qsk4kYRZihmPrTXWYS6g", "00", "[ test system ] AUTHORISED", "14610544313177922", "12345"}, sharedSecret: "Po8lRRT67a"}}
	if !reflect.DeepEqual(response, expectedResponse) {
		t.Errorf("Request Body = %v, want %v", response, expectedResponse)
	}
}

func TestCardStorageService_StoreCard(t *testing.T) {
	storeCardRequest := &CardStorageRequest{
		Account: "internet",
		OrderID: "F-2knQ0iShKK6ezfaSLh2Q",
		Card: &Card{
			Ref:            "3c4af936-483e-a393-f558bec2fb2a",
			PayerRef:       "0f357b45-9aa4-4453-a685-c69232e9024f",
			Number:         "4263970000005262",
			ExpDate:        "0519",
			CardHolderName: "James Mason",
			Type:           "VISA",
		}}

	client, mux, _, teardown := setup()
	defer teardown()
	mux.HandleFunc(apiURLPath, func(w http.ResponseWriter, r *http.Request) {
		requestXMLBody := `<request type="card-new" timestamp="20180614095000"><merchantid>realexsandbox</merchantid><account>internet</account><orderid>F-2knQ0iShKK6ezfaSLh2Q</orderid><payerref></payerref><sha1hash>c70058f6da8bb64202ce8096e3d7f36c2be4f781</sha1hash><card><ref>3c4af936-483e-a393-f558bec2fb2a</ref><payerref>0f357b45-9aa4-4453-a685-c69232e9024f</payerref><number>4263970000005262</number><expdate>0519</expdate><chname>James Mason</chname><type>VISA</type></card></request>`
		responseXMLBody := `<response timestamp="20180731090859">
							  <merchantid>MerchantId</merchantid>
							  <account>internet</account>
							  <orderid>N6qsk4kYRZihmPrTXWYS6g</orderid>
							  <result>00</result>
							  <message>Successful</message>
							  <pasref>14610544313177922</pasref>
							  <authcode/>
							  <batchid/>
							  <timetaken>1</timetaken>
							  <processingtimetaken/>
							  <sha1hash>a3084dac21a4fcbb8f66570f75db671998afce60</sha1hash>
							</response>`
		if got, want := r.Method, "POST"; got != want {
			t.Errorf("Request method: %v, want %v", got, want)
		}
		body, _ := ioutil.ReadAll(r.Body)
		if !reflect.DeepEqual(string(body), requestXMLBody) {
			t.Errorf("Request Body = %v, want %v", string(body), requestXMLBody)
		}
		fmt.Fprint(w, responseXMLBody)
	})

	response, _, err := client.CardStorage.StoreCard(storeCardRequest)
	if err != nil {
		t.Errorf("Error performing Client.Do: %v", err)
	}

	expectedResponse := &ServiceResponse{
		XMLName:              xml.Name{Local: "response"},
		Timestamp:            "20180731090859",
		MerchantID:           "MerchantId",
		Account:              "internet",
		OrderID:              "N6qsk4kYRZihmPrTXWYS6g",
		Result:               "00",
		Message:              "Successful",
		TimeTaken:            "1",
		PasRef:               "14610544313177922",
		Sha1Hash:             "a3084dac21a4fcbb8f66570f75db671998afce60",
		serviceAuthenticator: serviceAuthenticator{elementsToHash: []string{"20180731090859", "MerchantId", "N6qsk4kYRZihmPrTXWYS6g", "00", "Successful", "14610544313177922", ""}, sharedSecret: "Po8lRRT67a"}}
	if !reflect.DeepEqual(response, expectedResponse) {
		t.Errorf("Request Body = %v, want %v", response, expectedResponse)
	}
}

func TestCardStorageService_EditCard(t *testing.T) {
	editCardRequest := &CardStorageRequest{
		Account: "internet",
		OrderID: "F-2knQ0iShKK6ezfaSLh2Q",
		Card: &Card{
			Ref:            "3c4af936-483e-a393-f558bec2fb2a",
			PayerRef:       "0f357b45-9aa4-4453-a685-c69232e9024f",
			Number:         "4263970000005262",
			ExpDate:        "0519",
			CardHolderName: "James Mason",
			Type:           "VISA",
		}}

	client, mux, _, teardown := setup()
	defer teardown()
	mux.HandleFunc(apiURLPath, func(w http.ResponseWriter, r *http.Request) {
		requestXMLBody := `<request type="card-update-card" timestamp="20180614095000"><merchantid>realexsandbox</merchantid><account>internet</account><orderid>F-2knQ0iShKK6ezfaSLh2Q</orderid><payerref></payerref><sha1hash>206184f8fdc6813cc39231b309ca65c991f4b814</sha1hash><card><ref>3c4af936-483e-a393-f558bec2fb2a</ref><payerref>0f357b45-9aa4-4453-a685-c69232e9024f</payerref><number>4263970000005262</number><expdate>0519</expdate><chname>James Mason</chname><type>VISA</type></card></request>`
		responseXMLBody := `<response timestamp="20180731090859">
							  <merchantid>MerchantId</merchantid>
							  <account>internet</account>
							  <orderid>N6qsk4kYRZihmPrTXWYS6g</orderid>
							  <result>00</result>
							  <message>Successful</message>
							  <pasref>14610544313177922</pasref>
							  <authcode/>
							  <batchid/>
							  <timetaken>1</timetaken>
							  <processingtimetaken/>
							  <sha1hash>a3084dac21a4fcbb8f66570f75db671998afce60</sha1hash>
							</response>`
		if got, want := r.Method, "POST"; got != want {
			t.Errorf("Request method: %v, want %v", got, want)
		}
		body, _ := ioutil.ReadAll(r.Body)
		if !reflect.DeepEqual(string(body), requestXMLBody) {
			t.Errorf("Request Body = %v, want %v", string(body), requestXMLBody)
		}
		fmt.Fprint(w, responseXMLBody)
	})

	response, _, err := client.CardStorage.EditCard(editCardRequest)
	if err != nil {
		t.Errorf("Error performing Client.Do: %v", err)
	}

	expectedResponse := &ServiceResponse{
		XMLName:              xml.Name{Local: "response"},
		Timestamp:            "20180731090859",
		MerchantID:           "MerchantId",
		Account:              "internet",
		OrderID:              "N6qsk4kYRZihmPrTXWYS6g",
		Result:               "00",
		Message:              "Successful",
		TimeTaken:            "1",
		PasRef:               "14610544313177922",
		Sha1Hash:             "a3084dac21a4fcbb8f66570f75db671998afce60",
		serviceAuthenticator: serviceAuthenticator{elementsToHash: []string{"20180731090859", "MerchantId", "N6qsk4kYRZihmPrTXWYS6g", "00", "Successful", "14610544313177922", ""}, sharedSecret: "Po8lRRT67a"}}
	if !reflect.DeepEqual(response, expectedResponse) {
		t.Errorf("Request Body = %v, want %v", response, expectedResponse)
	}
}

func TestCardStorageService_DeleteCard(t *testing.T) {
	editCardRequest := &CardStorageRequest{
		Account: "internet",
		OrderID: "F-2knQ0iShKK6ezfaSLh2Q",
		Card: &Card{
			Ref:      "3c4af936-483e-a393-f558bec2fb2a",
			PayerRef: "0f357b45-9aa4-4453-a685-c69232e9024f"}}

	client, mux, _, teardown := setup()
	defer teardown()
	mux.HandleFunc(apiURLPath, func(w http.ResponseWriter, r *http.Request) {
		requestXMLBody := `<request type="card-update-card" timestamp="20180614095000"><merchantid>realexsandbox</merchantid><account>internet</account><orderid>F-2knQ0iShKK6ezfaSLh2Q</orderid><payerref></payerref><sha1hash>e2bbe2065a11c5d6829ab988f0c4234f78599363</sha1hash><card><ref>3c4af936-483e-a393-f558bec2fb2a</ref><payerref>0f357b45-9aa4-4453-a685-c69232e9024f</payerref><number></number><expdate></expdate><chname></chname><type></type></card></request>`
		responseXMLBody := `<response timestamp="20180731090859">
							  <merchantid>MerchantId</merchantid>
							  <account>internet</account>
							  <orderid>N6qsk4kYRZihmPrTXWYS6g</orderid>
							  <result>00</result>
							  <message>Successful</message>
							  <pasref>14610544313177922</pasref>
							  <authcode/>
							  <batchid/>
							  <timetaken>1</timetaken>
							  <processingtimetaken/>
							  <sha1hash>a3084dac21a4fcbb8f66570f75db671998afce60</sha1hash>
							</response>`
		if got, want := r.Method, "POST"; got != want {
			t.Errorf("Request method: %v, want %v", got, want)
		}
		body, _ := ioutil.ReadAll(r.Body)
		if !reflect.DeepEqual(string(body), requestXMLBody) {
			t.Errorf("Request Body = %v, want %v", string(body), requestXMLBody)
		}
		fmt.Fprint(w, responseXMLBody)
	})

	response, _, err := client.CardStorage.EditCard(editCardRequest)
	if err != nil {
		t.Errorf("Error performing Client.Do: %v", err)
	}

	expectedResponse := &ServiceResponse{
		XMLName:              xml.Name{Local: "response"},
		Timestamp:            "20180731090859",
		MerchantID:           "MerchantId",
		Account:              "internet",
		OrderID:              "N6qsk4kYRZihmPrTXWYS6g",
		Result:               "00",
		Message:              "Successful",
		TimeTaken:            "1",
		PasRef:               "14610544313177922",
		Sha1Hash:             "a3084dac21a4fcbb8f66570f75db671998afce60",
		serviceAuthenticator: serviceAuthenticator{elementsToHash: []string{"20180731090859", "MerchantId", "N6qsk4kYRZihmPrTXWYS6g", "00", "Successful", "14610544313177922", ""}, sharedSecret: "Po8lRRT67a"}}
	if !reflect.DeepEqual(response, expectedResponse) {
		t.Errorf("Request Body = %v, want %v", response, expectedResponse)
	}
}
