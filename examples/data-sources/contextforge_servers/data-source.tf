# Copyright (c) HashiCorp, Inc.

data "contextforge_servers" "example" {
  include_inactive = false
}

output "server_count" {
  value = length(data.contextforge_servers.example.servers)
}
