resource "volcano_frontend_domain" "web" {
  project_id      = volcano_project.example.id
  frontend_id     = volcano_frontend.web.id
  domain          = "app.example.com"
  certificate_pem = file("${path.module}/certs/app.example.com.pem")
  private_key_pem = file("${path.module}/certs/app.example.com.key")
}
