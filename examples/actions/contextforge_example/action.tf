# Copyright (c) HashiCorp, Inc.

resource "terraform_data" "example" {
  input = "fake-string"

  lifecycle {
    action_trigger {
      events  = [before_create]
      actions = [action.contextforge_example.example]
    }
  }
}

action "contextforge_example" "example" {
  config {
    configurable_attribute = "some-value"
  }
}