# In order for the import to work, you should provide a ID composed of the following:
#  {ORG_ID}

terraform import \
  -var-file params.tfvars.json \          #variables file
  anypoint_bg.bg \                        #resource name
  aa1f55d6-213d-4f60-845c-201282484cd1    #resource ID
