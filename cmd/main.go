package main

import (
	"os"
	"time"

	"github.com/msf/ftxgo"
	"github.com/namsral/flag"
	log "github.com/sirupsen/logrus"
)

const ETH_EUR = "ETH/EUR"

func calcQuantity(price, budget float64) float64 {
	// budget = price * quantity
	return budget / price
}

func main() {
	apiKey := flag.String("ftx_api_key", "", "FTX API Key")
	secretKey := flag.String("ftx_secret_key", "", "FTX Secret Key")
	budget := flag.Float64("budget", 50.0, "Budget in Euros to buy Eth")
	buyInterval := flag.Duration("interval", 7*24*time.Hour, "Buy Interval")
	marketTicker := flag.String("market_ticker", ETH_EUR, "Market Sticker name")
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 03:04:05.000",
		FullTimestamp:   true,
	})

	client := ftxgo.NewFTXClient(*apiKey, *secretKey)

	price, err := client.GetPrice(*marketTicker)
	if err != nil {
		log.Printf("Failed to getPrice: %v, aborting\n", err)
		os.Exit(1)
	}
	log.Printf("%v price is: %v, %v", *marketTicker, price, err)
	howMuch := calcQuantity(price, *budget)

	shouldBuy, err := ftxgo.ConfirmDCAPlaceOrder(client, *marketTicker, *budget, *buyInterval)
	if err != nil || !shouldBuy {
		log.WithFields(log.Fields{
			"err":       err,
			"shouldBuy": shouldBuy,
			"market":    *marketTicker,
			"budget":    *budget,
		}).Warn("Must Stop. failed valdation for DCA Place Order")
		os.Exit(1)
	}

	orderResult, err := client.PostBuyOrder(*marketTicker, price, howMuch)
	if err != nil || !orderResult.Success {
		log.WithFields(log.Fields{
			"err":     err,
			"Success": orderResult.Success,
			"Error":   orderResult.Error,
		}).Warn("Must Stop. failed  Place Order")
		os.Exit(1)
	}
	log.Infof("BOUGHT %v: %+v", *marketTicker, orderResult)
}
