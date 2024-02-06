# In order for the import to work, you should provide a ID composed of the following:
#  {ORG_ID}/{ENV_ID}/{API_ID}/{API_POLICY_ID}

terraform import \
  -var-file params.tfvars.json \    #variables file
  anypoint_apim_policy_rate_limiting.policy01 \                #resource name
  aa1f55d6-213d-4f60-845c-207286484cd1/7074fcdd-9b23-4ab3-97c8-5db5f4adf17d/19250669/4720771      #resource ID
