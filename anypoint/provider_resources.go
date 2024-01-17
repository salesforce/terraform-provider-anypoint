package anypoint

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

var RESOURCES_MAP = map[string]*schema.Resource{
	"anypoint_vpc":                                   resourceVPC(),
	"anypoint_vpn":                                   resourceVPN(),
	"anypoint_bg":                                    resourceBG(),
	"anypoint_rolegroup_roles":                       resourceRoleGroupRoles(),
	"anypoint_rolegroup":                             resourceRoleGroup(),
	"anypoint_env":                                   resourceENV(),
	"anypoint_user":                                  resourceUser(),
	"anypoint_user_rolegroup":                        resourceUserRolegroup(),
	"anypoint_team":                                  resourceTeam(),
	"anypoint_team_roles":                            resourceTeamRoles(),
	"anypoint_team_member":                           resourceTeamMember(),
	"anypoint_team_group_mappings":                   resourceTeamGroupMappings(),
	"anypoint_dlb":                                   resourceDLB(),
	"anypoint_idp_oidc":                              resourceOIDC(),
	"anypoint_idp_saml":                              resourceSAML(),
	"anypoint_connected_app":                         resourceConnectedApp(),
	"anypoint_amq":                                   resourceAMQ(),
	"anypoint_ame":                                   resourceAME(),
	"anypoint_ame_binding":                           resourceAMEBinding(),
	"anypoint_apim_flexgateway":                      resourceApimFlexGateway(),
	"anypoint_apim_mule4":                            resourceApimMule4(),
	"anypoint_secretgroup":                           resourceSecretGroup(),
	"anypoint_secretgroup_keystore":                  resourceSecretGroupKeystore(),
	"anypoint_secretgroup_truststore":                resourceSecretGroupTruststore(),
	"anypoint_secretgroup_certificate":               resourceSecretGroupCertificate(),
	"anypoint_secretgroup_tlscontext_flexgateway":    resourceSecretGroupTlsContextFG(),
	"anypoint_secretgroup_tlscontext_mule":           resourceSecretGroupTlsContextMule(),
	"anypoint_secretgroup_tlscontext_securityfabric": resourceSecretGroupTlsContextSF(),
	"anypoint_secretgroup_crldistrib_cfgs":           resourceSecretGroupCrlDistribCfgs(),
}
