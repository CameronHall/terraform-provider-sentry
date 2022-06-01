default: testacc

.PHONY: deps
deps:
	go mod download

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 TF_LOG=debug go test ./... -v $(TESTARGS) -timeout 120m
