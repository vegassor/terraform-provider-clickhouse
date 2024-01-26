PROVIDER_VERSION = 0.1

default: install

generate:
	go generate ./...

install:
	go install .
	go build -o local/registry.terraform.io/vegassor/clickhouse/$(PROVIDER_VERSION)/linux_amd64/terraform-provider-clickhouse_v$(PROVIDER_VERSION) .


test:
	go test -count=1 -parallel=4 ./...

testacc:
	TF_ACC=1 go test -count=1 -parallel=4 -timeout 10m -v ./...
