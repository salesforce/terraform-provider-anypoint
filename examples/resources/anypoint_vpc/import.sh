# In order for the import to work, you should provide a ID composed of the following:
#  {ORG_ID}/{VPC_ID}

terraform import \
  -var-file params.tfvars.json \                                              #variables file
  anypoint_vpc.vpc01 \                                                        #resource name
  aa1f55d6-213d-4f60-845c-201282484cd1/vpc-0aea9f31a049ce288                  #resource ID
