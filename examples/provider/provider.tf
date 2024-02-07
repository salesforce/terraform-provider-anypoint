provider "anypoint" {
  # use either username/pwd or client id/secret to connect to the platform

  username = var.username               # optionally use ANYPOINT_USERNAME env var
  password = var.password               # optionally use ANYPOINT_PASSWORD env var

  client_id = var.client_id             # optionally use ANYPOINT_CLIENT_ID env var
  client_secret = var.client_secret     # optionally use ANYPOINT_CLIENT_SECRET env var

  access_token  = var.access_token      # optionally use ANYPOINT_ACCESS_TOKEN env var

  # You may need to change the anypoint control plane: use 'eu' or 'us'
  # by default the control plane is 'us'
  cplane= var.cplane                    # optionnaly use ANYPOINT_CPLANE env var
}