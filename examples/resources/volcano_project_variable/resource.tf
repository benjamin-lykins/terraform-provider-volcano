resource "volcano_project_variable" "api_key" {
  project_id = volcano_project.example.id
  name       = "THIRD_PARTY_API_KEY"
  value      = var.third_party_api_key
  shared     = false
}
