# Copyright (c) HashiCorp, Inc.

data "contextforge_prompts" "all" {
  include_inactive = false
}
