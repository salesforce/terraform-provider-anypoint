# In order for the import to work, you should provide a ID composed of the following:
#  {ORG_ID}/{TEAM_ID}

terraform import \
  -var-file params.tfvars.json \          #variables file
  anypoint_team_group_mappings.team_gmap \            #resource name
  aa1f55d6-213d-4f60-845c-201282484cd1/99c41e16-1075-40ae-8c8b-d722a8256f81    #resource ID
