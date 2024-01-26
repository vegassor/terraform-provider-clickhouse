PROVIDER_VERSION = 0.1

default: build

generate:
	go generate ./...

build:
	go build -o local/registry.terraform.io/vegassor/clickhouse/$(PROVIDER_VERSION)/linux_amd64/terraform-provider-clickhouse_v$(PROVIDER_VERSION) .

build-prod:
	echo Not implemented
	exit 1

init-config:
	ln -s $(PWD)/.terraformrc ~/.terraformrc

test:
	go test -count=1 -parallel=4 ./...

testacc:
	TF_ACC=1 go test -count=1 -parallel=4 -timeout 10m -v ./...
