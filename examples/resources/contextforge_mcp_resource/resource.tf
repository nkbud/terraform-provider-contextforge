# Copyright (c) HashiCorp, Inc.

resource "contextforge_mcp_resource" "example" {
  uri         = "file:///data/config.json"
  name        = "config"
  description = "Application configuration"
  mime_type   = "application/json"
  visibility  = "private"
  tags        = ["config"]
}
