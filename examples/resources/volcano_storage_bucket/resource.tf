resource "volcano_storage_bucket" "avatars" {
  project_id         = volcano_project.example.id
  name               = "avatars"
  file_size_limit    = 2097152
  allowed_mime_types = ["image/png", "image/jpeg"]
}
