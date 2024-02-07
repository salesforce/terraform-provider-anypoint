# In order for the import to work, you should provide a ID composed of the following:
#  {ORG_ID}/{TEAM_ID}/{USER_ID}

terraform import \
  -var-file params.tfvars.json \          #variables file
  anypoint_team_member.team_member \            #resource name
  aa1f55d6-213d-4f60-845c-201282484cd1/99c41e16-1075-40ae-8c8b-d722a8256f81/18f23771-c78a-4be2-af8f-1bae66f43942    #resource ID
