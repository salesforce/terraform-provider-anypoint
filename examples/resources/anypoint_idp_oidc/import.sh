# In order for the import to work, you should provide a ID composed of the following:
#  {ORG_ID}/{IDP_ID}

terraform import \
  -var-file params.tfvars.json \          #variables file
  anypoint_idp_oidc.example1 \            #resource name
  aa1f55d6-213d-4f60-845c-201282484cd1/452a2081-5bde-4fb9-9a8b-54d180ee2358    #resource ID
