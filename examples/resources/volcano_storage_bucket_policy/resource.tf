resource "volcano_storage_bucket_policy" "owner_only_select" {
  project_id  = volcano_project.example.id
  bucket_name = volcano_storage_bucket.avatars.name
  name        = "owner-only-select"
  operation   = "SELECT"
  definition  = "auth.uid() = owner_id"
}
