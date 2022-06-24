data "anypoint_vpn" "avpn" {
   org_id = "YOUR_ORG_ID"
   id     = "YOUR_VPN_ID"
   vpc_id = "YOUR_VPC_ID"
 }

 output "vpn" {
   value = data.anypoint_vpn.avpn
 }