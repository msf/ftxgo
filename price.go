package ftxgo

import (
	"fmt"
	"net/http"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"
)

type OrderBook struct {
	Success bool `json:"success"`
	Result  struct {
		Asks [][2]float64 `json:"asks"`
		Bids [][2]float64 `json:"bids"`
	} `json:"result"`
}

func (ftx *FTXClient) GetPrice(market string) (price float64, err error) {
	ts := time.Now()
	defer func() {
		log.WithFields(log.Fields{
			"elapsed": time.Since(ts),
			"err":     err,
			"market":  market,
			"price":   price,
		}).Info(("GetPrice()"))
	}()

	url := fmt.Sprintf("https://ftx.com/api/markets/%v/orderbook?depth=16", market)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0.0, err
	}
	var orders OrderBook
	err = ftx.Request(req, &orders)
	if err != nil {
		return 0.0, err
	}
	if len(orders.Result.Asks) < 1 || len(orders.Result.Bids) < 1 {
		return 0.0, fmt.Errorf("empty order book? %+v", orders)
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
	log.Infof("asks: %v", asks[0])
	log.WithFields(log.Fields{
		"market": market,
		"gap":    fmt.Sprintf("%.3f", askPrice-bidPrice),
	}).Infof("bids: %v", bids[0])
	price = askPrice
	return
}
