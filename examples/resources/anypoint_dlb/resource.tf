
resource "anypoint_dlb" "dlb" {
  org_id = var.root_org
  vpc_id = anypoint_vpc.vpc.id
  name = "my-tf-dlb"
  state = "started"               # 'started', 'stopped' or 'restarted'
  ip_whitelist = []
  http_mode = "redirect"
  tlsv1 = false
  ssl_endpoints {
    public_key_label = "tf-public-key-name"
    public_key = "-----BEGIN CERTIFICATE-----\nMIICpDCCAYwCCQCOpE/9k0ve8zANBgkqhkiG9w0BAQsFADAUMRIwEAYDVQQDDAls\nb2NhbGhvc3QwHhcNMjEwMzA1MTUyMTM1WhcNMjEwMzA2MTUyMTM1WjAUMRIwEAYD\nVQQDDAlsb2NhbGhvc3QwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDK\n93gvOvMrcyVUvPnzC2UtXzHnV+rxW8I6VM+lFASV2FS+oZtiNGCFlbeEEMCImtAx\npaBw8/GTX5qNshFYNkGkvM4uh2PxYPZXfhOhkO42R6zdL89yTkY7E6nT/HwDUVAC\njJw67Y88St9h8yN5OOU95V3qkCbqfGxpKXnxmzTQt8aDRZQz5juQazVjMo4lIEpB\nuTPbXHRnHJCyr0OBOcGAGBTq2d7z2mFFlE+5w7RIiPNtx5KvG7wfO6KrCwfUGU5j\nl8466kfniqydGbxH7dsR+daPWAHrTCmZND7AWSiptIVzoJ/Q3QgT/qK8/SmpW9Hf\nDJQffO+I5y+w6y5cU1l3AgMBAAEwDQYJKoZIhvcNAQELBQADggEBAGS1mTWes3za\nWGlubGf76TiSn8GjIO7jIeVxBeB6rYq6iUFLUfEPCNHSlA0g7JJ40KW/osPc6EEm\nQzptRdhAoRDM5ilRTVMvuoGflw04OqrSUqR26+7aVJ8JcBJWBeP/5kGaMjPhy7oX\ntYPwzK2wXDYLDUCLXefF59NQoHUtytritckT5tP0UYDcRf2upBxn/v9lbF7AVfLZ\nO/vGplnD8Kq4QaFGL26ioh7e/n9TldbDJnspHh389aG6nqOKIgnL785Ggr6914vH\n4AMJa3r9cYpoe9ZdXL6b3aW+9MQo2Th2hDc7Z4CfVzJTZ9mg3ouKxIYGj+B4bj61\nN+MUQ5Q7aCo=\n-----END CERTIFICATE-----"
    private_key_label = "tf-private-key-name"
    private_key = "-----BEGIN PRIVATE KEY-----\nMIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDK93gvOvMrcyVU\nvPnzC2UtXzHnV+rxW8I6VM+lFASV2FS+oZtiNGCFlbeEEMCImtAxpaBw8/GTX5qN\nshFYNkGkvM4uh2PxYPZXfhOhkO42R6zdL89yTkY7E6nT/HwDUVACjJw67Y88St9h\n8yN5OOU95V3qkCbqfGxpKXnxmzTQt8aDRZQz5juQazVjMo4lIEpBuTPbXHRnHJCy\nr0OBOcGAGBTq2d7z2mFFlE+5w7RIiPNtx5KvG7wfO6KrCwfUGU5jl8466kfniqyd\nGbxH7dsR+daPWAHrTCmZND7AWSiptIVzoJ/Q3QgT/qK8/SmpW9HfDJQffO+I5y+w\n6y5cU1l3AgMBAAECggEAe0TfZnf8FiiBxLxdZeJG2c6WJXY9B8d96CV4Uz8cJdHU\nbk8Caxt6f8dVRM1T0eOMjIqWLePKlYIcAPDkHdod9iqBYrrx1TjZhHva+mZmdusD\nLvcJm9e0Sc8AdvJCc1VgLZwuio+bTbf/gaLEqawHdpcmef6A1CsrQJdjK3zjD9tn\n45wk+S6lRoCdGvFXk8L/mZPhhktzTRA4GKODKKzfXtMPXpjzj9sY500KwnjBDsNW\nxg7acYA2NbvdZqStGWP3O56gpttH8Ye9JbYCwIFYiPq9KnXJMYYb/k1/qSI4LNPX\nSuv0xmj6QNnRh3sfPHIynd+iKIm0qvqpBl2Chg9UeQKBgQD2peuK8iuvl2P61d5V\nR5RlyjTMKL9f1Pm5Q+vhcD2q2Ubow4iQWUyMwMFHIxvscSDkD8+sneOz85WHfZx9\nOK8oX3MHHDkkWxs6lJBnHBayFHtbuiI0LfJzSGGio672rEmS3A7g8ZDx06QczaD5\nhVhaR1Z7z9PfHW2rBOOJFEjl6wKBgQDSqY6kvYwet4kCdTUTnMuJuZ5u85Yn8jjU\nlZgAsizYwvWWXlUEYIKlosOfc/j1NQejqoVDgsQSFqfHDEG4gnClnEXi5tBg+OhX\n/rolaak+fuJ/dLj0RrkAJGvymDsf6qZoXtV6winO6Y7D5vtcaaWBo3DqaD4+28n3\nM1/m3I47pQKBgBkueWzXKrSjrTZ3zVpBk5oM2fUaF+fN060hjRyYHAOsaTvscq3i\nIBmiuFjt8bTjG+uM3bQO7qd5sAOERIzYU7G4hQLt07utfYsujcupJ3wI8Us9Jq7T\nHhS9CBLVyVAv6NcQlohKwXSfGftC1zOCdLHK5L6BSm1WENNMDXr6UjL/AoGAWKwq\ncMmga2WR9EjluIWtXyGUwNsjf1kD9ueo/dIB8pPN0CeQ3bDKDXJ/qWSljIFv38Jt\nKcenRH3ozW4pU8MEK5GmESZa3BappjCApjLdnILIUCIPoDMMuDScg5b0fDDHLvOM\nJIoKEyBYibl2YKXPlsv3QZPzb34Qe09StNhtvkkCgYEA2tOGGyiqcjG1fDhvdYvf\nbpja2/7OetClQKmjQJRLECRkJmEJk/mpOruyFn9cg/4wPBVi2AqMCqG/KyTzuImT\nY/kqPJ+UmYLBDnxIXzff/6nUjuxTZXgcdtnlaK/xq2HoU3XsCyHjOcaCjIUSLQsx\neb6YXmFBGK62BISiWmm3aPQ=\n-----END PRIVATE KEY-----"
    verify_client_mode = "off"
    mappings {
      input_uri = "{app}/"
      app_name = "{app}"
      app_uri = "/" 
    }
  }
}