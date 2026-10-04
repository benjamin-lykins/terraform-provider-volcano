resource "volcano_project_access_token" "ci" {
  project_id = volcano_project.example.id
  name       = "ci"
  scope      = "full"
}

output "ci_token" {
  value     = volcano_project_access_token.ci.token
  sensitive = true
}
