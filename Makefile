ITER_COUNT = 15

.PHONY: build
build:
	cd cmd/shortener && go build -o shortener *.go

.PHONY: lint
lint:
	golangci-lint run

.PHONY: run
run:
	./cmd/shortener/shortener

.PHONY: run_with_file
run_with_file:
	./cmd/shortener/shortener -f './tmp/storage.json'

.PHONY: run_with_db
run_with_db:
	./cmd/shortener/shortener -d 'postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable'

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

.PHONY: test_with_db
test_with_db:
	DATABASE_DSN=postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable go test -cover ./...

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
	./shortenertestbeta -test.v -test.run=^TestIteration9$$ -binary-path=cmd/shortener/shortener -source-path=. -file-storage-path=./tmp/storage.json

.PHONY: test_iter10
test_iter10:
	./shortenertestbeta -test.v -test.run=^TestIteration10$$ -binary-path=cmd/shortener/shortener -source-path=. -database-dsn='postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable'

.PHONY: test_iter11
test_iter11:
	./shortenertestbeta -test.v -test.run=^TestIteration11$$ -binary-path=cmd/shortener/shortener -source-path=. -database-dsn='postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable'

.PHONY: test_iter12
test_iter12:
	./shortenertestbeta -test.v -test.run=^TestIteration12$$ -binary-path=cmd/shortener/shortener -source-path=. -database-dsn='postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable'

.PHONY: test_iter13
test_iter13:
	./shortenertestbeta -test.v -test.run=^TestIteration13$$ -binary-path=cmd/shortener/shortener -source-path=. -database-dsn='postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable'

.PHONY: test_iter14
test_iter14:
	./shortenertestbeta -test.v -test.run=^TestIteration14$$ -binary-path=cmd/shortener/shortener -source-path=. -database-dsn='postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable'

.PHONY: test_iter15
test_iter15:
	./shortenertestbeta -test.v -test.run=^TestIteration15$$ -binary-path=cmd/shortener/shortener -source-path=. -database-dsn='postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable'

.PHONY: test_iter16
test_iter15:
	./shortenertestbeta -test.v -test.run=^TestIteration16$$ -binary-path=cmd/shortener/shortener -source-path=. -database-dsn='postgres://postgres:postgres@localhost:5432/praktikum?sslmode=disable'

.PHONY: base_profile_file
base_profile:
	curl -v "http://localhost:8080/debug/pprof/heap?seconds=40" > profiles/base.pprof

.PHONY: serve_base_profile_url
serve_base_profile_url:
	go tool pprof -http=":9090" -seconds=40 http://localhost:8080/debug/pprof/heap

.PHONY: serve_base_profile_file
serve_base_profile_file:
	go tool pprof -http=":9090" -seconds=40 profiles/base.pprof

.PHONY: compare_profiles
compare_profiles:
	go tool pprof -top -diff_base=profiles/base.pprof profiles/result.pprof