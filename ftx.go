package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/namsral/flag"
)

type FTXCredentials struct {
	ApiKey    string
	SecretKey string
}

type OrderBook struct {
	Success bool `json:"success"`
	Result  struct {
		Asks [][2]float64 `json:"asks"`
		Bids [][2]float64 `json:"bids"`
	} `json:"result"`
}

type PlaceOrderRequest struct {
	Market     string      `json:"market"`
	Side       string      `json:"side"`
	Price      float64     `json:"price"`
	Type       string      `json:"type"`
	Size       float64     `json:"size"`
	ReduceOnly bool        `json:"reduceOnly"`
	Ioc        bool        `json:"ioc"`
	PostOnly   bool        `json:"postOnly"`
	ClientID   interface{} `json:"clientId"`
}

type PlaceOrderResponse struct {
	Success bool `json:"success"`
	Result  struct {
		CreatedAt     time.Time   `json:"createdAt"`
		FilledSize    int         `json:"filledSize"`
		Future        string      `json:"future"`
		ID            int         `json:"id"`
		Market        string      `json:"market"`
		Price         float64     `json:"price"`
		RemainingSize int         `json:"remainingSize"`
		Side          string      `json:"side"`
		Size          int         `json:"size"`
		Status        string      `json:"status"`
		Type          string      `json:"type"`
		ReduceOnly    bool        `json:"reduceOnly"`
		Ioc           bool        `json:"ioc"`
		PostOnly      bool        `json:"postOnly"`
		ClientID      interface{} `json:"clientId"`
	} `json:"result"`
}

func signedRequest(creds FTXCredentials, req *http.Request) *http.Request {
	ts := time.Now()
	signature := fmt.Sprintf("%v-%v-%v", ts.UnixMilli(), req.Method, req.URL.Path)

	sign := hmac.New(sha256.New, []byte(creds.SecretKey))
	sign.Write([]byte(signature))
	signDigest := hex.EncodeToString(sign.Sum(nil))

	req.Header.Add("FTX-KEY", creds.ApiKey)
	req.Header.Add("FTX-SIGN", signDigest)
	req.Header.Add("FTX-TS", strconv.FormatInt(ts.UnixMilli(), 10))

	return req
}

const ETH_EUR = "ETH/EUR"

func getPrice(creds FTXCredentials, market string) (float64, error) {
	ts := time.Now()
	defer func() {
		log.Printf("getPrice() took: %v\n", time.Since(ts))
	}()

	url := fmt.Sprintf("https://ftx.com/api/markets/%v/orderbook?depth=16", market)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0.0, err
	}
	cli := &http.Client{}
	out, err := cli.Do(signedRequest(creds, req))
	if err != nil {
		return 0.0, err
	}

	var orders OrderBook
	err = json.NewDecoder(out.Body).Decode(&orders)
	if err != nil {
		return 0.0, err
	}
	asks := orders.Result.Asks
	bids := orders.Result.Bids
	sort.Slice(asks, func(i, j int) bool {
		return asks[i][0] < asks[j][0]
	})
	sort.Slice(bids, func(i, j int) bool {
		return bids[i][0] > bids[j][0]
	})
	bidPrice := bids[0][0]
	askPrice := asks[0][0]
	log.Printf("%v bids: %v \n", market, bids[0])
	log.Printf("%v asks: %v\n", market, asks[0])
	log.Printf("%v gap: %.3f\n", market, askPrice-bidPrice)
	return askPrice, nil
}

func calcQuantity(price, budget float64) float64 {
	// budget = price * quantity
	return budget / price
}

func placeBuyOrder(creds FTXCredentials, market string, price, quantity float64) (*PlaceOrderResponse, error) {
	ts := time.Now()
	defer func() {
		log.Printf("placeBuyOrder() took: %v\n", time.Since(ts))
	}()

	order := &PlaceOrderRequest{
		Market: market,
		Side:   "buy",
		Type:   "limit",
		Size:   quantity,
		Price:  price,
	}

	cli := &http.Client{}
	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(order)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "https://ftx.com/api/orders", buf)
	if err != nil {
		return nil, err
	}
	out, err := cli.Do(signedRequest(creds, req))
	if err != nil {
		return nil, err
	}
	var resp PlaceOrderResponse
	err = json.NewDecoder(out.Body).Decode(&resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func main() {
	apiKey := flag.String("ftx_api_key", "", "FTX API Key")
	secretKey := flag.String("ftx_secret_key", "", "FTX Secret Key")
	budget := flag.Float64("budget", 50.0, "Budget in Euros to buy Eth")
	creds := FTXCredentials{*apiKey, *secretKey}
	price, err := getPrice(creds, ETH_EUR)
	if err != nil {
		log.Printf("Failed to getPrice: %v, aborting\n", err)
		os.Exit(1)
	}
	log.Printf("ETH price is: %v, %v", price, err)
	howMuch := calcQuantity(price, *budget)
	log.Printf("Placing BUY order: %.1f * %.6f = %v TOTAL\n", price, howMuch, price*howMuch)
	orderResult, err := placeBuyOrder(creds, ETH_EUR, price, howMuch)
	if err != nil || !orderResult.Success {
		log.Printf("Failed to placeBuyOrder(%v, %.1f, %.6f): %v, aborting\n", ETH_EUR, price, howMuch, err)
		os.Exit(1)
	}
	log.Printf("BUY order ETH: \n%+v\n", orderResult)

}
