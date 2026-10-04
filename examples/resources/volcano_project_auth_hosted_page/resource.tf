resource "volcano_project_auth_hosted_page" "login" {
  project_id = volcano_project.example.id
  page_type  = "login"
  html       = file("${path.module}/auth-pages/login.html")
  css        = file("${path.module}/auth-pages/login.css")
}
