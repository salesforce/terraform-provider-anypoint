# In order for the import to work, you should provide a ID composed of the following:
#  {ORG_ID}/{VPC_ID}/{DLB_ID}

terraform import \
  -var-file params.tfvars.json \          #variables file
  anypoint_dlb.dlb \            #resource name
  aa1f55d6-213d-4f60-845c-201282484cd1/vpc-0cfd5cb3d3010cd44/64e8d2284b5e9b1c388d1bbc    #resource ID
