resource "volcano_oauth_config" "google" {
  project_id     = volcano_project.example.id
  oauth_provider = "google"
  client_id      = var.google_oauth_client_id
  client_secret  = var.google_oauth_client_secret
  scopes         = ["email", "profile"]
  enabled        = true
}
