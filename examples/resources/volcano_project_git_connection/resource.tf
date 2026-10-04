data "volcano_user_git_connections" "all" {}

resource "volcano_project_git_connection" "example" {
  project_id        = volcano_project.example.id
  connection_id     = data.volcano_user_git_connections.all.connections[0].id
  installation_id   = 123456
  repo_full_name    = "acme/app"
  root_directory    = "apps/web"
  production_branch = "main"
}
