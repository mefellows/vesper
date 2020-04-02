TEST?=./...

.DEFAULT_GOAL := ci

ci:: deps test goveralls

deps:
	@echo "--- 🐿  Fetching build dependencies "
	go get github.com/axw/gocov/gocov
	go get github.com/mattn/goveralls
	go get golang.org/x/tools/cmd/cover
	go get github.com/modocache/gover
	go get github.com/mitchellh/gox

goveralls:
	goveralls -service="travis-ci" -coverprofile=coverage.txt -repotoken $(COVERALLS_TOKEN)

release:
	echo "--- 🚀 Releasing it"
	"$(CURDIR)/scripts/release.sh"

test: deps
	@echo "--- ✅ Running tests"
	@if [ -f coverage.txt ]; then rm coverage.txt; fi;
	@echo "mode: count" > coverage.txt
	@for d in $$(go list ./... | grep -v vendor | grep -v examples); \
		do \
			go test -race -coverprofile=profile.out -covermode=atomic $$d; \
			if [ -f profile.out ]; then \
					cat profile.out | tail -n +2 >> coverage.txt; \
					rm profile.out; \
			fi; \
	done; \

	go tool cover -func coverage.txt


.PHONY: ci deps goveralls release test
