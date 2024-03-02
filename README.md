# Terraform Provider ClickHouse

Once you've written your provider, you'll want to [publish it on the Terraform Registry](https://developer.hashicorp.com/terraform/registry/providers/publishing) so that others can use it.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21

## TODO
- [ ] Provider
  - [x] Add HTTP (port 8123) configuration support
- [ ] Resources
  - [ ] Handle missing resources: do not fail if a resource does not exist, but set empty state
  - [x] Role resource
  - [x] Grant privilege resource
    - [ ] Check if grants break the resource and cause re-creation on every plan
    - [ ] Implement partial revoke support
  - [x] Grant role resource
  - [ ] Dictionary resource
  - [x] View resource
  - [ ] MatView resource
  - [ ] Complex table resources
    - [ ] MergeTree family
    - [ ] RabbitMQ table
  - [ ] Implement import
  - [x] Table
    - [x] Add support for `settings` block
    - [x] Add `full_name` output as a computed field, equal to `db_name.table_name`
- [ ] Add datasources
- [ ] Tests
  - [ ] Acceptance tests
  - [x] Test more ClickHouse versions
  - [ ] Run SQL tests in parallel
- [ ] Release
  - [ ] Configure GitHub Actions to publish the provider to the Terraform Registry
