resource "volcano_project_auth_methods" "example" {
  project_id            = volcano_project.example.id
  enable_email_password = true
  enable_anonymous      = false

  oauth_providers = [
    { provider = "google", enabled = true },
    { provider = "github", enabled = false },
  ]
}
