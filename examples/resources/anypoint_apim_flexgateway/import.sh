# In order for the import to work, you should provide a ID composed of the following:
#  {ORG_ID}/{ENV_ID}/{API_ENV_ID}

terraform import \
  -var-file params.tfvars.json \    #variables file
  anypoint_apim_flexgateway.fg \                #resource name
  aa1f55d6-213d-4f60-845c-201282484cd1/7074fcdd-9b23-4ab6-97r8-5db5f4adf17d/19218070    #resource ID
