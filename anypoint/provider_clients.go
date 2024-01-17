package anypoint

import (
	ame "github.com/mulesoft-anypoint/anypoint-client-go/ame"
	ame_binding "github.com/mulesoft-anypoint/anypoint-client-go/ame_binding"
	amq "github.com/mulesoft-anypoint/anypoint-client-go/amq"
	apim "github.com/mulesoft-anypoint/anypoint-client-go/apim"
	apim_upstream "github.com/mulesoft-anypoint/anypoint-client-go/apim_upstream"
	connected_app "github.com/mulesoft-anypoint/anypoint-client-go/connected_app"
	dlb "github.com/mulesoft-anypoint/anypoint-client-go/dlb"
	env "github.com/mulesoft-anypoint/anypoint-client-go/env"
	flexgateway "github.com/mulesoft-anypoint/anypoint-client-go/flexgateway"
	idp "github.com/mulesoft-anypoint/anypoint-client-go/idp"
	org "github.com/mulesoft-anypoint/anypoint-client-go/org"
	role "github.com/mulesoft-anypoint/anypoint-client-go/role"
	rolegroup "github.com/mulesoft-anypoint/anypoint-client-go/rolegroup"
	secretgroup "github.com/mulesoft-anypoint/anypoint-client-go/secretgroup"
	secretgroup_certificate "github.com/mulesoft-anypoint/anypoint-client-go/secretgroup_certificate"
	secretgroup_crl_distributor_configs "github.com/mulesoft-anypoint/anypoint-client-go/secretgroup_crl_distributor_configs"
	secretgroup_keystore "github.com/mulesoft-anypoint/anypoint-client-go/secretgroup_keystore"
	secretgroup_tlscontext "github.com/mulesoft-anypoint/anypoint-client-go/secretgroup_tlscontext"
	secretgroup_truststore "github.com/mulesoft-anypoint/anypoint-client-go/secretgroup_truststore"
	team "github.com/mulesoft-anypoint/anypoint-client-go/team"
	team_group_mappings "github.com/mulesoft-anypoint/anypoint-client-go/team_group_mappings"
	team_members "github.com/mulesoft-anypoint/anypoint-client-go/team_members"
	team_roles "github.com/mulesoft-anypoint/anypoint-client-go/team_roles"
	user "github.com/mulesoft-anypoint/anypoint-client-go/user"
	user_rolegroups "github.com/mulesoft-anypoint/anypoint-client-go/user_rolegroups"
	vpc "github.com/mulesoft-anypoint/anypoint-client-go/vpc"
	vpn "github.com/mulesoft-anypoint/anypoint-client-go/vpn"
)

type ProviderConfOutput struct {
	access_token            string
	server_index            int
	vpcclient               *vpc.APIClient
	vpnclient               *vpn.APIClient
	orgclient               *org.APIClient
	roleclient              *role.APIClient
	rolegroupclient         *rolegroup.APIClient
	userclient              *user.APIClient
	envclient               *env.APIClient
	userrgpclient           *user_rolegroups.APIClient
	teamclient              *team.APIClient
	teammembersclient       *team_members.APIClient
	teamrolesclient         *team_roles.APIClient
	teamgroupmappingsclient *team_group_mappings.APIClient
	dlbclient               *dlb.APIClient
	idpclient               *idp.APIClient
	connectedappclient      *connected_app.APIClient
	amqclient               *amq.APIClient
	ameclient               *ame.APIClient
	amebindingclient        *ame_binding.APIClient
	apimclient              *apim.APIClient
	apimupstreamclient      *apim_upstream.APIClient
	flexgatewayclient       *flexgateway.APIClient
	secretgroupclient       *secretgroup.APIClient
	sgkeystoreclient        *secretgroup_keystore.APIClient
	sgtruststoreclient      *secretgroup_truststore.APIClient
	sgcertificateclient     *secretgroup_certificate.APIClient
	sgtlscontextclient      *secretgroup_tlscontext.APIClient
	sgcrldistribcfgsclient  *secretgroup_crl_distributor_configs.APIClient
}

