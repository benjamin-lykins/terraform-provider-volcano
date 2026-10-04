data "volcano_user_git_connections" "all" {}

data "volcano_user_git_installations" "example" {
  connection_id = data.volcano_user_git_connections.all.connections[0].id
}
