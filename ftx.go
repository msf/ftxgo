package ftxgo

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type FTXCredentials struct {
	ApiKey    string
	SecretKey string
}

type FTXClient struct {
	httpClient *http.Client
	creds      FTXCredentials
}

func NewFTXClient(apiKey, secretKey string) *FTXClient {
	return &FTXClient{
		httpClient: &http.Client{},
		creds:      FTXCredentials{apiKey, secretKey},
	}
}

func (ftx *FTXClient) Request(req *http.Request, resp interface{}) error {
	out, err := ftx.httpClient.Do(ftx.signedRequest(req))
	if err != nil {
		return err
	}
	err = json.NewDecoder(out.Body).Decode(&resp)
	if err != nil {
		return err
	}
	return nil
}

func (ftx *FTXClient) signedRequest(req *http.Request) *http.Request {
	ts := time.Now()
	signature := fmt.Sprintf("%v-%v-%v", ts.UnixMilli(), req.Method, req.URL.Path)

	sign := hmac.New(sha256.New, []byte(ftx.creds.SecretKey))
	sign.Write([]byte(signature))
	signDigest := hex.EncodeToString(sign.Sum(nil))

	req.Header.Add("FTX-KEY", ftx.creds.ApiKey)
	req.Header.Add("FTX-SIGN", signDigest)
	req.Header.Add("FTX-TS", strconv.FormatInt(ts.UnixMilli(), 10))

	return req
}
