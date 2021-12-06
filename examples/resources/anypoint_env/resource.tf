resource "anypoint_env" "env" {
  org_id = anypoint_bg.bg.id    # environment related business group
  name = "DEV"                  # environment name
  type = "sandbox"              # environment type : sandbox/production
}
