# Terraform Provider for IBM MCP ContextForge

This Terraform provider enables infrastructure-as-code management for the [IBM MCP ContextForge](https://ibm.github.io/mcp-context-forge/) system.

> **Note**: To publish this provider to the Terraform Registry, you need to create and push at least one release tag. See [RELEASE_INSTRUCTIONS.md](RELEASE_INSTRUCTIONS.md) for details.

## About IBM MCP ContextForge

IBM MCP Context Forge is an open-source gateway and registry that centralizes the management of tools, resources, and prompts accessible to MCP-compatible LLM applications. It acts as a secure proxy and registry, federates both REST and MCP servers, and provides key enterprise features such as observability, authentication, rate limiting, and more.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the Go `install` command:

```shell
go install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

```hcl
terraform {
  required_providers {
    contextforge = {
      source  = "nkbud/contextforge"
      version = ">= 0.0.1"
    }
  }
}

provider "contextforge" {
  endpoint = "https://your-mcp-gateway.example.com"
}
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `make generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
make testacc
```
