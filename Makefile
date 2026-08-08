CONTEST ?= contest/sample

.PHONY: bundle submit

bundle:
	go run ./cmd/ac-bundle -in $(CONTEST)/main.go -out $(CONTEST)/submit.go
	@echo "生成: $(CONTEST)/submit.go"

submit: bundle
	@echo "=== 提出用コード ==="
	@cat $(CONTEST)/submit.go
