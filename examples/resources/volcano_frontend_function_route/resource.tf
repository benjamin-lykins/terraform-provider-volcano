resource "volcano_frontend_function_route" "api" {
  project_id   = volcano_project.example.id
  frontend_id  = volcano_frontend.web.id
  function_id  = volcano_function.api.id
  path_prefix  = "/api"
  strip_prefix = true
}
