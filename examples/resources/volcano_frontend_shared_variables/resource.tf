resource "volcano_frontend_shared_variables" "example" {
  project_id     = volcano_project.example.id
  variable_names = ["API_URL", "APP_NAME"]
}
