package globalpayments

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type CardStorageRequest struct {
	XMLName        xml.Name     `xml:"request"`
	Type           string       `xml:"type,attr"`
	Timestamp      string       `xml:"timestamp,attr"`
	MerchantID     string       `xml:"merchantid"`
	Account        string       `xml:"account,omitempty"`
	Channel        string       `xml:"channel,omitempty"`
	OrderID        string       `xml:"orderid"`
	PayerRef       string       `xml:"payerref"`
	PaymentMethod  string       `xml:"paymentmethod,omitempty"`
	Sha1Hash       string       `xml:"sha1hash"`
	Amount         *Amount      `xml:"amount,omitempty"`
	AutoSettle     *AutoSettle  `xml:"autosettle,omitempty"`
	PaymentData    *PaymentData `xml:"paymentdata,omitempty"`
	Card           *Card        `xml:"card,omitempty"`
	elementsToHash []string
	sharedSecret   string
}

type Amount struct {
	Amount   string `xml:",chardata"`
	Currency string `xml:"currency,attr"`
}

type AutoSettle struct {
	Flag string `xml:"flag,attr"`
}

type PaymentData struct {
	CVN CVN `xml:"cvn"`
}

type CVN struct {
	Number string `xml:"number"`
}

type Payer struct {
	Ref               string        `xml:"ref,attr"`
	PayerType         string        `xml:"type,attr"`
	Title             string        `xml:"title,attr"`
	FirstName         string        `xml:"firstname"`
	Surname           string        `xml:"surname"`
	Company           string        `xml:"company"`
	Email             string        `xml:"email"`
	DateOfBirth       string        `xml:"dateofbirth"`
	State             string        `xml:"state"`
	PassPhrase        string        `xml:"passphrase"`
	VatNumber         string        `xml:"vatnumber"`
	VariableReference string        `xml:"varref"`
	CustomerNumber    string        `xml:"custnum"`
	Address           *Address      `xml:"address"`
	PhoneNumbers      *PhoneNumbers `xml:"phonenumbers"`
}

type PhoneNumbers struct {
	Home   string `xml:"home"`
	Work   string `xml:"work"`
	Fax    string `xml:"fax"`
	Mobile string `xml:"mobile"`
}

type Address struct {
	Line1    string `xml:"line1"`
	Line2    string `xml:"line2"`
	Line3    string `xml:"line3"`
	City     string `xml:"city"`
	County   string `xml:"county"`
	PostCode string `xml:"postcode"`
	Country  string `xml:"country"`
}

type Card struct {
	Ref            string `xml:"ref"`
	PayerRef       string `xml:"payerref"`
	Number         string `xml:"number"`
	ExpDate        string `xml:"expdate"`
	CardHolderName string `xml:"chname"`
	Type           string `xml:"type"`
}

type CardStorageResponse struct {
	XMLName             xml.Name `xml:"response"`
	Timestamp           string   `xml:"timestamp,attr"`
	Account             string   `xml:"account"`
	OrderID             string   `xml:"orderid"`
	AuthCode            string   `xml:"authcode"`
	Result              string   `xml:"result"`
	CVN                 string   `xml:"cvnresult"`
	AVSPostcodeResponse string   `xml:"avspostcoderesponse"`
	AVSAddressResponse  string   `xml:"avsaddressresponse"`
	BatchId             string   `xml:"batchid"`
	Message             string   `xml:"message"`
	PasRef              string   `xml:"pasref"`
	TimeTaken           string   `xml:"timetaken"`
	AuthTimeTaken       string   `xml:"authtimetaken"`
	CardIssuer          string   `xml:"cardIssuer"`
	SHA1Hash            string   `xml:"sha1hash"`
}

type CardIssuer struct {
	Bank        string `xml:"bank"`
	Country     string `xml:"country"`
	CountryCode string `xml:"countrycode"`
	Region      string `xml:"region"`
}

type CardStorageService service

type ValidationError struct {
	Response *http.Response
}

