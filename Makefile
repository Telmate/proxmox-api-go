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

PHONY += test
test: test-unit test-integration

test-unit: # Unit tests
	@go test -race -vet=off ./internal/... ./proxmox/...

test-integration: # Integration tests
	@go test -parallel 1 ./test/...

PHONY += clean
clean:
	rm -f proxmox-api-go proxmox-api-go-completion.bash

.PHONY: $(PHONY)
