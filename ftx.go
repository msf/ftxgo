package ftxgo

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

type FTXCredentials struct {
	APIKey    string
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
	unixTS := strconv.FormatInt(ts.UnixMilli(), 10)

	sign := hmac.New(sha256.New, []byte(ftx.creds.SecretKey))
	sign.Write([]byte(unixTS))
	sign.Write([]byte(req.Method))
	sign.Write([]byte(req.URL.Path))
	if req.URL.RawQuery != "" {
		sign.Write([]byte("?" + req.URL.RawQuery))
	} else if req.Method == "POST" {
		body, err := req.GetBody()
		if err != nil {
			log.Errorf("signedRequest w/ POST, failed to get body: %v", err)
			return req
		}
		defer body.Close()
		_, err = io.Copy(sign, body)
		if err != nil {
			log.Errorf("signedRequest w/ POST, failed to write body on signature: %v", err)
			return req
		}
	}

	signDigest := hex.EncodeToString(sign.Sum(nil))

	req.Header.Add("FTX-KEY", ftx.creds.APIKey)
	req.Header.Add("FTX-SIGN", signDigest)
	req.Header.Add("FTX-TS", unixTS)
	return req
}