func newProviderConfOutput(access_token string, server_index int) ProviderConfOutput {
	//preparing clients
	vpccfg := vpc.NewConfiguration()
	vpncfg := vpn.NewConfiguration()
	orgcfg := org.NewConfiguration()
	rolecfg := role.NewConfiguration()
	rolegroupcfg := rolegroup.NewConfiguration()
	usercfg := user.NewConfiguration()
	envcfg := env.NewConfiguration()
	userrolegroupscfg := user_rolegroups.NewConfiguration()
	teamcfg := team.NewConfiguration()
	teammemberscfg := team_members.NewConfiguration()
	teamrolescfg := team_roles.NewConfiguration()
	teamgroupmappingscfg := team_group_mappings.NewConfiguration()
	dlbcfg := dlb.NewConfiguration()
	idpcfg := idp.NewConfiguration()
	connectedappcfg := connected_app.NewConfiguration()
	amqcfg := amq.NewConfiguration()
	amecfg := ame.NewConfiguration()
	amebindingcfg := ame_binding.NewConfiguration()
	apimcfg := apim.NewConfiguration()
	apimupstreamcfg := apim_upstream.NewConfiguration()
	flexgatewaycfg := flexgateway.NewConfiguration()
	secretgroupcfg := secretgroup.NewConfiguration()
	sgkeystorecfg := secretgroup_keystore.NewConfiguration()
	sgtruststorecfg := secretgroup_truststore.NewConfiguration()
	sgcertificatecfg := secretgroup_certificate.NewConfiguration()
	sgtlscontextcfg := secretgroup_tlscontext.NewConfiguration()
	sgcrldistribcfgs_cfg := secretgroup_crl_distributor_configs.NewConfiguration()

	vpcclient := vpc.NewAPIClient(vpccfg)
	vpnclient := vpn.NewAPIClient(vpncfg)
	orgclient := org.NewAPIClient(orgcfg)
	roleclient := role.NewAPIClient(rolecfg)
	rolegroupclient := rolegroup.NewAPIClient(rolegroupcfg)
	userclient := user.NewAPIClient(usercfg)
	envclient := env.NewAPIClient(envcfg)
	userrgpclient := user_rolegroups.NewAPIClient(userrolegroupscfg)
	teamclient := team.NewAPIClient(teamcfg)
	teammembersclient := team_members.NewAPIClient(teammemberscfg)
	teamrolesclient := team_roles.NewAPIClient(teamrolescfg)
	teamgroupmappingsclient := team_group_mappings.NewAPIClient(teamgroupmappingscfg)
	dlbclient := dlb.NewAPIClient(dlbcfg)
	idpclient := idp.NewAPIClient(idpcfg)
	connectedappclient := connected_app.NewAPIClient(connectedappcfg)
	amqclient := amq.NewAPIClient(amqcfg)
	ameclient := ame.NewAPIClient(amecfg)
	amebindingclient := ame_binding.NewAPIClient(amebindingcfg)
	apimclient := apim.NewAPIClient(apimcfg)
	apimupstreamclient := apim_upstream.NewAPIClient(apimupstreamcfg)
	flexgatewayclient := flexgateway.NewAPIClient(flexgatewaycfg)
	secretgroupclient := secretgroup.NewAPIClient(secretgroupcfg)
	sgkeystoreclient := secretgroup_keystore.NewAPIClient(sgkeystorecfg)
	sgtruststoreclient := secretgroup_truststore.NewAPIClient(sgtruststorecfg)
	sgcertificateclient := secretgroup_certificate.NewAPIClient(sgcertificatecfg)
	sgtlscontextclient := secretgroup_tlscontext.NewAPIClient(sgtlscontextcfg)
	sgcrldistribcfgsclient := secretgroup_crl_distributor_configs.NewAPIClient(sgcrldistribcfgs_cfg)

	return ProviderConfOutput{
		access_token:            access_token,
		server_index:            server_index,
		vpcclient:               vpcclient,
		vpnclient:               vpnclient,
		orgclient:               orgclient,
		roleclient:              roleclient,
		rolegroupclient:         rolegroupclient,
		userclient:              userclient,
		envclient:               envclient,
		userrgpclient:           userrgpclient,
		teamclient:              teamclient,
		teammembersclient:       teammembersclient,
		teamrolesclient:         teamrolesclient,
		teamgroupmappingsclient: teamgroupmappingsclient,
		dlbclient:               dlbclient,
		idpclient:               idpclient,
		connectedappclient:      connectedappclient,
		amqclient:               amqclient,
		ameclient:               ameclient,
		amebindingclient:        amebindingclient,
		apimclient:              apimclient,
		apimupstreamclient:      apimupstreamclient,
		flexgatewayclient:       flexgatewayclient,
		secretgroupclient:       secretgroupclient,
		sgkeystoreclient:        sgkeystoreclient,
		sgtruststoreclient:      sgtruststoreclient,
		sgcertificateclient:     sgcertificateclient,
		sgtlscontextclient:      sgtlscontextclient,
		sgcrldistribcfgsclient:  sgcrldistribcfgsclient,
	}
}
