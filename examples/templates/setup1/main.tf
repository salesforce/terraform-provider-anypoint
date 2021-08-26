terraform {
  required_providers {
    anypoint = {
      //versions = ["0.2"]
      source = "anypoint.mulesoft.com/automation/anypoint"
    }
  }
}

provider "anypoint" {
  username = var.username
  password = var.password
}