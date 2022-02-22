package ftxgo

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type Order struct {
	Price     float64
	Quantity  float64
	Timestamp time.Time
	Market    string
	Status    string
	Side      string
}

func (o Order) Spend() float64 {
	return o.Quantity * o.Price
}

type FTXOrdersResponse struct {
	Success bool `json:"success"`
	Result  []struct {
		AvgFillPrice  float64   `json:"avgFillPrice"`
		ClientID      string    `json:"clientId"`
		CreatedAt     time.Time `json:"createdAt"`
		FilledSize    float64   `json:"filledSize"`
		Future        string    `json:"future"`
		ID            int       `json:"id"`
		Ioc           bool      `json:"ioc"`
		Market        string    `json:"market"`
		PostOnly      bool      `json:"postOnly"`
		Price         float64   `json:"price"`
		ReduceOnly    bool      `json:"reduceOnly"`
		RemainingSize int       `json:"remainingSize"`
		Side          string    `json:"side"`
		Size          float64   `json:"size"`
		Status        string    `json:"status"`
		Type          string    `json:"type"`
	} `json:"result"`
}

func (ftx *FTXClient) GetOpenBuyOrders(market string) (orders []Order, err error) {
	all, err := ftx.GetOpenOrders(market)
	if err != nil {
		return orders, err
	}
	for _, v := range all {
		if v.Side == "buy" {
			orders = append(orders, v)
		}
	}
	return orders, nil
}

func (ftx *FTXClient) GetOpenOrders(market string) (orders []Order, err error) {
	ts := time.Now()
	defer func() {
		log.Printf("GetOpenBuyOrder() took: %v\n", time.Since(ts))
	}()

	url := fmt.Sprintf("https://ftx.com/api/orders?market=%v", market)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return orders, err
	}
	var resp FTXOrdersResponse
	err = ftx.Request(req, &resp)
	if err != nil || !resp.Success {
		errors.Wrapf(err, "failed GetOpenOrders, resp.Success: %v", resp.Success)
		return orders, err
	}
	for _, v := range resp.Result {
		orders = append(orders, Order{
			Price:     v.Price,
			Status:    v.Status,
			Quantity:  v.Size,
			Timestamp: v.CreatedAt,
			Market:    v.Market,
			Side:      v.Side,
		})
	}
	return orders, nil
}

func (ftx *FTXClient) GetClosedBuyOrders(market string, interval time.Duration) (orders []Order, err error) {
	ts := time.Now()
	defer func() {
		log.Printf("GetPrice() took: %v\n", time.Since(ts))
	}()

	url := fmt.Sprintf("https://ftx.com/api/orders/history")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return orders, err
	}
	req.URL.Query().Add("market", market)
	req.URL.Query().Add("side", "buy")
	unixTime := time.Now().Add(-interval).Unix()

	req.URL.Query().Add("start_time", strconv.FormatInt(unixTime, 10))
	var resp FTXOrdersResponse
	err = ftx.Request(req, &resp)
	if err != nil || !resp.Success {
		errors.Wrapf(err, "failed GetOpenOrders, resp.Success: %v", resp.Success)
		return orders, err
	}
	for _, v := range resp.Result {
		orders = append(orders, Order{
			Price:     v.Price,
			Status:    v.Status,
			Quantity:  v.Size,
			Timestamp: v.CreatedAt,
			Market:    v.Market,
			Side:      v.Side,
		})
	}
	return orders, nil
}
