# In order for the import to work, you should provide a ID composed of the following:
#  {ORG_ID}/{ENV_ID}/{DEPLOYMENT_ID}

terraform import \
  -var-file params.tfvars.json \                  #variables file
  anypoint_rtf_deployment.deployment \             #resource name
  aa1f55d6-213d-4f60-845c-201282484cd1/7074fcdd-9b23-4ae3-97e8-5db5f4adf17e/de32fc9d-6b25-4d6f-bd5e-cac32272b2f7    #resource ID
