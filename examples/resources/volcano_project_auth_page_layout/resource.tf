resource "volcano_project_auth_page_layout" "login" {
  project_id = volcano_project.example.id
  page_type  = "login"
  layout     = "split-left"
}
