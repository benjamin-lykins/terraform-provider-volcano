resource "volcano_storage_object" "logo" {
  project_id  = volcano_project.example.id
  bucket_name = volcano_storage_bucket.avatars.name
  path        = "branding/logo.png"
  source      = "${path.module}/assets/logo.png"
  is_public   = true
}
