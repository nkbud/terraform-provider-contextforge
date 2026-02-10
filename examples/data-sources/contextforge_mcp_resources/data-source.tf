# Copyright (c) HashiCorp, Inc.

data "contextforge_mcp_resources" "all" {
  include_inactive = false
}
