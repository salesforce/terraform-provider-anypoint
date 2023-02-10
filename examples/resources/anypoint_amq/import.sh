# In order for the import to work, you should provide a ID composed of the following:
#  {ORG_ID}/{ENV_ID}/{REGION_ID}/{QUEUE_ID}

terraform import \
  -var-file params.tfvars.json \    #variables file
  anypoint_amq.amq \                #resource name
  aa1f55d6-213d-4f60-845c-201282484cd1/7074fcdd-9b23-4ab6-97r8-5db5f4adf17d/us-east-1/myAwesomeQ    #resource ID
