module github.com/mulesoft-consulting/terraform-provider-cloudhub

go 1.15

require ( 
  github.com/hashicorp/terraform-plugin-sdk/v2 v2.4.4
  github.com/mulesoft-consulting/cloudhub-client-go/vpc v1.0.0
  github.com/mulesoft-consulting/cloudhub-client-go/authenticate v1.0.0
)

replace (
  github.com/mulesoft-consulting/cloudhub-client-go/vpc v1.0.0 => "../cloudhub-client-go/vpc"
  github.com/mulesoft-consulting/cloudhub-client-go/authenticate v1.0.0 => ../cloudhub-client-go/authenticate
)
