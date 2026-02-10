# Copyright (c) HashiCorp, Inc.

provider "contextforge" {
  # Configuration for ContextForge MCP Gateway
  endpoint     = "https://your-mcp-gateway.example.com"
  bearer_token = var.mcpgateway_bearer_token
}
