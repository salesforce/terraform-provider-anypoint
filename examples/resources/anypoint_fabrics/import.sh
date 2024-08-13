# In order for the import to work, you should provide a ID composed of the following:
#  {ORG_ID}/{FABRICS_ID}

terraform import \
  -var-file params.tfvars.json \          #variables file
  anypoint_fabrics.rtf \            #resource name
  aa1f55d6-213d-4f60-845c-201282484cd1/4c641268-3917-45b0-acb8-f7cb0c0318ab    #resource ID
