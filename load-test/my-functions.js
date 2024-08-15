module.exports = { createSubscribeMessage,enableMessageLogging };


var pairs = generateCryptoPairs()

function createSubscribeMessage(userContext, events, done) {
  const data = {
    'type':'subscribe',
    'channel': getRandomPair()
  }
  userContext.vars.subscribe = data
  return done();
}

function getRandomPair() {
    const randomIndex = Math.floor(Math.random() * pairs.length);
    return pairs[randomIndex];
}

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
    "BTC", "ETH", "BNB", "ADA",
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

function enableMessageLogging(userContext, events, next) {
  userContext.ws.on(`message`, (message) => {
    events.emit('counter', 'websocket.messages_received', 1);
    events.emit('rate', 'websocket.receive_rate');
    calculateTimeDifference(message,events);
  });

  next();
}



function calculateTimeDifference(jsonString, events) {
  // console.error(jsonString);
  try {
      // Parse the JSON string
      const parsedData = JSON.parse(jsonString);

      // Extract the "E" field (timestamp)
      const eventTimestamp = parsedData.e;

      // if (typeof eventTimestamp !== 'number') {
      //     console.error('Invalid or missing "e" field in the JSON string');
      //     return;
      // }

      // Get the current time in milliseconds
      const currentTime = Date.now();
      const date = new Date(eventTimestamp);
      // Calculate the time difference
      const timeDifference = currentTime - date.getTime();

      events.emit('histogram','at_client_latency',timeDifference)
      // Print the time difference in a readable format
      // console.log(`Time difference: ${timeDifference} ms`);

  } catch (error) {
      console.error('Error parsing JSON string:', error.message);
  }
}
