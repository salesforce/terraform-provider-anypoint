resource "anypoint_amq" "amq" {
  org_id = var.root_org
  env_id = var.env_id
  region_id = "us-east-1"
  queue_id = "YOUR_QUEUE_ID"
  fifo = false
  default_ttl = 604800000
  default_lock_ttl = 120000
  dead_letter_queue_id = "MY_EMPTY_DLQ"
  max_deliveries = 10
}