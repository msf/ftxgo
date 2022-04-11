package main

import (
	"os"
	"time"

	"github.com/msf/ftxgo"
	"github.com/namsral/flag"
	log "github.com/sirupsen/logrus"
)

const ethToken = "stETH/USD"

func calcQuantity(price, budget float64) float64 {
	// budget = price * quantity
	raw := budget / price
	rounded := ftxgo.RoundFloat(raw)
	log.WithFields(log.Fields{
		"quantity-raw":     raw,
		"quantity-rounded": rounded,
	}).Info("CalcQuantity")
	return rounded
}

type UTCFormatter struct {
	log.Formatter
}

func (u UTCFormatter) Format(e *log.Entry) ([]byte, error) {
	e.Time = e.Time.UTC()
	return u.Formatter.Format(e)
}

func main() {
	apiKey := flag.String("ftx_api_key", "", "FTX API Key")
	secretKey := flag.String("ftx_secret_key", "", "FTX Secret Key")
	budget := flag.Float64("budget", 57.0, "Budget to buy Eth")
	buyInterval := flag.Duration("interval", 24*time.Hour, "Buy Interval")
	marketTicker := flag.String("market_ticker", ethToken, "Market Sticker name")
	executeBuy := flag.Bool("yes", false, "execute the buy order")
	avgWindow := flag.Int("avg_window", 7, "number of buy intervals to consider for the DCA")
	flag.Parse()

	log.SetFormatter(UTCFormatter{&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05.000Z",
		FullTimestamp:   true,
	}})
	log.WithFields(log.Fields{
		"budget":       *budget,
		"buyInterval":  *buyInterval,
		"marketTicker": *marketTicker,
		"executeBuy":   *executeBuy,
		"avgWindow":    *avgWindow,
	}).Info("Starting ftx DCA")

	client := ftxgo.NewFTXClient(*apiKey, *secretKey)

	price, err := client.GetPrice(*marketTicker)
	if err != nil {
		log.Printf("Failed to getPrice: %v, aborting\n", err)
		os.Exit(1)
	}
	howMuch := calcQuantity(price, *budget)

	shouldBuy, err := ftxgo.ConfirmDCAPlaceOrder(client, *marketTicker, *budget, *buyInterval, time.Duration(*avgWindow))
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
