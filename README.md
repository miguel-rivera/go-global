# go-global

go-global is a Go client library for accessing Global Payments API. go-global does not implement Global Payments API fully, but only what is currently being used by the Payment Service Team.
 
## Table of Contents
TODO:



## Usage 
 
### Installation 
go-global is an internal go module. Go is using git to pull the specified versions of dependencies, the git configuration needs to contain the appropriate credentials to access any private repositories.

### Configuration
To Construct a new Global Payments client, then use the various services on the client to perform different operations pertaining to global payments api.
Following values can be configured by projects that wish to use this library. 

Default values will point to Global Payments sandbox environment. 


### Performing Requests.
The services of the go-global client divide the API into logical chunks and correspond to the structure of the Global Payments documentation found within Global Payments [API Explorer](https://developer.globalpay.com/api/getting-started).

#### Examples