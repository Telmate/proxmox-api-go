package api_test

import (
	"context"
	"crypto/tls"

	pxapi "github.com/Telmate/proxmox-api-go/proxmox"
)

type Test struct {
	APIurl      string
	UserID      string
	Password    string
	OTP         string
	HttpHeaders string
	RequireSSL  bool

	_client *pxapi.Client
}

func (test *Test) CreateClient() (err error) {
	if test.APIurl == "" {
		test.APIurl = "https://127.0.0.1:8006/api2/json"
	}
	if test.UserID == "" {
		test.UserID = "root@pam"
	}
	if test.Password == "" {
		test.Password = "root"
	}
	tlsConfig := &tls.Config{InsecureSkipVerify: true}

	if test.RequireSSL {
		tlsConfig = nil
	}

	test._client, err = pxapi.NewClient(test.APIurl, nil, test.HttpHeaders, tlsConfig, "", 300)
	return err
}

func (test *Test) GetClient() (client *pxapi.Client) {
	return test._client
}

func (test *Test) Login() (err error) {
	if test._client == nil {
		err = test.CreateClient()
		if err != nil {
			return err
		}
	}
	err = test._client.Login(context.Background(), test.UserID, test.Password, test.OTP)
	return err
}

func (test *Test) CreateTest() (err error) {
	err = test.Login()
	return err
}
