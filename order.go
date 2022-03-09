package ftxgo

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
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
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Result  struct {
		CreatedAt     time.Time `json:"createdAt"`
		FilledSize    float64   `json:"filledSize"`
		Future        string    `json:"future"`
		ID            int       `json:"id"`
		Market        string    `json:"market"`
		Price         float64   `json:"price"`
		RemainingSize float64   `json:"remainingSize"`
		Side          string    `json:"side"`
		Size          float64   `json:"size"`
		Status        string    `json:"status"`
		Type          string    `json:"type"`
		ReduceOnly    bool      `json:"reduceOnly"`
		Ioc           bool      `json:"ioc"`
		PostOnly      bool      `json:"postOnly"`
		ClientID      string    `json:"clientId"`
	} `json:"result"`
}

func (ftx *FTXClient) PostBuyOrder(market string, price, quantity float64, doIt bool) (resp PlaceOrderResponse, err error) {
	ts := time.Now()
	defer func() {
		log.WithFields(log.Fields{
			"elapsed":      time.Since(ts),
			"err":          err,
			"market":       market,
			"price":        price,
			"quantity":     quantity,
			"resp.Error":   resp.Error,
			"resp.Success": resp.Success,
		}).Info(("PostBuyOrder()"))
	}()

	order := &PlaceOrderRequest{
		Market: market,
		Side:   "buy",
		Type:   "limit",
		Size:   quantity,
		Price:  price,
	}

	buf := new(bytes.Buffer)
	err = json.NewEncoder(buf).Encode(order)
	if err != nil {
		return
	}
	req, err := http.NewRequest("POST", "https://ftx.com/api/orders", buf)
	if err != nil {
		return
	}

	err = ftx.Request(req, &resp)
	return
}
