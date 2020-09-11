package globalpayments

import (
	"encoding/xml"
	"net/http"
	"time"
)

type CardStorageRequest struct {
	XMLName       xml.Name     `xml:"request"`
	Type          string       `xml:"type,attr"`
	Timestamp     string       `xml:"timestamp,attr"`
	MerchantID    string       `xml:"merchantid"`
	Account       string       `xml:"account,omitempty"`
	Channel       string       `xml:"channel,omitempty"`
	OrderID       string       `xml:"orderid"`
	PayerRef      string       `xml:"payerref"`
	PaymentMethod string       `xml:"paymentmethod,omitempty"`
	Sha1Hash      string       `xml:"sha1hash"`
	Amount        *Amount      `xml:"amount,omitempty"`
	AutoSettle    *AutoSettle  `xml:"autosettle,omitempty"`
	PaymentData   *PaymentData `xml:"paymentdata,omitempty"`
	Payer         *Payer       `xml:"payer,omitempty"`
	Card          *Card        `xml:"card,omitempty"`
	serviceAuthenticator
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
	Line1    string   `xml:"line1"`
	Line2    string   `xml:"line2"`
	Line3    string   `xml:"line3"`
	City     string   `xml:"city"`
	County   string   `xml:"county"`
	PostCode string   `xml:"postcode"`
	Country  *Country `xml:"country"`
}

type Country struct {
	Code string `xml:"ref,attr"`
}

type Card struct {
	Ref            string `xml:"ref"`
	PayerRef       string `xml:"payerref"`
	Number         string `xml:"number"`
	ExpDate        string `xml:"expdate"`
	CardHolderName string `xml:"chname"`
	Type           string `xml:"type"`
}

type CardStorageService struct {
	service
}

type CardStorageServiceAPI interface {
	Authorize(request *CardStorageRequest) (*ServiceResponse, *http.Response,
		error)
	Validate(request *CardStorageRequest) (*ServiceResponse, *http.Response,
		error)
	Credit(request *CardStorageRequest) (*ServiceResponse, *http.Response,
		error)
	CreateCustomer(request *CardStorageRequest) (*ServiceResponse, *http.Response,
		error)
	EditCustomer(request *CardStorageRequest) (*ServiceResponse, *http.Response,
		error)
	StoreCard(request *CardStorageRequest) (*ServiceResponse, *http.Response,
		error)
	EditCard(request *CardStorageRequest) (*ServiceResponse, *http.Response,
		error)
	DeleteCard(request *CardStorageRequest) (*ServiceResponse, *http.Response,
		error)
}

type TimeFormatter interface {
	Format(layout string) string
}

var Now = time.Now

func formatTime(t TimeFormatter, layout string) string {
	return t.Format(layout)
}

//used getters for objects used within the hash

func (request CardStorageRequest) getPayerRef() string {
	if request.Payer != nil {
		return request.Payer.Ref
	}
	return ""
}

func (request CardStorageRequest) getAmount() string {
	if request.Amount != nil {
		return request.Amount.Amount
	}
	return ""
}

func (request CardStorageRequest) getCurrency() string {
	if request.Amount != nil {
		return request.Amount.Currency
	}
	return ""
}

func (request CardStorageRequest) getCardHolderName() string {
	if request.Card != nil {
		return request.Card.CardHolderName
	}
	return ""
}

func (request CardStorageRequest) getCardNumber() string {
	if request.Card != nil {
		return request.Card.Number
	}
	return ""
}

func (request CardStorageRequest) getCardRef() string {
	if request.Card != nil {
		return request.Card.Ref
	}
	return ""
}

func (request CardStorageRequest) getCardExpDate() string {
	if request.Card != nil {
		return request.Card.ExpDate
	}
	return ""
}
func (cardStorage *CardStorageService) Authorize(request *CardStorageRequest) (*ServiceResponse, *http.Response,
	error) {
	request.Timestamp = formatTime(Now(), "20060102150405")
	request.MerchantID = cardStorage.client.MerchantID
	request.Type = "receipt-in"
	request.elementsToHash = []string{request.Timestamp, request.MerchantID, request.OrderID, request.Amount.Amount, request.Amount.Currency, request.PayerRef}
	request.sharedSecret = cardStorage.client.HashSecret
	signature, err := request.buildSignature()
	if err != nil {
		return nil, nil, err
	}
	request.Sha1Hash = signature
	return cardStorage.transmitRequest(request)
}

