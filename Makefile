.PHONY: test test-integration vet build-examples examples \
        example-basic example-websocket example-order example-candles

test:
	go test ./rofex/...

test-integration:
	go test -v ./rofex/...

vet:
	go vet ./...

build-examples:
	go build ./examples/...

examples: example-basic example-order example-candles example-websocket

example-basic:
	go run ./examples/basic/

example-websocket:
	go run ./examples/websocket/

example-order:
	go run ./examples/order/

example-candles:
	go run ./examples/candles/
