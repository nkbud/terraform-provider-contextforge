resource "contextforge_server" "example" {
  name        = "fast-time"
  description = "Demo server"
  tags        = ["demo"]
  visibility  = "private"
}