func (cardStorage *CardStorageService) Validate(request *CardStorageRequest) (*ServiceResponse, *http.Response,
	error) {
	request.Timestamp = formatTime(Now(), "20060102150405")
	request.MerchantID = cardStorage.client.MerchantID
	request.Type = "receipt-in-otb"
	request.elementsToHash = []string{request.Timestamp, request.MerchantID, request.OrderID, request.PayerRef}
	request.sharedSecret = cardStorage.client.HashSecret
	signature, err := request.buildSignature()
	if err != nil {
		return nil, nil, err
	}
	request.Sha1Hash = signature
	return cardStorage.transmitRequest(request)
}

func (cardStorage *CardStorageService) Credit(request *CardStorageRequest) (*ServiceResponse, *http.Response,
	error) {
	request.Timestamp = formatTime(Now(), "20060102150405")
	request.MerchantID = cardStorage.client.MerchantID
	request.Type = "payment-out"
	request.elementsToHash = []string{request.Timestamp, request.MerchantID, request.OrderID, request.getAmount(), request.getCurrency(), request.PayerRef}
	request.sharedSecret = cardStorage.client.RebateHashSecret
	signature, err := request.buildSignature()
	if err != nil {
		return nil, nil, err
	}
	request.Sha1Hash = signature
	return cardStorage.transmitRequest(request)
}

func (cardStorage *CardStorageService) CreateCustomer(request *CardStorageRequest) (*ServiceResponse, *http.Response,
	error) {
	request.Timestamp = formatTime(Now(), "20060102150405")
	request.MerchantID = cardStorage.client.MerchantID
	request.Type = "payer-new"
	request.elementsToHash = []string{request.Timestamp, request.MerchantID, request.OrderID, request.getAmount(), request.getCurrency(), request.getPayerRef()}
	request.sharedSecret = cardStorage.client.HashSecret
	signature, err := request.buildSignature()
	if err != nil {
		return nil, nil, err
	}
	request.Sha1Hash = signature
	return cardStorage.transmitRequest(request)
}

func (cardStorage *CardStorageService) EditCustomer(request *CardStorageRequest) (*ServiceResponse, *http.Response,
	error) {
	request.Timestamp = formatTime(Now(), "20060102150405")
	request.MerchantID = cardStorage.client.MerchantID
	request.Type = "payer-edit"
	request.elementsToHash = []string{request.Timestamp, request.MerchantID, request.OrderID, request.getAmount(), request.getCurrency(), request.PayerRef}
	request.sharedSecret = cardStorage.client.HashSecret
	signature, err := request.buildSignature()
	if err != nil {
		return nil, nil, err
	}
	request.Sha1Hash = signature
	return cardStorage.transmitRequest(request)
}

func (cardStorage *CardStorageService) StoreCard(request *CardStorageRequest) (*ServiceResponse, *http.Response,
	error) {
	request.Timestamp = formatTime(Now(), "20060102150405")
	request.MerchantID = cardStorage.client.MerchantID
	request.Type = "card-new"
	request.elementsToHash = []string{request.Timestamp, request.MerchantID, request.OrderID, request.getAmount(), request.getCurrency(), request.PayerRef, request.getCardHolderName(), request.getCardNumber()}
	request.sharedSecret = cardStorage.client.HashSecret
	signature, err := request.buildSignature()
	if err != nil {
		return nil, nil, err
	}
	request.Sha1Hash = signature
	return cardStorage.transmitRequest(request)
}

func (cardStorage *CardStorageService) EditCard(request *CardStorageRequest) (*ServiceResponse, *http.Response,
	error) {
	request.Timestamp = formatTime(Now(), "20060102150405")
	request.MerchantID = cardStorage.client.MerchantID
	request.Type = "card-update-card"
	request.elementsToHash = []string{request.Timestamp, request.MerchantID, request.PayerRef, request.getCardRef(), request.getCardExpDate(), request.getCardNumber()}
	request.sharedSecret = cardStorage.client.HashSecret
	signature, err := request.buildSignature()
	if err != nil {
		return nil, nil, err
	}
	request.Sha1Hash = signature
	return cardStorage.transmitRequest(request)
}

func (cardStorage *CardStorageService) DeleteCard(request *CardStorageRequest) (*ServiceResponse, *http.Response,
	error) {
	request.Timestamp = formatTime(Now(), "20060102150405")
	request.MerchantID = cardStorage.client.MerchantID
	request.Type = "card-cancel-card"
	request.elementsToHash = []string{request.Timestamp, request.MerchantID, request.PayerRef, request.getCardRef()}
	request.sharedSecret = cardStorage.client.HashSecret
	signature, err := request.buildSignature()
	if err != nil {
		return nil, nil, err
	}
	request.Sha1Hash = signature
	return cardStorage.transmitRequest(request)
}
