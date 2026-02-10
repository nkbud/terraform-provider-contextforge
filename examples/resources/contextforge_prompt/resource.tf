# Copyright (c) HashiCorp, Inc.

resource "contextforge_prompt" "example" {
  name        = "summarize"
  description = "Summarize a document"
  visibility  = "public"

  arguments = jsonencode([
    {
      name        = "text"
      description = "The text to summarize"
      required    = true
    },
    {
      name        = "max_length"
      description = "Maximum summary length"
      required    = false
    }
  ])

  tags = ["nlp"]
}
