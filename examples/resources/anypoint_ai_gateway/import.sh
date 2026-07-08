#!/bin/bash
# anypoint_ai_gateway is a composite over the api_instance's policies.
# Import is deferred to a follow-up release; for v1.11.0 manage the composite
# via terraform from the start.
#
# Composite id shape (for reference, once import lands):
#   {org_id}/{env_id}/{api_instance_id}/ai_gateway
#
# example (DOES NOT WORK in v1.11.0 — import handler not yet implemented):
#   terraform import anypoint_ai_gateway.prod aa1f55d6-.../7074fcdd-.../20935984/ai_gateway
