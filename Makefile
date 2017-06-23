default: vet test

test:
	go test ./...

vet:
	go vet ./...

bench:
	go test ./... -run=NONE -bench=. -benchmem -benchtime=5s

# go get -u github.com/davelondon/rebecca/cmd/becca

doc: README.md client/README.md resp/README.md

README.md: README.md.tpl $(wildcard *.go)
	becca -package .

client/README.md: client/README.md.tpl $(wildcard client/*.go)
	cd client && becca -package .

resp/README.md: resp/README.md.tpl $(wildcard resp/*.go)
	cd resp && becca -package .
