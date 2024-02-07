# In order for the import to work, you should provide a ID composed of the following:
#  {ORG_ID}/{ENV_ID}/{SG_ID}

terraform import \
  -var-file params.tfvars.json \    #variables file
  anypoint_secretgroup.sg \                #resource name
  aa1f55d6-213d-4f60-845c-201282484cd1/7074fcdd-9b23-4ab3-97c8-5db5f4adf17d/39731075-0521-47aa-82b2-d9745f2ac2eb   #resource ID
