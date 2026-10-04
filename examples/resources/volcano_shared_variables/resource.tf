resource "volcano_shared_variables" "example" {
  project_id     = volcano_project.example.id
  variable_names = ["API_KEY", "DEBUG"]
}
