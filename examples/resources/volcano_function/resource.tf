resource "volcano_function" "api" {
  project_id = volcano_project.example.id
  name       = "my-api-function"
  runtime    = "nodejs24.x"
  source     = "${path.module}/dist/my-api-function.zip"
  is_public  = true
}
