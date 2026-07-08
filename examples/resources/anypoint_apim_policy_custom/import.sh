# In order for the import to work, you should provide a composite ID in one of two shapes:
#
#  Inbound policies (default — backwards-compatible 4-segment shape):
#    {ORG_ID}/{ENV_ID}/{API_ID}/{API_POLICY_ID}
#
#  Outbound policies (5-segment shape, required when injection_point = "outbound"):
#    {ORG_ID}/{ENV_ID}/{API_ID}/outbound/{API_POLICY_ID}
#
# The injection_point segment must be either "inbound" or "outbound" and routes the
# Read call to the matching endpoint family (.../policies vs .../policies/outbound-policies).

# Inbound example
terraform import \
  -var-file params.tfvars.json \
  anypoint_apim_policy_custom.policy_custom_01 \
  aa1f55d6-213d-4f60-845c-207286484cd1/7074fcdd-9b23-4ab3-97c8-5db5f4adf17d/19250669/4720771

# Outbound example
terraform import \
  -var-file params.tfvars.json \
  anypoint_apim_policy_custom.policy_custom_06_outbound_obo \
  aa1f55d6-213d-4f60-845c-207286484cd1/7074fcdd-9b23-4ab3-97c8-5db5f4adf17d/19250669/outbound/4720772
