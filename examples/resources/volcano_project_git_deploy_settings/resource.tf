resource "volcano_project_git_deploy_settings" "example" {
  project_id          = volcano_project.example.id
  auto_deploy_enabled = true
  deploy_functions    = true
  frontend_name       = volcano_frontend.web.name
  frontend_app_root   = "apps/web"
}
