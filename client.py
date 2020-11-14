# Simple python websocket client
# https://github.com/websocket-client/websocket-client
from websocket import create_connection
options = {}
options['origin'] = 'https://exchange.blockchain.com'
url = "wss://ws.prod.blockchain.info/mercury-gateway/v1/ws"
ws = create_connection(url, **options)
msg = '{"action": "subscribe", "channel": "l2", "symbol": "BTC-USD"}'
ws.send(msg)
result =  ws.recv()
print(result)
ws.close()

