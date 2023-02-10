# In order for the import to work, you should provide a ID composed of the following:
#  {ORG_ID}/{ENV_ID}/{REGION_ID}/{EXCHANGE_ID}/{QUEUE_ID}

terraform import \
  -var-file params.tfvars.json \    #variable file
  anypoint_ame_binding.ame_b \      #resource name
  aa1f55d6-213d-4f60-845c-201282484cd1/7074fcdd-9b23-4ab6-97r8-5db5f4adf17d/us-east-1/MY-AWESOME-EXCHANGE   #resource id
