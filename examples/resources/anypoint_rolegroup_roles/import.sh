# In order for the import to work, you should provide a ID composed of the following:
#  {ORG_ID}/{ROLE_GROUP_ID}

terraform import \
  -var-file params.tfvars.json \                  #variables file
  anypoint_rolegroup_roles.rg_roles \             #resource name
  aa1f55d6-213d-4f60-845c-201282484cd1/de32fc9d-6b25-4d6f-bd5e-cac32272b2f7    #resource ID
