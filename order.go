package ftxgo

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

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

func (ftx *FTXClient) PostBuyOrder(market string, price, quantity float64) (*PlaceOrderResponse, error) {
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

	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(order)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", "https://ftx.com/api/orders", buf)
	if err != nil {
		return nil, err
	}

	var resp PlaceOrderResponse
	err = ftx.Request(req, &resp)
	return &resp, err
}
