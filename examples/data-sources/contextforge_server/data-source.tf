data "contextforge_server" "example" {
  id = "srv-id"
}

output "server_name" {
  value = data.contextforge_server.example.name
}
