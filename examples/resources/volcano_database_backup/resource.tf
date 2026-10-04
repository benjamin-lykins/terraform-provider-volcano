resource "volcano_database_backup" "before_migration" {
  project_id    = volcano_project.example.id
  database_name = volcano_database.main.name
  name          = "before_migration"
}
