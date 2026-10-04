resource "volcano_email_template" "confirmation" {
  project_id    = volcano_project.example.id
  template_type = "confirmation"
  subject       = "Confirm your email"
  html_body     = "<p>Click <a href=\"{{ .ConfirmationURL }}\">here</a> to confirm.</p>"
  text_body     = "Confirm your email: {{ .ConfirmationURL }}"
}
