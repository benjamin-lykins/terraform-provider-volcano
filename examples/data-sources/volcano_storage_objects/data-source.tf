data "volcano_storage_objects" "logos" {
  project_id = volcano_project.example.id
  search     = "logo"
}
