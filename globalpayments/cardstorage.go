package globalpayments

import (
	"encoding/xml"
	"net/http"
	"time"
)

//CardStorageRequest request struct for all apis
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

//Amount request struct
type Amount struct {
	Amount   string `xml:",chardata"`
	Currency string `xml:"currency,attr"`
}

//AutoSettle request struct
type AutoSettle struct {
	Flag string `xml:"flag,attr"`
}

//PaymentData request struct
type PaymentData struct {
	CVN CVN `xml:"cvn"`
}

//CVN request struct
type CVN struct {
	Number string `xml:"number"`
}

//Payer request struct
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

//PhoneNumbers request struct
type PhoneNumbers struct {
	Home   string `xml:"home"`
	Work   string `xml:"work"`
	Fax    string `xml:"fax"`
	Mobile string `xml:"mobile"`
}

//Address request struct
type Address struct {
	Line1    string   `xml:"line1"`
	Line2    string   `xml:"line2"`
	Line3    string   `xml:"line3"`
	City     string   `xml:"city"`
	County   string   `xml:"county"`
	PostCode string   `xml:"postcode"`
	Country  *Country `xml:"country"`
}

//Country request struct
type Country struct {
	Code string `xml:"ref,attr"`
}

//Card request struct
type Card struct {
	Ref            string `xml:"ref"`
	PayerRef       string `xml:"payerref"`
	Number         string `xml:"number"`
	ExpDate        string `xml:"expdate"`
	CardHolderName string `xml:"chname"`
	Type           string `xml:"type"`
}

//CardStorageService  Card Storage API offers a range of easy-to-use requests to store, charge, update and delete cards.
type CardStorageService struct {
	service
}

//CardStorageServiceAPI interface contain all request types that are allowed within this service for mocking on upstream consumers
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

//TimeFormatter interface
type TimeFormatter interface {
	Format(layout string) string
}

//Now monkey patching for replacing time.Now in unit tests
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

//Authorize Once you have the card stored you can easily raise an authorization against it. In the place of collected card data you
//simply send the customer token (Payer) and card (Payment Method) reference, Global Payments obtains the securely stored
//card data from our vault and builds an authorization which we then send on to the Issuer.
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

//Validate Open to Buy (OTB) allows you to check that a stored card is still valid and active without actually processing a payment
//against it. This is an alternative to charging the card a small amount (for example 10c) to obtain the same result.
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

//Credit request type allows you to credit an amount to a stored card.
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

//CreateCustomer In order to store a card, the first thing we need to do is set up a customer reference (Payer). You can also choose to
//store address and contact details alongside it.
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

//EditCustomer Once a customer has been created you can update their name, address or contact details which can be viewed in Ecommerce Portal.
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

//StoreCard Once we have our customer entity created, we can now add cards to it. This request must contain the card data to be stored,
//a unique reference for it and the customer reference it is to be added to. We'd always recommend processing an authorization
//against a card or validating it (OTB) before adding it.
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

//EditCard If your customer's card details change, for example if the expiry date is updated or they get a new card number, you can
//update the reference in Card Storage using this request. You can update all the card details at once or just the individual
//bits of data, for example just the expiry date. In the example below we are completely replacing the card with a new one.
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

//DeleteCard If you want to remove a card from Card Storage you can send us a Card Delete request.
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
