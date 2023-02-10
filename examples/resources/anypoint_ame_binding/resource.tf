resource "anypoint_amq" "amq_01" {
  org_id = var.root_org
  env_id = var.env_id
  region_id = "us-east-1"
  queue_id = "yourQueueID"
  fifo = false
  default_ttl = 604800000
  default_lock_ttl = 120000
  dead_letter_queue_id = "myEmptyDLQ"
  max_deliveries = 10
}

resource "anypoint_amq" "amq_02" {
  org_id = var.root_org
  env_id = var.env_id
  region_id = "us-east-1"
  queue_id = "yourQueueID"
  fifo = false
  default_ttl = 604800000
  default_lock_ttl = 120000
  dead_letter_queue_id = "myEmptyDLQ"
  max_deliveries = 10
}

resource "anypoint_amq" "amq_03" {
  org_id = var.root_org
  env_id = var.env_id
  region_id = "us-east-1"
  queue_id = "yourQueueID"
  fifo = false
  default_ttl = 604800000
  default_lock_ttl = 120000
  dead_letter_queue_id = "myEmptyDLQ"
  max_deliveries = 10
}

resource "anypoint_ame" "ame" {
  org_id = var.root_org
  env_id = var.env_id
  region_id = "us-east-1"
  exchange_id = "myExchangeId"
  encrypted = true
}


resource "anypoint_ame_binding" "ame_b_01" {
  org_id = var.root_org
  env_id = anypoint_amq.ame.env_id
  region_id = anypoint_amq.ame.region_id
  exchange_id = anypoint_ame.ame.exchange_id
  queue_id = anypoint_amq.amq_01.queue_id

  rule_str_compare {
    property_name = "my_property_name"
    property_type = "STRING"
    matcher_type = "EQ"
    value = "full"
  }
}

resource "anypoint_ame_binding" "ame_b_02" {
  org_id = var.root_org
  env_id = anypoint_amq.ame.env_id
  region_id = anypoint_amq.ame.region_id
  exchange_id = anypoint_ame.ame.exchange_id
  queue_id = anypoint_amq.amq_02.queue_id

  rule_str_state {
    property_name = "TO_ROUTE"
    property_type = "STRING"
    matcher_type = "EXISTS"
    value = true
  }
}

resource "anypoint_ame_binding" "ame_b_03" {
  org_id = var.root_org
  env_id = anypoint_amq.ame.env_id
  region_id = anypoint_amq.ame.region_id
  exchange_id = anypoint_ame.ame.exchange_id
  queue_id = anypoint_amq.amq_03.queue_id

  rule_str_set {
    property_name = "horse_name"
    property_type = "STRING"
    matcher_type = "ANY_OF"
    value = tolist(["sugar", "cash", "magic"])
  }

}

resource "anypoint_ame_binding" "ame_b_04" {
  org_id = var.root_org
  env_id = anypoint_amq.ame.env_id
  region_id = anypoint_amq.ame.region_id
  exchange_id = anypoint_ame.ame.exchange_id
  queue_id = anypoint_amq.amq_04.queue_id

  rule_num_compare {
    property_name = "nbr_horses"
    property_type = "NUMERIC"
    matcher_type = "GT"
    value = 12
  }

}

resource "anypoint_ame_binding" "ame_b_05" {
  org_id = var.root_org
  env_id = anypoint_amq.ame.env_id
  region_id = anypoint_amq.ame.region_id
  exchange_id = anypoint_ame.ame.exchange_id
  queue_id = anypoint_amq.amq_05.queue_id

  rule_num_state {
    property_name = "to_ship"
    property_type = "NUMERIC"
    matcher_type = "EXISTS"
    value = true
  }
}


resource "anypoint_ame_binding" "ame_b_06" {
  org_id = var.root_org
  env_id = anypoint_amq.ame.env_id
  region_id = anypoint_amq.ame.region_id
  exchange_id = anypoint_ame.ame.exchange_id
  queue_id = anypoint_amq.amq_06.queue_id

  rule_num_set {
    property_name = "nbr_horses"
    property_type = "NUMERIC"
    matcher_type = "RANGE"
    value = tolist([2,10])
  }
}