type CardStorageServiceAPI interface {
	Authorize(request *CardStorageRequest) (*CardStorageResponse, *http.Response,
		error)
	Validate(request *CardStorageRequest) (*CardStorageResponse, *http.Response,
		error)
	Credit(request *CardStorageRequest) (*CardStorageResponse, *http.Response,
		error)
	CreateCustomer(request *CardStorageRequest) (*CardStorageResponse, *http.Response,
		error)
	EditCustomer(request *CardStorageRequest) (*CardStorageResponse, *http.Response,
		error)
	StoreCard(request *CardStorageRequest) (*CardStorageResponse, *http.Response,
		error)
	EditCard(request *CardStorageRequest) (*CardStorageResponse, *http.Response,
		error)
	DeleteCard(request *CardStorageRequest) (*CardStorageResponse, *http.Response,
		error)
	validateResponseHash(httpResponse *http.Response, response *CardStorageResponse) (err error)
	hashAndEncode(m Marshaller, str string) (hashAndEncodedString string, err error)
	buildRequestHash(request *CardStorageRequest) (err error)
	formatTime(t TimeFormatter, layout string) string
	addCommonValuesToRequest(request *CardStorageRequest) (err error)
	transmitRequest(request *CardStorageRequest) (response *CardStorageResponse, httpResponse *http.Response,
		err error)
}

type Marshaller interface {
	io.Writer
	Sum(b []byte) []byte
}

type TimeFormatter interface {
	Format(layout string) string
}

func (err *ValidationError) Error() string {
	return fmt.Sprintf("Validation Hash Error: %v, %v : %d", err.Response.Request.Method,
		err.Response.Request.URL, err.Response.StatusCode)
}

func (cardStorage *CardStorageService) Authorize(request *CardStorageRequest) (*CardStorageResponse, *http.Response,
	error) {
	request.Type = "receipt-in"
	request.elementsToHash = []string{request.Timestamp, request.MerchantID, request.OrderID, request.Amount.Amount, request.Amount.Currency, request.PayerRef}
	request.sharedSecret = cardStorage.client.HashSecret
	cardStorage.addCommonValuesToRequest(request)
	return cardStorage.transmitRequest(request)
}

func (cardStorage *CardStorageService) Validate(request *CardStorageRequest) (*CardStorageResponse, *http.Response,
	error) {
	request.Type = "receipt-in-otb"
	request.elementsToHash = []string{request.Timestamp, request.MerchantID, request.OrderID, request.PayerRef}
	request.sharedSecret = cardStorage.client.HashSecret
	cardStorage.addCommonValuesToRequest(request)
	return cardStorage.transmitRequest(request)
}

func (cardStorage *CardStorageService) Credit(request *CardStorageRequest) (*CardStorageResponse, *http.Response,
	error) {
	request.Type = "payment-outs"
	request.elementsToHash = []string{request.Timestamp, request.MerchantID, request.OrderID, request.Amount.Amount, request.Amount.Currency, request.PayerRef}
	request.sharedSecret = cardStorage.client.RebateHashSecret
	cardStorage.addCommonValuesToRequest(request)
	return cardStorage.transmitRequest(request)
}

func (cardStorage *CardStorageService) CreateCustomer(request *CardStorageRequest) (*CardStorageResponse, *http.Response,
	error) {
	request.Type = "payer-new"
	request.elementsToHash = []string{request.Timestamp, request.MerchantID, request.OrderID, request.Amount.Amount, request.Amount.Currency, request.PayerRef}
	request.sharedSecret = cardStorage.client.HashSecret
	cardStorage.addCommonValuesToRequest(request)
	return cardStorage.transmitRequest(request)
}

func (cardStorage *CardStorageService) EditCustomer(request *CardStorageRequest) (*CardStorageResponse, *http.Response,
	error) {
	request.Type = "payer-edit"
	request.elementsToHash = []string{request.Timestamp, request.MerchantID, request.OrderID, request.Amount.Amount, request.Amount.Currency, request.PayerRef}
	request.sharedSecret = cardStorage.client.HashSecret
	cardStorage.addCommonValuesToRequest(request)
	return cardStorage.transmitRequest(request)
}

