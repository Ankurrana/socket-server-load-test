package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

var (
	ctx           = context.Background()
	redisClient   *redis.Client
	upgrader      = websocket.Upgrader{}
	subscriptions = make(map[string][]*websocket.Conn) // Channel to client connections map
	mu            sync.Mutex                           // Mutex to handle concurrent access to subscriptions
)

type SubscribeMessage struct {
	Type    string `json:"type"`
	Channel string `json:"channel"`
}

// List of available cryptocurrency pairs
var cryptoPairs = generateChannelPairs()

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

func main() {
	// Initialize Redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Change this to your Redis address if needed
	})

	// Start WebSocket server
	http.HandleFunc("/", handleWebSocket)
	go func() {
		log.Println("WebSocket server started on :8080")
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			log.Fatal("WebSocket server error:", err)
		}
	}()

	// Subscribe to all crypto pairs channels
	for _, pair := range cryptoPairs {
		go subscribeToRedisChannel(pair)
	}

	// Block forever
	select {}
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true } // Allow connections from any origin

	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	defer conn.Close()

	// Handle subscription messages
	handleSubscriptions(conn)
}

func handleSubscriptions(conn *websocket.Conn) {
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("WebSocket read error:", err)
			removeConnection(conn)
			break
		}

		var subscribeMsg SubscribeMessage
		if err := json.Unmarshal(message, &subscribeMsg); err != nil {
			log.Println("Invalid JSON:", err)
			continue
		}

		// Handle subscribe message
		if subscribeMsg.Type == "subscribe" {
			channel := strings.ToUpper(subscribeMsg.Channel)
			if isValidCryptoPair(channel) {
				mu.Lock()
				subscriptions[channel] = append(subscriptions[channel], conn)
				mu.Unlock()
				log.Printf("Client subscribed to channel: %s", channel)
			} else {
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "Invalid channel"}`))
			}
		}
	}
}

func removeConnection(conn *websocket.Conn) {
	mu.Lock()
	defer mu.Unlock()

	// Remove the connection from all channels
	for channel, conns := range subscriptions {
		for i, c := range conns {
			if c == conn {
				// Remove the connection by slicing
				subscriptions[channel] = append(conns[:i], conns[i+1:]...)
				log.Printf("Removed connection from channel: %s", channel)
				break
			}
		}
	}
}

func subscribeToRedisChannel(channel string) {
	sub := redisClient.Subscribe(ctx, channel)
	ch := sub.Channel()

	for msg := range ch {
		mu.Lock()
		connections := subscriptions[channel]
		mu.Unlock()

		payload := extractJson(msg.Payload)

		// Broadcast to all active connections
		for i := 0; i < len(connections); i++ {
			conn := connections[i]

			if err := conn.WriteMessage(websocket.TextMessage, []byte(payload)); err != nil {
				log.Println("WebSocket write error, removing connection:", err)
				removeConnection(conn)
				i-- // Adjust index after removal
			}
		}
	}
}

func isValidCryptoPair(pair string) bool {
	for _, p := range cryptoPairs {
		if p == pair {
			return true
		}
	}
	return false
}

func extractJson(inputString string) string {
	// Split the string by the '/' character to get PART1 and the rest
	firstSplit := strings.SplitN(inputString, "/", 2)
	if len(firstSplit) < 2 {
		// Handle error if the split did not return two parts
		return ""
	}
	// part1 := firstSplit[0]
	restOfString := firstSplit[1]

	// Split the rest of the string by the '#' character to get PART2 and PART3
	secondSplit := strings.SplitN(restOfString, "#", 2)
	if len(secondSplit) < 1 {
		// Handle error if the split did not return two parts
		return ""
	}
	part2 := secondSplit[0]

	return part2
}
