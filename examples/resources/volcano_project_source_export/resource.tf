resource "volcano_project_source_export" "example" {
  project_id        = volcano_project.example.id
  production_branch = "main"
}
