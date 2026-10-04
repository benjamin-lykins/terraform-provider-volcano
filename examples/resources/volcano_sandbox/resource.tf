resource "volcano_sandbox" "dev" {
  project_id = volcano_project.example.id
  name       = "dev"
  preset     = "small"
}
