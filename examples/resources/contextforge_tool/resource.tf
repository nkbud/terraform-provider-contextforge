# Copyright (c) HashiCorp, Inc.

resource "contextforge_tool" "example" {
  name        = "my-tool"
  description = "A custom tool"
  visibility  = "private"

  input_schema = jsonencode({
    type = "object"
    properties = {
      query = {
        type        = "string"
        description = "Search query"
      }
    }
    required = ["query"]
  })

  tags = ["custom"]
}
