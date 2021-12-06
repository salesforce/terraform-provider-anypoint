 data "anypoint_users" "users" {
   org_id = "YOUR_ORG_ID"
   params {
     offset = 0         # page number
     limit  = 200       # number of users per page
     type   = "all"     # users type
   }
 }

 output "users" {
   value = data.anypoint_users.users
 }