 data "anypoint_user" "user" {
   org_id = "YOUR_ORG_ID"
   id     = "YOUR_USER_ID"
 }

 output "user" {
   value = data.anypoint_user.user
 }