# Copyright (c) HashiCorp, Inc.

data "contextforge_health" "example" {}

output "gateway_status" {
  value = data.contextforge_health.example.status
}
