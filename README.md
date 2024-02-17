# Terraform Provider ClickHouse

Once you've written your provider, you'll want to [publish it on the Terraform Registry](https://developer.hashicorp.com/terraform/registry/providers/publishing) so that others can use it.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21

## TODO
- [ ] Provider
  - [ ] Add HTTP (port 8123) configuration support
- [ ] Resources
  - [x] Role resource
  - [x] Grant privilege resource
    - [ ] Check if grants break the resource and cause re-creation on every plan
    - [ ] Implement partial revoke support
  - [x] Grant role resource
  - [ ] Dictionary resource
  - [ ] View resource
  - [ ] MatView resource
  - [ ] Complex table resources
    - [ ] MergeTree family
    - [ ] RabbitMQ table
  - [ ] Implement import
- [ ] Add datasources
- [ ] Tests
  - [ ] Acceptance tests
  - [ ] Test more ClickHouse versions
  - [ ] Run SQL tests in parallel
- [ ] Release
  - [ ] Configure GitHub Actions to publish the provider to the Terraform Registry
