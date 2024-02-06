# In order for the import to work, you should provide a ID composed of the following:
#  {ORG_ID}/{CONNECTED_APP_ID}

terraform import \
  -var-file params.tfvars.json \          #variables file
  anypoint_connected_app.client \            #resource name
  aa1f55d6-213d-4f60-845c-201282484cd1/rer123ze-213d-7f10-344c-909282484rr3    #resource ID
