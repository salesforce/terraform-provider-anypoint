package anypoint

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	private_space "github.com/mulesoft-anypoint/anypoint-client-go/private_space"
)

func resourcePrivateSpaceVpn() *schema.Resource {
	return &schema.Resource{
		// SchemaVersion stays at 0 for the initial release. Bump and add
		// StateUpgraders if the on-disk shape ever changes (renaming nested
		// attributes, restructuring tunnels, etc.).
		SchemaVersion: 0,
		CreateContext: resourcePrivateSpaceVpnCreate,
		ReadContext:   resourcePrivateSpaceVpnRead,
		UpdateContext: resourcePrivateSpaceVpnUpdate,
		DeleteContext: resourcePrivateSpaceVpnDelete,
		CustomizeDiff: customizeDiffPrivateSpaceVpn,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Manages a VPN connection on an Anypoint CloudHub 2.0 private space. " +
			"A connection groups one or more VPN members (BGP or static-routes), each composed of two IPsec tunnels. " +
			"Mutable in place: connection `name`, member `name`, member `static_routes`, member `startup_action`, and appending or removing members at the tail of the list. " +
			"Changes to `local_asn`, `remote_asn`, `remote_ip_address`, `psk` or `ptp_cidr` force replacement of the whole connection.",
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Composite id ORG_ID/PRIVATE_SPACE_ID/CONNECTION_ID.",
			},
			"last_updated": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The last time this resource has been updated locally.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The organization id where the private space lives.",
			},
			"private_space_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The private space id.",
			},
			"connection_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The VPN connection id (GUID).",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Display name of the VPN connection.",
			},
			"vpns": {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Description: "VPN members composing the connection.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Display name of the VPN member.",
						},
						"local_asn": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Local BGP ASN announced by Anypoint (string-encoded integer). Changing this on an existing VPN member forces replacement of the whole connection.",
						},
						"remote_asn": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Remote BGP ASN announced by the customer endpoint. Omit when using static routes. Changing this on an existing VPN member forces replacement.",
						},
						"remote_ip_address": {
							Type:             schema.TypeString,
							Required:         true,
							Description:      "Public IP address of the customer VPN endpoint. Changing this on an existing VPN member forces replacement.",
							ValidateDiagFunc: validation.ToDiagFunc(validation.IsIPv4Address),
						},
						"static_routes": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Static routes (CIDR strings) carried over the tunnel. Mutually exclusive with BGP. Mutable — replaced in-place via PATCH.",
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: validation.ToDiagFunc(validation.IsCIDR),
							},
						},
						"startup_action": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Tunnel startup action applied to ALL tunnels of this VPN member. 'start' means Anypoint initiates the tunnel (Automatic Tunnel Initiation ON in the UI); 'add' means it only responds to traffic from the customer gateway. Mutable via PATCH. Known upstream issue: Anypoint's GET endpoint returns the stale value 'start' even after a successful PATCH to 'add' (the PATCH does take effect — the PATCH response itself confirms the new value). Until the platform fixes this, terraform will plan a perpetual `start -> add` diff on every refresh when the desired value is 'add'. Workaround: `lifecycle { ignore_changes = [vpns] }`.",
							ValidateFunc: func(val any, key string) (warns []string, errs []error) {
								v := val.(string)
								if v != "start" && v != "add" {
									errs = append(errs, fmt.Errorf("%q must be 'start' or 'add', got: %s", key, v))
								}
								return
							},
						},
						"vpn_tunnels": {
							Type:        schema.TypeList,
							Required:    true,
							MinItems:    1,
							MaxItems:    2,
							Description: "IPsec tunnels composing this VPN member (typically two for high availability).",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"psk": {
										Type:        schema.TypeString,
										Required:    true,
										Sensitive:   true,
										Description: "Pre-shared key used to authenticate the IPsec tunnel. 8-64 characters; alphanumerics plus '.' and '_' only; must NOT start with '0'. Write-only — the API never returns it, so drift is not detected. Changing this on an existing tunnel forces replacement.",
									},
									"ptp_cidr": {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										Description:      "Point-to-point /30 CIDR used for the inside-tunnel addresses. Empty in API responses. Changing this on an existing tunnel forces replacement.",
										ValidateDiagFunc: validation.ToDiagFunc(validation.IsCIDR),
									},
									"is_logs_enabled": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "True when tunnel logs are enabled.",
									},
								},
							},
						},
						"vpn_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "VPN member id (GUID).",
						},
						"connection_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Parent connection id (GUID).",
						},
						"connection_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Parent connection display name.",
						},
						"vpn_connection_status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Current status of this VPN member.",
						},
					},
				},
			},
		},
	}
}

func resourcePrivateSpaceVpnCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	psid := d.Get("private_space_id").(string)
	authctx := getPrivateSpaceAuthCtx(ctx, &pco)
	body := newPrivateSpaceVpnConnectionPostBody(d)
	res, httpr, err := pco.privatespaceclient.DefaultAPI.CreatePrivateSpaceVpnConnection(authctx, orgid, psid).PrivateSpaceVpnConnectionPostBody(*body).Execute()
	if err != nil {
		diags = append(diags, diagFromHttp("Unable to Create vpn connection "+d.Get("name").(string)+" for private space "+psid, httpr, err))
		return diags
	}
	defer httpr.Body.Close()
	cid := res.GetId()
	d.SetId(ComposeResourceId([]string{orgid, psid, cid}))
	d.Set("connection_id", cid)

	// POST /connections ignores per-vpn name. Patch each member individually to apply
	// the user-supplied name. Pair plan vpns[] with returned vpns[] positionally.
	planVpns, _ := d.Get("vpns").([]any)
	createdVpns := res.GetVpns()
	for i, raw := range planVpns {
		if i >= len(createdVpns) {
			break
		}
		pm, _ := raw.(map[string]any)
		desiredName, _ := pm["name"].(string)
		if desiredName == "" || desiredName == createdVpns[i].GetName() {
			continue
		}
		patch := private_space.NewPrivateSpaceVpnPatchBody()
		patch.SetName(desiredName)
		_, pr, perr := pco.privatespaceclient.DefaultAPI.UpdatePrivateSpaceVpnConnectionMember(authctx, orgid, psid, cid, createdVpns[i].GetVpnId()).PrivateSpaceVpnPatchBody(*patch).Execute()
		if perr != nil {
			diags = append(diags, diagFromHttp(fmt.Sprintf("Unable to apply name %q to vpn member %s", desiredName, createdVpns[i].GetVpnId()), pr, perr))
			return diags
		}
		if pr != nil {
			pr.Body.Close()
		}
	}

	return resourcePrivateSpaceVpnRead(ctx, d, m)
}

func resourcePrivateSpaceVpnRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid, psid, cid, derr := splitPrivateSpaceVpnId(d)
	if derr.HasError() {
		return derr
	}
	authctx := getPrivateSpaceAuthCtx(ctx, &pco)
	res, httpr, err := pco.privatespaceclient.DefaultAPI.GetPrivateSpaceVpnConnection(authctx, orgid, psid, cid).Execute()
	if err != nil {
		if httpr != nil && httpr.StatusCode == 404 {
			defer httpr.Body.Close()
			d.SetId("")
			return diags
		}
		diags = append(diags, diagFromHttp("Unable to Read vpn connection "+cid+" for private space "+psid, httpr, err))
		return diags
	}
	defer httpr.Body.Close()
	if err := setPrivateSpaceVpnConnectionAttrs(d, res); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set vpn connection attributes for " + cid,
			Detail:   err.Error(),
		})
		return diags
	}
	d.Set("org_id", orgid)
	d.Set("private_space_id", psid)
	d.Set("connection_id", cid)
	return diags
}

func resourcePrivateSpaceVpnUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid, psid, cid, derr := splitPrivateSpaceVpnId(d)
	if derr.HasError() {
		return derr
	}
	authctx := getPrivateSpaceAuthCtx(ctx, &pco)

	if d.HasChange("name") {
		body := private_space.NewPrivateSpaceVpnConnectionPatchBody()
		body.SetName(d.Get("name").(string))
		_, httpr, err := pco.privatespaceclient.DefaultAPI.UpdatePrivateSpaceVpnConnection(authctx, orgid, psid, cid).PrivateSpaceVpnConnectionPatchBody(*body).Execute()
		if err != nil {
			diags = append(diags, diagFromHttp("Unable to Update vpn connection name for "+cid, httpr, err))
			return diags
		}
		defer httpr.Body.Close()
	}

	if d.HasChange("vpns") {
		if d := reconcilePrivateSpaceVpnMembers(ctx, &pco, orgid, psid, cid, d); d.HasError() {
			return d
		}
	}

	d.Set("last_updated", time.Now().Format(time.RFC850))
	return resourcePrivateSpaceVpnRead(ctx, d, m)
}

func resourcePrivateSpaceVpnDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid, psid, cid, derr := splitPrivateSpaceVpnId(d)
	if derr.HasError() {
		return derr
	}
	authctx := getPrivateSpaceAuthCtx(ctx, &pco)
	httpr, err := pco.privatespaceclient.DefaultAPI.DeletePrivateSpaceVpnConnection(authctx, orgid, psid, cid).Execute()
	if err != nil {
		if httpr != nil && httpr.StatusCode == 404 {
			defer httpr.Body.Close()
			d.SetId("")
			return diags
		}
		diags = append(diags, diagFromHttp("Unable to Delete vpn connection "+cid+" for private space "+psid, httpr, err))
		return diags
	}
	defer httpr.Body.Close()
	d.SetId("")
	return diags
}

// reconcilePrivateSpaceVpnMembers applies positional diffs between old and new vpns:
//   - if a position exists in both old and new with the same computed vpn_id, PATCH if name/startup_action changed
//     and reject other in-place changes;
//   - if a position exists only in new (or new entry has no vpn_id), POST add;
//   - if a position exists only in old, DELETE the member.
func reconcilePrivateSpaceVpnMembers(ctx context.Context, pco *ProviderConfOutput, orgid, psid, cid string, d *schema.ResourceData) diag.Diagnostics {
	var diags diag.Diagnostics
	authctx := getPrivateSpaceAuthCtx(ctx, pco)
	oldRaw, newRaw := d.GetChange("vpns")
	oldList, _ := oldRaw.([]any)
	newList, _ := newRaw.([]any)

	// Reverse-iterate deletions to keep state indices stable.
	for i := len(oldList) - 1; i >= 0; i-- {
		om := oldList[i].(map[string]any)
		vpnId := om["vpn_id"].(string)
		if i < len(newList) && newList[i] != nil {
			nm := newList[i].(map[string]any)
			nVpnId, _ := nm["vpn_id"].(string)
			if nVpnId == vpnId || nVpnId == "" {
				continue
			}
		}
		if vpnId == "" {
			continue
		}
		httpr, err := pco.privatespaceclient.DefaultAPI.DeletePrivateSpaceVpnConnectionMember(authctx, orgid, psid, cid, vpnId).Execute()
		if err != nil {
			if httpr == nil || httpr.StatusCode != 404 {
				diags = append(diags, diagFromHttp("Unable to Delete vpn member "+vpnId+" from connection "+cid, httpr, err))
				return diags
			}
		}
		if httpr != nil {
			httpr.Body.Close()
		}
	}

	// Updates / additions, in order.
	for i, raw := range newList {
		nm := raw.(map[string]any)
		nVpnId, _ := nm["vpn_id"].(string)
		if nVpnId == "" {
			body := buildPrivateSpaceVpnPostBody(nm)
			res, httpr, err := pco.privatespaceclient.DefaultAPI.AddPrivateSpaceVpnConnectionMember(authctx, orgid, psid, cid).PrivateSpaceVpnPostBody(*body).Execute()
			if err != nil {
				diags = append(diags, diagFromHttp(fmt.Sprintf("Unable to Add vpn member at position %d to connection %s", i, cid), httpr, err))
				return diags
			}
			httpr.Body.Close()
			desiredName, _ := nm["name"].(string)
			if desiredName != "" && res != nil {
				newMembers := res.GetVpns()
				if i < len(newMembers) && newMembers[i].GetName() != desiredName {
					patch := private_space.NewPrivateSpaceVpnPatchBody()
					patch.SetName(desiredName)
					_, pr, perr := pco.privatespaceclient.DefaultAPI.UpdatePrivateSpaceVpnConnectionMember(authctx, orgid, psid, cid, newMembers[i].GetVpnId()).PrivateSpaceVpnPatchBody(*patch).Execute()
					if perr != nil {
						diags = append(diags, diagFromHttp(fmt.Sprintf("Unable to apply name %q to vpn member %s", desiredName, newMembers[i].GetVpnId()), pr, perr))
						return diags
					}
					if pr != nil {
						pr.Body.Close()
					}
				}
			}
			continue
		}
		if i >= len(oldList) {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Cannot mutate existing vpn member " + nVpnId,
				Detail:   "Position " + fmt.Sprint(i) + " holds an existing vpn_id but is not present in the old state. Re-create the resource.",
			})
			return diags
		}
		om := oldList[i].(map[string]any)
		oVpnId, _ := om["vpn_id"].(string)
		if oVpnId != nVpnId {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Vpn member position reordering is not supported",
				Detail:   fmt.Sprintf("Position %d had vpn_id=%s, plan has vpn_id=%s. Append/remove only at the tail.", i, oVpnId, nVpnId),
			})
			return diags
		}
		if vpnMemberHasMutableChange(om, nm) {
			body := buildPrivateSpaceVpnPatchBody(om, nm)
			_, httpr, err := pco.privatespaceclient.DefaultAPI.UpdatePrivateSpaceVpnConnectionMember(authctx, orgid, psid, cid, nVpnId).PrivateSpaceVpnPatchBody(*body).Execute()
			if err != nil {
				diags = append(diags, diagFromHttp("Unable to Update vpn member "+nVpnId, httpr, err))
				return diags
			}
			httpr.Body.Close()
		}
	}
	return diags
}

