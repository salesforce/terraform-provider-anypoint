# In order for the import to work, you should provide a ID composed of the following:
#  {ORG_ID}/{USER_ID}

terraform import \
  -var-file params.tfvars.json \                                              #variables file
  anypoint_user.user \                                                        #resource name
  aa1f55d6-213d-4f60-845c-201282484cd1/00dc3850-1e83-4b3b-918b-86aa646e0daf   #resource ID
