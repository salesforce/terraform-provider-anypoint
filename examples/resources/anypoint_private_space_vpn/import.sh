# Import an existing anypoint_private_space_vpn using a composite id:
#
#   {ORG_ID}/{PRIVATE_SPACE_ID}/{CONNECTION_ID}
#
# All three components are GUIDs.
#   - ORG_ID:           your Anypoint organization id.
#   - PRIVATE_SPACE_ID: found in the Anypoint UI under Runtime Manager > Private Spaces.
#   - CONNECTION_ID:    the GUID returned by the Anypoint Private Space "Connections" API,
#                       also visible in the URL when editing a VPN connection in the UI.
#
# Run `terraform import` to bring the connection under Terraform management:

terraform import \
  anypoint_private_space_vpn.my_vpn \
  aa1f55d6-213d-4f60-845c-201282484cd1/7f747999-a9bb-41d5-bfe8-d6b8cca68c62/4d2c0e91-2c11-4f5e-91a0-3a7c3b1f7a90

# Post-import behaviour
# ---------------------
# The first `terraform plan` after import may show diffs for write-only fields:
#
#   psk        - the API never returns PSKs. Set the real PSK values in your
#                configuration; the provider will hold them in state without
#                exposing them. Until you do, every plan shows
#                `~ psk = (sensitive value)`.
#
#   ptp_cidr   - the API does not return inside-tunnel CIDRs. Plan shows
#                `+ ptp_cidr = "169.254.x.x/30"`. The first `terraform apply`
#                is a safe state-fill — no tunnel restart. The second plan
#                shows No changes.
#
#   startup_action - the API returns a stale value. Plan may show
#                `+ startup_action = "add"` if your config differs from the
#                default "start". The first apply is a safe state-fill.
#                See the resource documentation for the known platform GET bug.
#
# In all cases, tunnels remain "available" through the first post-import apply.
# The second plan will show No changes for ptp_cidr and startup_action.
# Only psk requires manual configuration alignment to clear permanently.
