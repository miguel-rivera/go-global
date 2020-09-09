# go-global

go-global is a Go client library for accessing Global Payments API. This currently only supports the following Services.
[Card Storage](https://developer.globalpay.com/api/card-storage)

## Usage 
 
### Installation 
```go
import "github.com/miguel-rivera/go-global/globalpayments" 
```


### Configuration
Construct a new Global Payments client that can then beused with various services on the client to access different parts of Global Payments API.
The services of the go-global client divide the API into logical chunks and correspond to the structure of the Global Payments documentation found within Global Payments [API Explorer](https://developer.globalpay.com/api/getting-started).

For Example: 

```go
client, err := globalpayments.NewClient() // default sandbox credentials 

authRequest := &globalpayments.CardStorageRequest{}

authResponse, httpResponse, err := client.CardStorage.authorize(authRequest)
```

Global Payments Client can be updated with merchant specific credentials using variadic functional options

For Example: 

```go
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

	client, _ := NewClient(baseUrl, hashSecret, merchantId, setHttpClient)
```