func (cardStorage *CardStorageService) StoreCard(request *CardStorageRequest) (*CardStorageResponse, *http.Response,
	error) {
	request.Type = "card-new"
	request.elementsToHash = []string{request.Timestamp, request.MerchantID, request.OrderID, request.Amount.Amount, request.Amount.Currency, request.PayerRef, request.Card.CardHolderName, request.Card.Number}
	request.sharedSecret = cardStorage.client.HashSecret
	cardStorage.addCommonValuesToRequest(request)
	return cardStorage.transmitRequest(request)
}

func (cardStorage *CardStorageService) EditCard(request *CardStorageRequest) (*CardStorageResponse, *http.Response,
	error) {
	request.Type = "card-update-card"
	request.elementsToHash = []string{request.Timestamp, request.MerchantID, request.PayerRef, request.Card.Ref, request.Card.ExpDate, request.Card.Number}
	request.sharedSecret = cardStorage.client.HashSecret
	cardStorage.addCommonValuesToRequest(request)
	return cardStorage.transmitRequest(request)
}

func (cardStorage *CardStorageService) DeleteCard(request *CardStorageRequest) (*CardStorageResponse, *http.Response,
	error) {
	request.Type = "card-cancel-card"
	request.elementsToHash = []string{request.Timestamp, request.MerchantID, request.PayerRef, request.Card.Ref}
	request.sharedSecret = cardStorage.client.HashSecret
	cardStorage.addCommonValuesToRequest(request)
	return cardStorage.transmitRequest(request)
}

func (cardStorage *CardStorageService) transmitRequest(request *CardStorageRequest) (response *CardStorageResponse, httpResponse *http.Response,
	err error) {

	response = &CardStorageResponse{}
	httpRequest, err := cardStorage.client.NewRequest("POST", cardStorage.Path, request)

	if err != nil {
		return nil, nil, err
	}

	httpResponse, err = cardStorage.client.Do(httpRequest, response)
	if err != nil {
		return nil, httpResponse, err
	}

	err = cardStorage.validateResponseHash(httpResponse, response)
	if err != nil {
		return nil, httpResponse, err
	}

	return response, httpResponse, nil
}

func (cardStorage *CardStorageService) addCommonValuesToRequest(request *CardStorageRequest) (err error) {
	request.Timestamp = cardStorage.formatTime(time.Now(), "20060102150405")
	request.MerchantID = cardStorage.client.MerchantID
	err = cardStorage.buildRequestHash(request)
	return err
}

func (cardStorage *CardStorageService) formatTime(t TimeFormatter, layout string) string {
	return t.Format(layout)
}

func (cardStorage *CardStorageService) buildRequestHash(request *CardStorageRequest) (err error) {

	hashedElementsString, err := cardStorage.hashAndEncode(sha1.New(), strings.Join(request.elementsToHash, "."))
	if err != nil {
		return err
	}

	requestHash, err := cardStorage.hashAndEncode(sha1.New(), hashedElementsString+"."+request.sharedSecret)
	if err != nil {
		return err
	}
	request.Sha1Hash = requestHash
	return nil
}

func (cardStorage *CardStorageService) hashAndEncode(m Marshaller, str string) (hashAndEncodedString string, err error) {

	_, err = io.WriteString(m, str)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(m.Sum(nil)), nil
}

func (cardStorage *CardStorageService) validateResponseHash(httpResponse *http.Response, response *CardStorageResponse) (err error) {

	elementsToHash := []string{response.Timestamp, cardStorage.client.MerchantID, response.OrderID, response.Result, response.Message, response.PasRef, response.AuthCode}

	h1 := sha1.New()

	_, err = io.WriteString(h1, strings.Join(elementsToHash, "."))
	if err != nil {
		return err
	}
	hashedElementsString := hex.EncodeToString(h1.Sum(nil))

	h2 := sha1.New()
	_, err = io.WriteString(h2, hashedElementsString+"."+cardStorage.client.HashSecret)
	if err != nil {
		return err
	}
	responseHash := hex.EncodeToString(h1.Sum(nil))

	if responseHash == response.SHA1Hash {
		return nil
	}
	return &ValidationError{httpResponse}
}
