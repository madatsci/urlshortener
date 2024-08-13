.PHONY: build
build:
	cd cmd/shortener && go build -o shortener *.go

.PHONY: test_iter1
test_iter1:
	./shortenertestbeta -test.v -test.run=^TestIteration1$$ -binary-path=cmd/shortener/shortener -source-path=.

.PHONY: test_iter2
test_iter2:
	./shortenertestbeta -test.v -test.run=^TestIteration2$$ -binary-path=cmd/shortener/shortener -source-path=.

.PHONY: test_iter3
test_iter3:
	./shortenertestbeta -test.v -test.run=^TestIteration3$$ -binary-path=cmd/shortener/shortener -source-path=.

.PHONY: test_iter4
test_iter4:
	./shortenertestbeta -test.v -test.run=^TestIteration4$$ -binary-path=cmd/shortener/shortener -source-path=. -server-port=8081