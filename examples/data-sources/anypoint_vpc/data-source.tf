 data "anypoint_vpc" "avpc" {
   org_id = "YOUR_ORG_ID"
   id     = "YOUR_VPC_ID"
 }

 output "vpc" {
   value = data.anypoint_vpc.avpc
 }