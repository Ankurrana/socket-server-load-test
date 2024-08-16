package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"encoding/json"

	"github.com/go-redis/redis/v8"
)

/*
	This script produces "Workers" number of workers.
	Each worker pushes 1 redis message to a random channel in "Interval" duration.
	channel names are generated randomly using the generateChannelPairs method below

	Each message is of the form
	order-book/2024-08-14T14:37:01+05:30#{"ts":1723626421306,"vs":19159414,"asks":{"65037.75":"44.61253","65120.68":"11.18466","65142.67":"33.30210","65145.42"}
*/

/* configuration */
const (
	workers             int           = 8
	scriptTotalDuration time.Duration = time.Hour
	interval            time.Duration = (time.Millisecond * 1000) / 25
	redisAddress        string        = "localhost:6379"
)

type OrderBook struct {
	Ts   int64             `json:"ts"`
	Vs   int               `json:"vs"`
	Asks map[string]string `json:"asks"`
	Bids map[string]string `json:"bids"`
	Pr   string            `json:"pr"`
	S    string            `json:"s"`
	E    int64             `json:"e"`
}

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddress, // Redis server address
	})

	for _ = range workers {
		go publishOrdersToRedis(rdb)
	}

	time.Sleep(scriptTotalDuration)

}

func publishOrdersToRedis(rdb *redis.Client) {
	channels := generateChannelPairs()

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	ctx := context.Background()
	for {
		ctime := time.Now().UnixMilli()

		// fmt.Println("time:", ctime)
		// Generate random OrderBook data
		orderBook := OrderBook{
			Ts:   ctime,
			Vs:   rand.Intn(100000000),
			Asks: generateRandomOrders(),
			Bids: generateRandomOrders(),
			Pr:   "spot",
			S:    "BTCUSDT",
			E:    ctime,
		}

		// Convert OrderBook to JSON
		orderBookJsonFirst, err1 := json.Marshal(orderBook)
		orderBookJson := string(orderBookJsonFirst)
		if err1 != nil {
			fmt.Printf("Error marshaling JSON: %v , %v\n", err1, err1)
			continue
		}

		// Format the output string
		output := fmt.Sprintf("order-book/%s#%s", orderBookJson, "RANDOMLONGLONGLONGLONGSTRINGSTRING")

		ch := channels[rand.Intn(len(channels))]
		payload := output

		err := rdb.Publish(ctx, ch, payload).Err()
		if err != nil {
			log.Printf("failed to publish message: %v", err)
		}
		// else {
		// 	// fmt.Printf("message published to channel:%s and payload:%s\n", ch, payload)
		// }

		// Sleep for a short duration to avoid flooding
		time.Sleep(interval) // Adjust sleep duration as needed
	}
}

// generateRandomOrders is a helper function to generate random orders
func generateRandomOrders() map[string]string {
	orders := make(map[string]string)
	numOrders := rand.Intn(50) + 1 // Random number of orders between 1 and 7
	for i := 0; i < numOrders; i++ {
		price := fmt.Sprintf("%.2f", rand.Float64()*1000+65000)
		quantity := fmt.Sprintf("%.5f", rand.Float64()*100)
		orders[price] = quantity
	}
	return orders
}

/* generateChannelPairs will generate symbol pairs as channel names for redis publisher   */
func generateChannelPairs() []string {
	// baseCurrencies := []string{
	// 	"BTC", "ETH", "BNB", "ADA", "XRP", "SOL", "DOGE", "DOT", "LTC", "AVAX",
	// 	"MATIC", "TRX", "SHIB", "ATOM", "LINK", "BCH", "XLM", "ALGO", "VET", "HBAR",
	// 	"ICP", "FIL", "SAND", "MANA", "EGLD", "AXS", "FTT", "NEAR", "GRT", "LUNA",
	// 	"CAKE", "XTZ", "KLAY", "MIOTA", "RUNE", "MKR", "ZEC", "ENJ", "ZIL", "BAT",
	// 	"STX", "XMR", "WAVES", "CHZ", "HNT", "1INCH", "CRV", "CELO", "RAY", "RSR",
	// 	"GALA", "YFI", "COMP", "LRC", "OMG", "QNT", "FTM", "BTT", "MIM", "USDC",
	// 	"UNI", "AAVE", "SUSHI", "YGG", "REN", "BAL", "SRM", "ANKR", "SKL", "MTL",
	// 	"REVV", "UOS", "POLS", "ALPHA", "LINA", "RGT", "ORN", "BNT", "MLN", "AKRO",
	// 	"DIA", "OCEAN", "CVC", "VTHO", "SXP", "CTSI", "CHR", "STORJ", "PERP", "QUICK",
	// 	"XEM", "MASK", "TOMO", "AVA", "FORTH", "REQ", "DGB", "LSK", "ANT", "BAND",
	// }
	// quoteCurrencies := []string{
	// 		"USDT", "BTC", "ETH", "BUSD", "DAI", "USD", "EUR", "JPY", "GBP", "AUD",
	// 	}

	baseCurrencies := []string{
		"BTC", "ETH", "BNB", "ADA",
	}

	quoteCurrencies := []string{
		"USDT", "BTC",
	}

	//

	// Generate 200 crypto pairs
	var cryptoPairs []string
	for _, base := range baseCurrencies {
		for _, quote := range quoteCurrencies {
			cryptoPairs = append(cryptoPairs, base+quote)
		}
	}

	return cryptoPairs
}
