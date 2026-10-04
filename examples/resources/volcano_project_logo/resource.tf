resource "volcano_project_logo" "example" {
  project_id = volcano_project.example.id
  source     = "${path.module}/logo.png"
}
