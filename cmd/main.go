package main

import (
	"log"
	"os"
	"time"

	"github.com/msf/ftxgo"
	"github.com/namsral/flag"
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
	client := ftxgo.NewFTXClient(*apiKey, *secretKey)

	price, err := client.GetPrice(*marketTicker)
	if err != nil {
		log.Printf("Failed to getPrice: %v, aborting\n", err)
		os.Exit(1)
	}
	log.Printf("ETH price is: %v, %v", price, err)
	howMuch := calcQuantity(price, *budget)
	log.Printf("Placing BUY order: %.1f * %.6f = %v TOTAL\n", price, howMuch, price*howMuch)

	shouldBuy, err := ftxgo.ConfirmDCAPlaceOrder(client, *marketTicker, *budget, *buyInterval)
	if err == nil || !shouldBuy {
		log.Printf("Must stop, error (or shouldn't buy!) on validating DCA Place Order", err)
		os.Exit(1)
	}

	orderResult, err := client.PostBuyOrder(*marketTicker, price, howMuch)
	if err != nil || !orderResult.Success {
		log.Printf("Failed to placeBuyOrder(%v, %.1f, %.6f): %v, aborting\n", ETH_EUR, price, howMuch, err)
		os.Exit(1)
	}
	log.Printf("BUY order ETH: \n%+v\n", orderResult)
}
