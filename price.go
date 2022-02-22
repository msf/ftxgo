package ftxgo

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"time"
)

type OrderBook struct {
	Success bool `json:"success"`
	Result  struct {
		Asks [][2]float64 `json:"asks"`
		Bids [][2]float64 `json:"bids"`
	} `json:"result"`
}

func (ftx *FTXClient) GetPrice(market string) (float64, error) {
	ts := time.Now()
	defer func() {
		log.Printf("GetPrice() took: %v\n", time.Since(ts))
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
	log.Printf("%v bids: %v \n", market, bids[0])
	log.Printf("%v asks: %v\n", market, asks[0])
	log.Printf("%v gap: %.3f\n", market, askPrice-bidPrice)
	return askPrice, nil
}
