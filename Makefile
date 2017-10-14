PKG=$(shell go list ./... | grep -v vendor)

default: vet test

test:
	go test $(PKG)

vet:
	go vet $(PKG)

bench:
	go test $(PKG) -run=NONE -bench=. -benchmem -benchtime=5s

fuzz:
	go test ./.fuzz

fuzzrace:
	go test -race ./.fuzz

# go get -u github.com/davelondon/rebecca/cmd/becca

doc: README.md client/README.md resp/README.md

README.md: README.md.tpl $(wildcard *.go)
	becca -package .

client/README.md: client/README.md.tpl $(wildcard client/*.go)
	cd client && becca -package .

resp/README.md: resp/README.md.tpl $(wildcard resp/*.go)
	cd resp && becca -package .
