resource "volcano_database" "main" {
  project_id    = volcano_project.example.id
  name          = "main"
  region        = "us-east-1"
  pg_version    = "16"
  database_type = "volcano-db-xs"
}
