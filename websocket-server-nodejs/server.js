const WebSocket = require('ws');
const redis = require('redis');

// Initialize Redis subscriber client
const redisSubscriber = redis.createClient({
    "url" : "redis://localhost:6379"
});

// List of available cryptocurrency pairs
const cryptoPairs = generateCryptoPairs()


function generateCryptoPairs() {
    //   const baseCurrencies = [
    //     "BTC", "ETH", "BNB", "ADA", "XRP", "SOL", "DOGE", "DOT", "LTC", "AVAX",
    //     "MATIC", "TRX", "SHIB", "ATOM", "LINK", "BCH", "XLM", "ALGO", "VET", "HBAR",
    //     "ICP", "FIL", "SAND", "MANA", "EGLD", "AXS", "FTT", "NEAR", "GRT", "LUNA",
    //     "CAKE", "XTZ", "KLAY", "MIOTA", "RUNE", "MKR", "ZEC", "ENJ", "ZIL", "BAT",
    //     "STX", "XMR", "WAVES", "CHZ", "HNT", "1INCH", "CRV", "CELO", "RAY", "RSR",
    //     "GALA", "YFI", "COMP", "LRC", "OMG", "QNT", "FTM", "BTT", "MIM", "USDC",
    //     "UNI", "AAVE", "SUSHI", "YGG", "REN", "BAL", "SRM", "ANKR", "SKL", "MTL",
    //     "REVV", "UOS", "POLS", "ALPHA", "LINA", "RGT", "ORN", "BNT", "MLN", "AKRO",
    //     "DIA", "OCEAN", "CVC", "VTHO", "SXP", "CTSI", "CHR", "STORJ", "PERP", "QUICK",
    //     "XEM", "MASK", "TOMO", "AVA", "FORTH", "REQ", "DGB", "LSK", "ANT", "BAND",
    // ];
    const baseCurrencies = [
        "BTC", "ETH", "BNB", "ADA"
    ];
  
    // const quoteCurrencies = [
    //     "USDT", "BTC", "ETH", "BUSD", "DAI", "USD", "EUR", "JPY", "GBP", "AUD",
    // ];
  
    const quoteCurrencies = [
      "USDT", "BTC"
    ];
  
    // Generate 200 crypto pairs
    let cryptoPairs = [];
    for (let base of baseCurrencies) {
        for (let quote of quoteCurrencies) {
            if (cryptoPairs.length < 200) {
                cryptoPairs.push(base + quote);
            }
        }
    }
    return cryptoPairs
  }


// Connect to Redis
redisSubscriber.on('error', (err) => console.error('Redis error:', err));
redisSubscriber.connect();

// Initialize WebSocket server
const wss = new WebSocket.Server({ port: 8080 });
console.log("WebSocket server is listening on port 8080");

// Mapping of channels to WebSocket clients
const channels = {};

// Handle WebSocket connections
wss.on('connection', (ws) => {
    ws.on('message', (message) => {
        try {
            const parsedMessage = JSON.parse(message);

            if (parsedMessage.type === 'subscribe' && parsedMessage.channel) {
                const channel = parsedMessage.channel.toUpperCase(); // Normalize channel name

                if (cryptoPairs.includes(channel)) {
                    // Subscribe the WebSocket client to the specified Redis channel
                    if (!channels[channel]) {
                        channels[channel] = [];
                    }
                    channels[channel].push(ws);

                    console.log(`Client subscribed to channel: ${channel}`);
                } else {
                    ws.send(JSON.stringify({ error: 'Invalid channel' }));
                }
            } else {
                ws.send(JSON.stringify({ error: 'Invalid message format' }));
            }
        } catch (err) {
            ws.send(JSON.stringify({ error: 'Invalid JSON' }));
        }
    });

    // Clean up on socket close
    ws.on('close', () => {
        for (const channel in channels) {
            channels[channel] = channels[channel].filter((client) => client !== ws);
            if (channels[channel].length === 0) {
                delete channels[channel];
            }
        }
    });
});

// Subscribe to all Redis channels
cryptoPairs.forEach((pair) => {
    redisSubscriber.subscribe(pair, (message) => {
        if (channels[pair]) {
            jsonStr = extractJson(message)
            channels[pair].forEach((client) => {
                if (client.readyState === WebSocket.OPEN) {
                    client.send(jsonStr);
                }
            });
        }
    });
});


function extractJson(inputString) {
    // Split the string by the '/' character to get PART1 and the rest
    const firstSplit = inputString.split('/', 2);
    const part1 = firstSplit[0];
    const restOfString = firstSplit[1];

    // Split the rest of the string by the '#' character to get PART2 and PART3
    const secondSplit = restOfString.split('#', 2);
    const part2 = secondSplit[0];
    // const part3 = secondSplit[1];

    return part2;
}