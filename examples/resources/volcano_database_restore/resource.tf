resource "volcano_database_restore" "rollback" {
  project_id    = volcano_project.example.id
  database_name = volcano_database.main.name
  backup_name   = volcano_database_backup.before_migration.name
}
