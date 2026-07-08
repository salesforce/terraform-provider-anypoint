# In order for the import to work, you should provide a ID composed of the following:
#  {ORG_ID}/{PRIVATE_SPACE_ID}/{TRANSIT_GATEWAY_ID}

terraform import \
  -var-file params.tfvars.json \
  anypoint_private_space_transit_gateway.my_tgw \
  aa1f55d6-213d-4f60-845c-201282484cd1/7f747999-a9bb-41d5-bfe8-d6b8cca68c62/tgw-017e20b9ce00c865c
