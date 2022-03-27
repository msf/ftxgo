package main

import (
	"math"
	"os"
	"time"

	"github.com/msf/ftxgo"
	"github.com/namsral/flag"
	log "github.com/sirupsen/logrus"
)

const ETH_EUR = "ETH/EUR"

func calcQuantity(price, budget float64) float64 {
	// budget = price * quantity
	raw := budget / price
	// round to 0.0000
	rounded := math.Round(raw*10000) / 10000.0
	log.WithFields(log.Fields{
		"quantity-raw":     raw,
		"quantity-rounded": rounded,
	}).Info("CalcQuantity")
	return rounded
}

func main() {
	apiKey := flag.String("ftx_api_key", "", "FTX API Key")
	secretKey := flag.String("ftx_secret_key", "", "FTX Secret Key")
	budget := flag.Float64("budget", 51.0, "Budget in Euros to buy Eth")
	buyInterval := flag.Duration("interval", 24*time.Hour, "Buy Interval")
	marketTicker := flag.String("market_ticker", ETH_EUR, "Market Sticker name")
	executeBuy := flag.Bool("yes", false, "execute the buy order")
	flag.Parse()

	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 03:04:05.000",
		FullTimestamp:   true,
	})
	log.WithFields(log.Fields{
		"budget":       *budget,
		"buyInterval":  *buyInterval,
		"marketTicker": *marketTicker,
		"executeBuy":   *executeBuy,
	}).Info("Starting ftx DCA")

	client := ftxgo.NewFTXClient(*apiKey, *secretKey)

	price, err := client.GetPrice(*marketTicker)
	if err != nil {
		log.Printf("Failed to getPrice: %v, aborting\n", err)
		os.Exit(1)
	}
	howMuch := calcQuantity(price, *budget)

	shouldBuy, err := ftxgo.ConfirmDCAPlaceOrder(client, *marketTicker, *budget, *buyInterval)
	if err != nil || !shouldBuy {
		log.WithFields(log.Fields{
			"err":       err,
			"shouldBuy": shouldBuy,
			"market":    *marketTicker,
			"howMuch":   howMuch,
			"budget":    *budget,
		}).Warn("Must Stop. failed valdation for DCA Place Order")
		os.Exit(1)
	}

	if !*executeBuy {
		log.WithFields(log.Fields{
			"symbol":  *marketTicker,
			"price":   price,
			"howMuch": howMuch,
			"budget":  *budget,
		}).Warn("Must Stop. missing '-yes' flag")
		os.Exit(0)
	}
	log.WithFields(log.Fields{
		"symbol":  *marketTicker,
		"price":   price,
		"howMuch": howMuch,
		"budget":  *budget,
	}).Info("Go-Ahead to BUY")
	orderResult, err := client.PostBuyOrder(*marketTicker, price, howMuch, *executeBuy)
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