func vpnMemberHasMutableChange(oldM, newM map[string]any) bool {
	if isEmptyOrNil(oldM["name"]) {
		// Post-import: trust the user's config, no PATCH needed.
	} else if oldM["name"] != newM["name"] {
		return true
	}
	if !stringSliceEqual(oldM["static_routes"], newM["static_routes"]) {
		return true
	}
	if isEmptyOrNil(oldM["startup_action"]) {
		// Post-import: API may not have populated startup_action in state; do
		// not trigger a PATCH that would re-init the actual tunnel.
	} else if oldM["startup_action"] != newM["startup_action"] {
		return true
	}
	return false
}

func isEmptyOrNil(v any) bool {
	if v == nil {
		return true
	}
	s, ok := v.(string)
	return ok && s == ""
}

func buildPrivateSpaceVpnPatchBody(oldM, newM map[string]any) *private_space.PrivateSpaceVpnPatchBody {
	body := private_space.NewPrivateSpaceVpnPatchBody()
	if oldM["name"] != newM["name"] {
		body.SetName(newM["name"].(string))
	}
	if !stringSliceEqual(oldM["static_routes"], newM["static_routes"]) {
		routes, _ := newM["static_routes"].([]any)
		body.SetStaticRoutes(ListInterface2ListStrings(routes))
	}
	if oldM["startup_action"] != newM["startup_action"] {
		startupAction := newM["startup_action"].(string)
		newTunnels, _ := newM["vpn_tunnels"].([]any)
		tunnels := make([]private_space.PrivateSpaceVpnTunnelPatchBody, 0, len(newTunnels))
		for range newTunnels {
			patch := private_space.NewPrivateSpaceVpnTunnelPatchBody()
			patch.SetStartupAction(startupAction)
			tunnels = append(tunnels, *patch)
		}
		body.SetVpnTunnels(tunnels)
	}
	return body
}

func buildPrivateSpaceVpnPostBody(m map[string]any) *private_space.PrivateSpaceVpnPostBody {
	startupAction := m["startup_action"].(string)
	tunnelsRaw, _ := m["vpn_tunnels"].([]any)
	tunnels := make([]private_space.PrivateSpaceVpnTunnelPostBody, 0, len(tunnelsRaw))
	for _, t := range tunnelsRaw {
		tm := t.(map[string]any)
		tun := private_space.NewPrivateSpaceVpnTunnelPostBody(
			tm["psk"].(string),
			startupAction,
		)
		if cidr, ok := tm["ptp_cidr"].(string); ok && cidr != "" {
			tun.SetPtpCidr(cidr)
		}
		tunnels = append(tunnels, *tun)
	}
	body := private_space.NewPrivateSpaceVpnPostBody(
		m["local_asn"].(string),
		m["remote_ip_address"].(string),
		tunnels,
	)
	if n, ok := m["name"].(string); ok && n != "" {
		body.SetName(n)
	}
	if asn, ok := m["remote_asn"].(string); ok && asn != "" {
		body.SetRemoteAsn(asn)
	}
	if routes, ok := m["static_routes"].([]any); ok && len(routes) > 0 {
		body.SetStaticRoutes(ListInterface2ListStrings(routes))
	}
	return body
}

func newPrivateSpaceVpnConnectionPostBody(d *schema.ResourceData) *private_space.PrivateSpaceVpnConnectionPostBody {
	raw, _ := d.Get("vpns").([]any)
	vpns := make([]private_space.PrivateSpaceVpnPostBody, 0, len(raw))
	for _, v := range raw {
		vpns = append(vpns, *buildPrivateSpaceVpnPostBody(v.(map[string]any)))
	}
	return private_space.NewPrivateSpaceVpnConnectionPostBody(d.Get("name").(string), vpns)
}

