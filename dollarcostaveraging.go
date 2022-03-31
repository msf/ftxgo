package ftxgo

import (
	"math"
	"time"

	log "github.com/sirupsen/logrus"
)

// confirm it is safe to buy order to continue strategy of Dollar Cost Averaging
func ConfirmDCAPlaceOrder(
	client *FTXClient, market string, budget float64, buyInterval, intervalCounts time.Duration,
) (bool, error) {
	orders, err := client.GetOpenBuyOrders(market)
	if err != nil {
		return false, err
	}
	if len(orders) > 0 {
		log.Printf("do NOT buy, found open buy orders: %+v", orders)
		return false, nil
	}
	// we get large buy window and check total spend, to protect against variance on buy interval
	pastBuys, err := client.GetClosedOrders(market, "buy", buyInterval*intervalCounts)
	if err != nil {
		return false, err
	}
	return IsBuyOrderOkay(pastBuys, budget, buyInterval, buyInterval*intervalCounts)
}

func IsBuyOrderOkay(pastBuys []Order, budget float64, interval time.Duration, timespan time.Duration) (bool, error) {

	total := 0.0
	for _, v := range pastBuys {
		// only include orders within 15% of budget
		if math.Abs(v.Spend()-budget) > (budget * 0.15) {
			log.Printf("found past order %+v, ignoring", v)
			continue
		}
		log.Printf("found past order %+v, considering it", v)
		total += v.Spend()
	}
	avgSpend := total / timespan.Hours() * 24
	budgetPerDay := budget / interval.Hours() * 24

	spendRatio := avgSpend / budgetPerDay

	shouldBuy := spendRatio < 0.98 // 3% below target is okay
	log.WithFields(log.Fields{
		"shouldBuy":    shouldBuy,
		"spendRatio":   RoundFloat(spendRatio),
		"avgSpend":     RoundFloat(avgSpend),
		"budgetPerDay": RoundFloat(budgetPerDay),
		"total":        total,
		"timespan":     timespan,
		"abortBuy":     !shouldBuy,
	}).Info("should buy?")
	return shouldBuy, nil
}

// Rounds a floa64 like FTX rounds market quantity volumes, to 3 decimal places
func RoundFloat(val float64) float64 {
	// round to 0.000
	return math.Round(val*1000) / 1000.0
}
