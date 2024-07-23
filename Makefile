.PHONY: test
test:
	./shortenertest -test.v -test.run=^TestIteration$(iter)$$ -binary-path=cmd/shortener/shortener -source-path=.