data "volcano_frontend_deployments" "web" {
  project_id  = volcano_project.example.id
  frontend_id = volcano_frontend.web.id
}
