default: test

.minimal.makefile:
	curl -fsSL -o $@ https://gitlab.com/bsm/misc/raw/master/make/go/minimal.makefile

include .minimal.makefile

# go get -u github.com/davelondon/rebecca/cmd/becca
README.md: README.md.tpl
	becca -package github.com/bsm/redeo/v2
