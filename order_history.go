package ftxgo

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
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
	Success bool   `json:"success"`
	Error   string `json:"error"`
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
		RemainingSize float64   `json:"remainingSize"`
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
		log.WithFields(log.Fields{
			"elapsed": time.Since(ts),
			"err":     err,
			"market":  market,
			"orders":  len(orders),
		}).Info("GetOpenOrders()")
	}()

	url := fmt.Sprintf("https://ftx.com/api/orders?market=%v", market)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	var resp FTXOrdersResponse
	err = ftx.Request(req, &resp)
	if err != nil || !resp.Success {
		err = errors.Wrapf(err, "failed GetOpenOrders, resp.error: %v", resp.Error)
		log.WithFields(log.Fields{
			"err":  err,
			"resp": resp.Error,
			"url":  req.URL,
		}).Error("GetOpenOrder request failed")
		return
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
		log.WithField("order", fmt.Sprintf("%+v", v)).Info("open order found")
	}
	return orders, nil
}

func (ftx *FTXClient) GetClosedOrders(market, buyOrSell string, interval time.Duration) (orders []Order, err error) {
	ts := time.Now()
	defer func() {
		log.WithFields(log.Fields{
			"elapsed":  time.Since(ts),
			"err":      err,
			"interval": interval,
			"market":   market,
			"orders":   len(orders),
		}).Info("GetClosedBuyOrders()")
	}()

	startTime := time.Now().UTC().Add(-interval).Unix()
	v := url.Values{
		"start_time": []string{strconv.FormatInt(startTime, 10)},
		"market":     []string{market},
	}
	req, err := http.NewRequest("GET", "https://ftx.com/api/orders/history?"+v.Encode(), nil)
	if err != nil {
		return
	}
	var resp FTXOrdersResponse
	err = ftx.Request(req, &resp)
	if err != nil || !resp.Success {
		err = errors.Wrapf(err, "failed GetClosedBuyOrders, resp.Error: %v", resp.Error)
		log.WithFields(log.Fields{
			"url":  req.URL,
			"err":  err,
			"resp": resp.Error,
		}).Error("GetOpenOrder request failed")
		return
	}
	for _, v := range resp.Result {
		if v.Side != buyOrSell {
			continue
		}
		if v.FilledSize != v.Size {
			log.WithFields(log.Fields{
				"filledSize": v.FilledSize,
				"size":       v.Size,
				"order":      fmt.Sprintf("%+v", v),
			}).Info("Skipping not fully filled order")
			continue
		}

		order := Order{
			Price:     v.Price,
			Status:    v.Status,
			Quantity:  v.Size,
			Timestamp: v.CreatedAt,
			Market:    v.Market,
			Side:      v.Side,
		}
		orders = append(orders, order)
		log.WithFields(log.Fields{
			"order":       fmt.Sprintf("%+v", order),
			"original":    fmt.Sprintf("%+v", v),
			"isBuyOrSell": v.Side == buyOrSell,
		}).Debug("matching order found")
	}
	return
}
