# In order for the import to work, you should provide a ID composed of the following:
#  {ORG_ID}/{PRIVATE_SPACE_ID}/{TLS_CONTEXT_ID}

terraform import \
  -var-file params.tfvars.json \        # variable file
  anypoint_private_space_tlscontext_pem.tls_context \
  aa1f55d6-213d-4f60-845c-201282484cd1/6fea33e9-f51c-4750-b182-312e842d4250/6fea33e9-f51c-4750-b182-312e842d4250