// setPrivateSpaceVpnConnectionAttrs writes API state back to Terraform state.
// psk and ptp_cidr come back empty from the API; preserve the existing state values to avoid drift.
func setPrivateSpaceVpnConnectionAttrs(d *schema.ResourceData, conn *private_space.PrivateSpaceVpnConnection) error {
	if err := d.Set("name", conn.GetName()); err != nil {
		return err
	}
	existing, _ := d.Get("vpns").([]any)
	return d.Set("vpns", flattenPrivateSpaceVpns(conn.GetVpns(), existing))
}

func flattenPrivateSpaceVpns(vpns []private_space.PrivateSpaceVpn, existing []any) []map[string]any {
	out := make([]map[string]any, 0, len(vpns))
	for i, v := range vpns {
		var existingMember map[string]any
		if i < len(existing) {
			existingMember, _ = existing[i].(map[string]any)
		}
		tunnels := v.GetVpnTunnels()
		startupAction := ""
		if len(tunnels) > 0 {
			startupAction = tunnels[0].GetStartupAction()
		}
		if startupAction == "" && existingMember != nil {
			if prev, ok := existingMember["startup_action"].(string); ok {
				startupAction = prev
			}
		}
		// remote_asn is set only on BGP members; the API omits it on static members.
		// Emit "" for the unset case so HasChange("vpns.N.remote_asn") works correctly
		// when transitioning between modes.
		remoteAsnStr := ""
		if v.GetRemoteAsn() != 0 {
			remoteAsnStr = fmt.Sprintf("%d", v.GetRemoteAsn())
		}
		entry := map[string]any{
			"name":                  v.GetName(),
			"local_asn":             fmt.Sprintf("%d", v.GetLocalAsn()),
			"remote_asn":            remoteAsnStr,
			"remote_ip_address":     v.GetRemoteIpAddress(),
			"static_routes":         v.GetStaticRoutes(),
			"startup_action":        startupAction,
			"vpn_tunnels":           flattenPrivateSpaceVpnTunnels(tunnels, existingMember),
			"vpn_id":                v.GetVpnId(),
			"connection_id":         v.GetConnectionId(),
			"connection_name":       v.GetConnectionName(),
			"vpn_connection_status": v.GetVpnConnectionStatus(),
		}
		out = append(out, entry)
	}
	return out
}

func flattenPrivateSpaceVpnTunnels(tunnels []private_space.PrivateSpaceVpnTunnel, existingMember map[string]any) []map[string]any {
	var existingTunnels []any
	if existingMember != nil {
		existingTunnels, _ = existingMember["vpn_tunnels"].([]any)
	}
	out := make([]map[string]any, 0, len(tunnels))
	for i, t := range tunnels {
		psk := t.GetPsk()
		ptpCidr := t.GetPtpCidr()
		if i < len(existingTunnels) {
			et, _ := existingTunnels[i].(map[string]any)
			if et != nil {
				if psk == "" {
					if prev, ok := et["psk"].(string); ok {
						psk = prev
					}
				}
				if ptpCidr == "" {
					if prev, ok := et["ptp_cidr"].(string); ok {
						ptpCidr = prev
					}
				}
			}
		}
		out = append(out, map[string]any{
			"psk":             psk,
			"ptp_cidr":        ptpCidr,
			"is_logs_enabled": t.GetIsLogsEnabled(),
		})
	}
	return out
}

func splitPrivateSpaceVpnId(d *schema.ResourceData) (string, string, string, diag.Diagnostics) {
	var diags diag.Diagnostics
	s := DecomposeResourceId(d.Id())
	if len(s) != 3 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid vpn connection resource id",
			Detail:   fmt.Sprintf("Expected ORG_ID/PRIVATE_SPACE_ID/CONNECTION_ID, got %s", d.Id()),
		})
		return "", "", "", diags
	}
	return s[0], s[1], s[2], diags
}

func diagFromHttp(summary string, httpr *http.Response, err error) diag.Diagnostic {
	details := extractAPIErrorDetail(err, httpr)
	return diag.Diagnostic{
		Severity: diag.Error,
		Summary:  summary,
		Detail:   details,
	}
}

func stringSliceEqual(a, b any) bool {
	as, _ := a.([]any)
	bs, _ := b.([]any)
	if len(as) != len(bs) {
		return false
	}
	for i := range as {
		if as[i] != bs[i] {
			return false
		}
	}
	return true
}

