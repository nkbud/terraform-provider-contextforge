# Copyright (c) HashiCorp, Inc.

data "contextforge_gateways" "all" {
  include_inactive = false
}
