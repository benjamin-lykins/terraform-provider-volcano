resource "volcano_anon_key" "frontend" {
  project_id  = volcano_project.example.id
  name        = "frontend-app"
  permissions = ["auth.signup", "auth.signin", "auth.refresh", "auth.logout"]
  is_default  = true
}
