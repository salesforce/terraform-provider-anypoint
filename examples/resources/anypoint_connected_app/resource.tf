resource "anypoint_connected_app" "my_conn_app_its_own_behalf" {
    name = "its own behalf"
    grant_types = ["client_credentials"]
    audience = "internal"

    scope {
        scope = "profile"
    }

    scope {
        scope = "aeh_admin"
        org_id = var.org_id
    }
    
    scope {
        scope = "read:audit_logs"
        org_id = var.org_id
    }

    scope {
        scope = "view:environment"
        org_id = var.org_id
        env_id = var.env_id
    }

    scope {
        scope = "edit:environment"
        org_id = var.org_id
        env_id = var.env_id
    }
}

resource "anypoint_connected_app" "my_conn_app_behalf_of_user" {
    name = "behalf of user"
    grant_types = [
        "authorization_code",
        "refresh_token",
        "password",
        "urn:ietf:params:oauth:grant-type:jwt-bearer"
    ]
    
    audience = "everyone"
    client_uri = "https://mysite.com"
    redirect_uris = [
        "https://myothersitei.com"
    ]
    public_keys = [
        "some_public_key"
    ]

    scope {
        scope = "full"
    }

    scope {
        scope = "read:full"
    }
}