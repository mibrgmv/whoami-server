lint:
	golangci-lint run

gen:
	/bin/bash gen.sh

.PHONY: lint gen