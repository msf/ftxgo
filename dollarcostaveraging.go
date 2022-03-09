package ftxgo

import (
	"math"
	"time"

	log "github.com/sirupsen/logrus"
)

// confirm it is safe to buy order to continue strategy of Dollar Cost Averaging
func ConfirmDCAPlaceOrder(client *FTXClient, market string, budget float64, buyInterval time.Duration) (bool, error) {
	orders, err := client.GetOpenBuyOrders(market)
	if err != nil {
		return false, err
	}
	if len(orders) > 0 {
		log.Printf("do NOT buy, found open buy orders: %+v", orders)
		return false, nil
	}
	// we get twice large buy window and check total spend, to protect against variance on buy interval
	const intervalCounts = 2
	pastBuys, err := client.GetClosedOrders(market, "buy", buyInterval*intervalCounts)
	if err != nil {
		return false, err
	}
	total := 0.0
	for _, v := range pastBuys {
		// only include orders within same budget
		if math.Abs(v.Spend()-budget) > 2.0 {
			log.Printf("found past order %+v, ignoring", v)
			continue
		}
		log.Printf("found past order %+v, considering it", v)
		total += v.Spend()
	}
	avgSpend := total / intervalCounts
	amountToBudget := budget - avgSpend
	log.WithFields(log.Fields{
		"total":          total,
		"budget":         budget,
		"timespan":       buyInterval * intervalCounts,
		"avgSpend":       avgSpend,
		"amountToBudget": amountToBudget,
	}).Info("should buy?")
	if amountToBudget < 1 {
		log.Printf("do NOT buy, found %v spent in last %v timespan, avg: %v, budget: %v", total, buyInterval*intervalCounts, avgSpend, budget)
		return false, nil
	}
	return true, nil
}
