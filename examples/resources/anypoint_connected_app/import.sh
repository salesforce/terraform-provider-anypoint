# In order for the import to work, you should provide a ID composed of the following:
#  {CONNECTED_APP_ID}

terraform import \
  -var-file params.tfvars.json \          #variables file
  anypoint_connected_app.app \            #resource name
  rer123ze-213d-7f10-344c-909282484rr3    #resource ID
