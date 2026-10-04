resource "volcano_frontend" "web" {
  project_id = volcano_project.example.id
  name       = "web"
  source     = "${path.module}/dist/web.zip"
  framework  = "nextjs"
}
