.PHONY: build
build:
	cd cmd/shortener && go build -o shortener *.go

.PHONY: test
test:
	./shortenertest -test.v -test.run=^TestIteration$(iter)$$ -binary-path=cmd/shortener/shortener -source-path=.