resource "volcano_project_auth_config" "example" {
  project_id                 = volcano_project.example.id
  enable_signup              = true
  enable_email_password      = true
  require_email_confirmation = true
  min_password_length        = 12
}
