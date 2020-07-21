TEST?=$(shell go list ./...)

test:
	@go test $(TEST)
