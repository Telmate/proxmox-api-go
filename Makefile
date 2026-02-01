SRC?=$(shell find cli proxmox internal main.go -name '*.go') go.mod go.sum

PHONY += all
all: build bash_completion

PHONY += build
build: proxmox-api-go

proxmox-api-go: $(SRC)
	go build -o proxmox-api-go

PHONY += bash_completion
bash_completion: proxmox-api-go-completion.bash

proxmox-api-go-completion.bash: proxmox-api-go
	NEW_CLI=true ./proxmox-api-go completion bash > $@

PHONY += install_bash_completion
install_bash_completion: $(HOME)/.local/share/bash-completion/completions/proxmox-api-go

$(HOME)/.local/share/bash-completion/completions/proxmox-api-go: proxmox-api-go-completion.bash
	@mkdir -p $(dir $@)
	cp $^ $@

UNIT_TEST_PATHS=./internal/... ./proxmox/...

PHONY += test
test: test-unit test-integration

PHONY += test_coverage
test_coverage:
	@go test -coverprofile=_coverage.out $(UNIT_TEST_PATHS) \
		&& go tool cover -html=_coverage.out -o _coverage.html

.PHONY: test-unit
test-unit: # Unit tests
	@go test -race -vet=off $(UNIT_TEST_PATHS)

.PHONY: test-integration
test-integration: # Integration tests
	@go test -parallel 1 ./test/...

.PHONY: test_integration_api
test_integration_api: # Integration
	@go test \
		./test/api/ApiToken/... \
		./test/api/Authentication/... \
		./test/api/Connection/... \
		./test/api/Group/... \
		./test/api/Pool/... \
		./test/api/Snapshot/... \
		./test/api/User/...

PHONY += clean
clean:
	rm -f proxmox-api-go proxmox-api-go-completion.bash

.PHONY: $(PHONY)