// customizeDiffPrivateSpaceVpn enforces two API constraints at plan time:
//
//  1. All VPN members in a connection must be homogeneous — either all dynamic (BGP via remote_asn)
//     or all static (static_routes). The API rejects mixed connections with 400.
//  2. Per-member immutable fields (local_asn, remote_asn, remote_ip_address, psk, ptp_cidr) can be
//     mutated only by replacing the resource. Changes to those fields on EXISTING positions force
//     replacement; new appended positions are exempt so the Update path can POST /vpns instead.
func customizeDiffPrivateSpaceVpn(ctx context.Context, d *schema.ResourceDiff, m any) error {
	raw, _ := d.Get("vpns").([]any)
	if len(raw) > 0 {
		anyDynamic := false
		anyStatic := false
		var firstLocalAsn string
		for i, v := range raw {
			vm, _ := v.(map[string]any)
			remoteAsn, _ := vm["remote_asn"].(string)
			routes, _ := vm["static_routes"].([]any)
			hasRoutes := len(routes) > 0
			hasAsn := remoteAsn != ""
			if hasRoutes && hasAsn {
				return fmt.Errorf("vpns[%d] sets both remote_asn and static_routes; pick one (BGP or static)", i)
			}
			if !hasRoutes && !hasAsn {
				return fmt.Errorf("vpns[%d] must set either static_routes (static mode) or remote_asn (BGP mode)", i)
			}
			if hasAsn {
				anyDynamic = true
			}
			if hasRoutes {
				anyStatic = true
			}
			localAsn, _ := vm["local_asn"].(string)
			if i == 0 {
				firstLocalAsn = localAsn
			} else if localAsn != firstLocalAsn {
				return fmt.Errorf("vpns[%d].local_asn=%q differs from vpns[0].local_asn=%q; all vpn members in a connection must share the same local_asn (the Anypoint API rejects mixed values)", i, localAsn, firstLocalAsn)
			}
		}
		if anyDynamic && anyStatic {
			return fmt.Errorf("all vpn members in a connection must be homogeneous (all BGP via remote_asn, or all static via static_routes); the Anypoint API rejects mixed connections")
		}
	}

	if !d.HasChange("vpns") {
		return nil
	}
	oldRaw, newRaw := d.GetChange("vpns")
	oldList, _ := oldRaw.([]any)
	newList, _ := newRaw.([]any)
	overlap := len(oldList)
	if len(newList) < overlap {
		overlap = len(newList)
	}
	// SDK v2: d.ForceNew on a TypeList parent ("vpns") is silently ignored.
	// Use the specific leaf path (e.g. "vpns.0.remote_asn") so the diff entry exists.
	for i := 0; i < overlap; i++ {
		om, _ := oldList[i].(map[string]any)
		nm, _ := newList[i].(map[string]any)
		if om == nil || nm == nil {
			continue
		}
		for _, field := range []string{"local_asn", "remote_asn", "remote_ip_address"} {
			if fmt.Sprint(om[field]) != fmt.Sprint(nm[field]) {
				return d.ForceNew(fmt.Sprintf("vpns.%d.%s", i, field))
			}
		}
		oldTunnels, _ := om["vpn_tunnels"].([]any)
		newTunnels, _ := nm["vpn_tunnels"].([]any)
		if len(oldTunnels) != len(newTunnels) {
			return d.ForceNew(fmt.Sprintf("vpns.%d.vpn_tunnels.#", i))
		}
		for j := range oldTunnels {
			ot, _ := oldTunnels[j].(map[string]any)
			nt, _ := newTunnels[j].(map[string]any)
			if ot == nil || nt == nil {
				continue
			}
			for _, tfield := range []string{"psk", "ptp_cidr"} {
				oldVal := fmt.Sprint(ot[tfield])
				newVal := fmt.Sprint(nt[tfield])
				// Suppress ForceNew when the state side is empty. This happens after
				// `terraform import`, since the Anypoint API never returns psk and
				// returns "" for ptp_cidr on static members. Without this check the
				// first apply after import would unnecessarily destroy the resource
				// to "set" a value that is in fact already what the user provided.
				if oldVal == "" {
					continue
				}
				if oldVal != newVal {
					return d.ForceNew(fmt.Sprintf("vpns.%d.vpn_tunnels.%d.%s", i, j, tfield))
				}
			}
		}
	}
	return nil
}
