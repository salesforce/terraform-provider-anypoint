# In order for the import to work, you should provide an ID composed of the following:
#  {ORG_ID}/{TEAM_ID}/{ROLE_ID}/{CONTEXT_ORG}
# or, for environment-scoped roles:
#  {ORG_ID}/{TEAM_ID}/{ROLE_ID}/{CONTEXT_ORG}/{ENV_ID}

# Business-group-scoped role (4 segments):
terraform import \
  -var-file params.tfvars.json \
  anypoint_team_role.org_admin \
  aa1f55d6-213d-4f60-845c-201282484cd1/99c41e16-1075-40ae-8c8b-d722a8256f81/00000000-0000-0000-0000-000000000000/bb2f55d6-213d-4f60-845c-201282484cd2

# Environment-scoped role (5 segments):
terraform import \
  -var-file params.tfvars.json \
  anypoint_team_role.env_admin \
  aa1f55d6-213d-4f60-845c-201282484cd1/99c41e16-1075-40ae-8c8b-d722a8256f81/11111111-1111-1111-1111-111111111111/bb2f55d6-213d-4f60-845c-201282484cd2/cc3f55d6-213d-4f60-845c-201282484cd3
