package ftxgo

import (
	"log"
	"math"
	"time"
)

// confirm it is safe to buy order to continue strategy of Dollar Cost Averaging
func ConfirmDCAPlaceOrder(client *FTXClient, market string, budget float64, buyInterval time.Duration) (bool, error) {
	orders, err := client.GetOpenBuyOrders(market)
	if err != nil {
		return false, err
	}
	if len(orders) > 0 {
		log.Print("do NOT buy, found open buy orders: %+v", orders)
		return false, nil
	}
	// we get twice large buy window and check total spend, to protect against variance on buy interval
	pastBuys, err := client.GetClosedBuyOrders(market, buyInterval*2)
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
		total += v.Price
	}
	if math.Abs(total/2-budget) > 1 {
		log.Printf("do NOT buy, found %v spent in last %v timespan", total, buyInterval*2)
		return false, nil
	}
	return true, nil
}
