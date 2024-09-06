ITER_COUNT = 9

.PHONY: build
build:
	cd cmd/shortener && go build -o shortener *.go

.PHONY: run
run:
	./cmd/shortener/shortener

.PHONY: test
test:
	for (( n = 1; n <= $(ITER_COUNT); n++ )) ; do \
		make test_iter$$n; \
		if [ $$? -ne 0 ]; then \
			echo "Error on iteration $$n, exiting..."; \
			exit 1; \
		fi; \
	done
	echo "Tests completed."

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

.PHONY: test_iter5
test_iter5:
	./shortenertestbeta -test.v -test.run=^TestIteration5$$ -binary-path=cmd/shortener/shortener -source-path=. -server-port=8081

.PHONY: test_iter6
test_iter6:
	./shortenertestbeta -test.v -test.run=^TestIteration6$$ -binary-path=cmd/shortener/shortener -source-path=.

.PHONY: test_iter7
test_iter7:
	./shortenertestbeta -test.v -test.run=^TestIteration7$$ -binary-path=cmd/shortener/shortener -source-path=.

.PHONY: test_iter8
test_iter8:
	./shortenertestbeta -test.v -test.run=^TestIteration8$$ -binary-path=cmd/shortener/shortener -source-path=.

.PHONY: test_iter9
test_iter9:
	./shortenertestbeta -test.v -test.run=^TestIteration9$$ -binary-path=cmd/shortener/shortener -source-path=. -file-storage-path=./tmp/storage.txt