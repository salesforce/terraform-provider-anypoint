 data "anypoint_vpcs" "all" {
   org_id = "YOUR_ORG_ID"
 }
 
 output "all" {
   value = data.anypoint_vpcs.all
 }