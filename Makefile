PROVIDER_VERSION := 0.1
BIN_PATH = local/registry.terraform.io/vegassor/clickhouse/${PROVIDER_VERSION}/linux_amd64/terraform-provider-clickhouse_v${PROVIDER_VERSION}
CH_DOCKER_DIR := tests/sqltests/fixtures/clickhouse

default: build

generate:
	go generate ./...

build:
	go build -o local/bin/terraform-provider-clickhouse .
	mkdir -p $$(dirname "${BIN_PATH}")
	cp local/bin/terraform-provider-clickhouse "${BIN_PATH}"

build-prod:
	echo Not implemented
	exit 1

init-config:
	PROVIDER_DIR=$(PWD) envsubst < .terraformrc.tmpl > .terraformrc
	ln -s $(PWD)/.terraformrc ~/.terraformrc

test:
	go test -count=1 -parallel=4 ./...

testacc:
	TF_ACC=1 go test -count=1 -parallel=4 -timeout 10m -v ./...

testsql: build
	cd tests && pytest -v -s --color=yes

chup:
	cd ${CH_DOCKER_DIR} && docker-compose up -d
chdn:
	cd ${CH_DOCKER_DIR} && docker-compose down
chrs:
	cd ${CH_DOCKER_DIR} && docker-compose down && docker-compose up -d
