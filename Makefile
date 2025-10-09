TEST?=$(shell go list ./...)

test:
	@go test -parallel 1 $(TEST)
