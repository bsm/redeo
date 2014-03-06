default: test

test:
	@go test

bench:
	@go test --bench=.

.PHONY: test

deps:
	@go get


