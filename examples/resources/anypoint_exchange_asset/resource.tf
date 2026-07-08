resource "anypoint_exchange_asset" "petstore_oas" {
  org_id     = var.org_id
  group_id   = var.org_id
  asset_id   = "petstore"
  version    = "1.0.0"
  name       = "Petstore"
  classifier = "oas"
  api_version = "v1"
  main_file  = "petstore.json"
  asset_file = "${path.module}/petstore.json"

  description = "Petstore OAS 3.0 specification"
  tags        = ["pets", "demo"]

  strict_package = false
}

resource "anypoint_exchange_asset" "external_http_api" {
  org_id     = var.org_id
  group_id   = var.org_id
  asset_id   = "weather-api"
  version    = "1.0.0"
  name       = "Weather API"
  classifier = "http"
  api_version = "v1"
  asset_link = "https://api.example.com/weather/openapi.json"

  description = "External weather API reference"
}
