package anypoint

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	private_space "github.com/mulesoft-anypoint/anypoint-client-go/private_space"
)

func resourcePrivateSpaceTransitGateway() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePrivateSpaceTransitGatewayCreate,
		ReadContext:   resourcePrivateSpaceTransitGatewayRead,
		UpdateContext: resourcePrivateSpaceTransitGatewayUpdate,
		DeleteContext: resourcePrivateSpaceTransitGatewayDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Manages an AWS Transit Gateway attachment on an Anypoint CloudHub 2.0 private space. " +
			"Requires a RAM resource share owned by your AWS account, with the MuleSoft AWS account whitelisted as principal and the TGW associated with the share.",
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Composite id ORG_ID/PRIVATE_SPACE_ID/TRANSIT_GATEWAY_ID.",
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
			"transit_gateway_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The AWS transit gateway id (e.g. tgw-...).",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Display name of the transit gateway attachment.",
			},
			"resource_share_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "RAM resource share UUID (trailing GUID of the ARN, NOT the full ARN).",
			},
			"resource_share_account": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "AWS account id that owns the RAM share (your AWS account, NOT MuleSoft's).",
			},
			"routes": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "Static routes (CIDR strings) pushed into the private space route table. Must not equal or be more specific than the private space CIDR.",
				Elem: &schema.Schema{
					Type:             schema.TypeString,
					ValidateDiagFunc: validation.ToDiagFunc(validation.IsCIDR),
				},
			},
			"account_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "MuleSoft AWS account id, region-bound to the private space.",
			},
			"region": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "AWS region (inherits the private space region).",
			},
			"space_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Private space display name.",
			},
			"status_gateway": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Gateway status (refreshing, available, ...).",
			},
			"status_attachment": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Attachment status (refreshing, available, ...).",
			},
			"tgw_resource_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "AWS console URL of the transit gateway.",
			},
		},
	}
}

func resourcePrivateSpaceTransitGatewayCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid := d.Get("org_id").(string)
	psid := d.Get("private_space_id").(string)
	authctx := getPrivateSpaceAuthCtx(ctx, &pco)
	body := newPrivateSpaceTransitGatewayPostBody(d)
	res, httpr, err := pco.privatespaceclient.DefaultAPI.CreatePrivateSpaceTransitGateway(authctx, orgid, psid).TransitGatewayPostBody(*body).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Create transit gateway " + d.Get("name").(string) + " for private space " + psid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	tgwid := res.GetId()
	d.SetId(ComposeResourceId([]string{orgid, psid, tgwid}))
	d.Set("transit_gateway_id", tgwid)
	return resourcePrivateSpaceTransitGatewayRead(ctx, d, m)
}

func resourcePrivateSpaceTransitGatewayRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid, psid, tgwid, derr := splitPrivateSpaceTransitGatewayId(d)
	if derr.HasError() {
		return derr
	}
	authctx := getPrivateSpaceAuthCtx(ctx, &pco)
	list, httpr, err := pco.privatespaceclient.DefaultAPI.GetPrivateSpaceTransitGateways(authctx, orgid, psid).Execute()
	if err != nil {
		if httpr != nil && httpr.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Read transit gateway " + tgwid + " for private space " + psid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	var found *private_space.TransitGateway
	for i := range list {
		if list[i].GetId() == tgwid {
			found = &list[i]
			break
		}
	}
	if found == nil {
		d.SetId("")
		return diags
	}
	if err := setPrivateSpaceTransitGatewayAttrs(d, found); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set transit gateway attributes for " + tgwid,
			Detail:   err.Error(),
		})
		return diags
	}
	d.Set("org_id", orgid)
	d.Set("private_space_id", psid)
	d.Set("transit_gateway_id", tgwid)
	return diags
}

func resourcePrivateSpaceTransitGatewayUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid, psid, tgwid, derr := splitPrivateSpaceTransitGatewayId(d)
	if derr.HasError() {
		return derr
	}
	authctx := getPrivateSpaceAuthCtx(ctx, &pco)
	if d.HasChange("routes") {
		body := private_space.NewTransitGatewayPatchRoutesBody(ListInterface2ListStrings(d.Get("routes").(*schema.Set).List()))
		_, httpr, err := pco.privatespaceclient.DefaultAPI.UpdatePrivateSpaceTransitGatewayRoutes(authctx, orgid, psid, tgwid).TransitGatewayPatchRoutesBody(*body).Execute()
		if err != nil {
			details := extractAPIErrorDetail(err, httpr)
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to Update routes for transit gateway " + tgwid,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()
	}
	if d.HasChange("name") {
		body := private_space.NewTransitGatewayPatchNameBody(d.Get("name").(string))
		_, httpr, err := pco.privatespaceclient.DefaultAPI.UpdateOrgTransitGatewayName(authctx, orgid, tgwid).TransitGatewayPatchNameBody(*body).Execute()
		if err != nil {
			details := extractAPIErrorDetail(err, httpr)
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to Update name for transit gateway " + tgwid,
				Detail:   details,
			})
			return diags
		}
		defer httpr.Body.Close()
	}
	d.Set("last_updated", time.Now().Format(time.RFC850))
	return resourcePrivateSpaceTransitGatewayRead(ctx, d, m)
}

func resourcePrivateSpaceTransitGatewayDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)
	orgid, psid, tgwid, derr := splitPrivateSpaceTransitGatewayId(d)
	if derr.HasError() {
		return derr
	}
	authctx := getPrivateSpaceAuthCtx(ctx, &pco)
	httpr, err := pco.privatespaceclient.DefaultAPI.DeletePrivateSpaceTransitGateway(authctx, orgid, psid, tgwid).Execute()
	if err != nil {
		details := extractAPIErrorDetail(err, httpr)
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Delete transit gateway " + tgwid + " for private space " + psid,
			Detail:   details,
		})
		return diags
	}
	defer httpr.Body.Close()
	d.SetId("")
	return diags
}

func newPrivateSpaceTransitGatewayPostBody(d *schema.ResourceData) *private_space.TransitGatewayPostBody {
	return private_space.NewTransitGatewayPostBody(
		d.Get("name").(string),
		d.Get("resource_share_id").(string),
		d.Get("resource_share_account").(string),
		ListInterface2ListStrings(d.Get("routes").(*schema.Set).List()),
	)
}

func setPrivateSpaceTransitGatewayAttrs(d *schema.ResourceData, tgw *private_space.TransitGateway) error {
	if err := d.Set("name", tgw.GetName()); err != nil {
		return err
	}
	if err := d.Set("account_id", tgw.GetAccountId()); err != nil {
		return err
	}
	spec := tgw.GetSpec()
	if err := d.Set("region", spec.GetRegion()); err != nil {
		return err
	}
	if err := d.Set("space_name", spec.GetSpaceName()); err != nil {
		return err
	}
	share := spec.GetResourceShare()
	if err := d.Set("resource_share_id", share.GetId()); err != nil {
		return err
	}
	if err := d.Set("resource_share_account", share.GetAccount()); err != nil {
		return err
	}
	status := tgw.GetStatus()
	if err := d.Set("status_gateway", status.GetGateway()); err != nil {
		return err
	}
	if err := d.Set("status_attachment", status.GetAttachment()); err != nil {
		return err
	}
	if err := d.Set("tgw_resource_url", status.GetTgwResource()); err != nil {
		return err
	}
	if err := d.Set("routes", status.GetRoutes()); err != nil {
		return err
	}
	return nil
}

func splitPrivateSpaceTransitGatewayId(d *schema.ResourceData) (string, string, string, diag.Diagnostics) {
	var diags diag.Diagnostics
	s := DecomposeResourceId(d.Id())
	if len(s) != 3 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid transit gateway resource id",
			Detail:   fmt.Sprintf("Expected ORG_ID/PRIVATE_SPACE_ID/TRANSIT_GATEWAY_ID, got %s", d.Id()),
		})
		return "", "", "", diags
	}
	return s[0], s[1], s[2], diags
}
