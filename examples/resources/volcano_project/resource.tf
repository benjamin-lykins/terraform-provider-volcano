resource "volcano_project" "example" {
  name = "my-awesome-app"
}

resource "volcano_project" "pinned_regions" {
  name             = "eu-only-app"
  all_regions      = false
  selected_regions = ["eu-west-1"]
}
