# Copyright (c) HashiCorp, Inc.

resource "contextforge_gateway" "example" {
  name        = "atlassian"
  url         = "https://mcp.atlassian.com/v1/mcp"
  description = "Atlassian MCP Gateway"
  transport   = "STREAMABLEHTTP"
  is_active   = true
  tags        = ["atlassian"]

  health_check_url      = "https://mcp.atlassian.com/v1/mcp/health"
  health_check_interval = 30
  health_check_timeout  = 10
  health_check_retries  = 3

  passthrough_headers = ["Authorization"]
}
