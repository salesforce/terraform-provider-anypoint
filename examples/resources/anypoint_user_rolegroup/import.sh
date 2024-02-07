# In order for the import to work, you should provide a ID composed of the following:
#  {ORG_ID}/{USER_ID}/{ROLE_GROUP_ID}

terraform import \
  -var-file params.tfvars.json \                                            #variables file
  anypoint_user_rolegroup.user_rolegroup \                                  #resource name
  aa1f55d6-213d-4f60-845c-201282484cd1/18f23771-c78a-4be2-af8f-1bae66f43942/00dc3850-1e83-4b3b-918b-86aa646e0daf    #resource ID
