PROVIDER_VERSION = 0.1

default: build

generate:
	go generate ./...

build:
	go build -o local/bin/terraform-provider-clickhouse .
	cp local/bin/terraform-provider-clickhouse local/registry.terraform.io/vegassor/clickhouse/$(PROVIDER_VERSION)/linux_amd64/terraform-provider-clickhouse_v$(PROVIDER_VERSION)

build-prod:
	echo Not implemented
	exit 1

init-config:
	ln -s $(PWD)/.terraformrc ~/.terraformrc

test:
	go test -count=1 -parallel=4 ./...

testacc:
	TF_ACC=1 go test -count=1 -parallel=4 -timeout 10m -v ./...

testsql:
	cd tests && pytest -v -s

chup:
	cd clickhouse && docker-compose up -d
chdn:
	cd clickhouse && docker-compose down
chrs:
	cd clickhouse && docker-compose down && docker-compose up -d
