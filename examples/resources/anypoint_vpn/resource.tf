//Values may reference to examples/datasources
//They are referenced the follow way: 
//{dataType}.{name}.{attribute}
resource "anypoint_vpn" "avpn" {
  org_id = anypoint_bg.bg.id
  vpc_id = anypoint_vpc.avpc.id
  name = "myDatacenterVpn"
  remote_asn = 65000
  remote_ip_address = "100.100.100.100"

  tunnel_configs {
    psk = "23847329fn3u__..."
    ptp_cidr = "169.254.12.0/30"
  }

  tunnel_configs {
    psk = "nsdkjfnsauf23f2linf"
    ptp_cidr = "169.254.12.4/30"
  }
}