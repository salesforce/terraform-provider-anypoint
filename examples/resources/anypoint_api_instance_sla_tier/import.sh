# Import an existing SLA tier using the composite id: org_id/env_id/api_id/tier_id
#
# tier_id is the numeric id assigned by Anypoint (visible in the UI URL and API list response).
#
# Post-import behaviour:
#   After import, run `terraform plan`. If api_version_id was not set in config,
#   a planned change may appear because the platform returns it on list but it
#   differs from the default empty string. Set api_version_id = <api_id> in your
#   config to suppress this.
#   Second plan after apply should show "No changes."

terraform import anypoint_api_instance_sla_tier.gold <org_id>/<env_id>/<api_id>/<tier_id>
