resource "volcano_project_auth_theme" "example" {
  project_id = volcano_project.example.id

  colors = {
    background  = "#ffffff"
    surface     = "#f5f5f5"
    text        = "#111111"
    accent      = "#5b21b6"
    accent_text = "#ffffff"
  }

  font    = "system"
  scale   = "default"
  density = "comfortable"
  radius  = "medium"
}
