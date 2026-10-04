data "volcano_projects" "all" {}

data "volcano_projects" "filtered" {
  search = "staging"
}
