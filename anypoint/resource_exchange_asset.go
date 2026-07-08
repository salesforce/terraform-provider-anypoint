package anypoint

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	exchange_assets "github.com/mulesoft-anypoint/anypoint-client-go/exchange_assets"
)

func resourceExchangeAsset() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceExchangeAssetCreate,
		ReadContext:   resourceExchangeAssetRead,
		UpdateContext: resourceExchangeAssetUpdate,
		DeleteContext: resourceExchangeAssetDelete,
		Description:   "Creates and manages an Exchange asset. Asset metadata is updatable (name, description); all other attributes force replacement.",
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The last time this resource has been updated locally.",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The composite id of this resource: org_id/group_id/asset_id/version.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The organization id where the asset will be created.",
			},
			"group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The business group id the asset belongs to. Often equals org_id for root-level assets.",
			},
			"asset_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The asset slug.",
			},
			"version": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The asset version (must follow Semver, e.g. 1.0.0).",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The visible name of the asset.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The asset description.",
			},
			"classifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"mule-application",
					"raml-fragment",
					"raml",
					"oas",
					"wsdl",
					"http",
					"custom",
				}, false),
				Description: "The asset classifier. One of: mule-application, raml-fragment, raml, oas, wsdl, http, custom.",
			},
			"api_version": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: diffSuppressExchangeAssetWriteOnly,
				Description:      "The product version of API assets. Required for raml, oas, wsdl, http. Write-only: the Exchange API does not return it, so the value cannot be refreshed from state.",
			},
			"main_file": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: diffSuppressExchangeAssetWriteOnly,
				Description:      "The main file of the asset package. Required for raml, raml-fragment, oas, wsdl. Write-only.",
			},
			"asset_file": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ConflictsWith:    []string{"asset_link"},
				DiffSuppressFunc: diffSuppressExchangeAssetWriteOnly,
				Description:      "Local path to the asset file to upload. Maximum 5 MB. Mutually exclusive with asset_link. Write-only.",
			},
			"asset_link": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ConflictsWith:    []string{"asset_file"},
				DiffSuppressFunc: diffSuppressExchangeAssetWriteOnly,
				Description:      "Remote URL of the asset (for http or wsdl classifiers). Mutually exclusive with asset_file. Write-only.",
			},
			"dependencies": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: diffSuppressExchangeAssetWriteOnly,
				Description:      "Stringified JSON array of dependency objects. Only used for api-group classifier. Write-only.",
			},
			"original_format_version": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: diffSuppressExchangeAssetWriteOnly,
				Description:      "The version of the API spec format (e.g. '2.0' for OAS 2.0). Write-only.",
			},
			"metadata": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: diffSuppressExchangeAssetWriteOnly,
				Description:      "Stringified JSON object describing asset projectId, branchId and commitId (Design Center metadata). Write-only.",
			},
			"tags": {
				Type:             schema.TypeList,
				Optional:         true,
				ForceNew:         true,
				Elem:             &schema.Schema{Type: schema.TypeString},
				DiffSuppressFunc: diffSuppressExchangeAssetTags,
				Description:      "Tags to associate with the asset. The provider stringifies the list before sending. Write-only.",
			},
			"strict_package": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Whether the asset package is immutable (X-Strict-Package header). Write-only: defaults to false, the Exchange API never returns it.",
			},
			"allowed_api_spec_formats": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: diffSuppressExchangeAssetWriteOnly,
				Description:      "Allowed API spec formats (X-Allowed-Api-Spec-Formats header). Write-only.",
			},
			"minor_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The minor version derived by the platform.",
			},
			"version_group": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The version group derived by the platform.",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The asset type derived from the classifier.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The asset lifecycle status.",
			},
			"is_public": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the asset is published publicly.",
			},
			"is_snapshot": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the asset version is a snapshot.",
			},
			"organization_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The organization id returned by the platform.",
			},
			"created_by_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The id of the user who created the asset.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date the asset version was created.",
			},
			"contact_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The contact name of the asset owner.",
			},
			"contact_email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The contact email of the asset owner.",
			},
			"labels": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Labels attached to the asset.",
			},
			"versions": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "All versions of this asset known to the platform.",
			},
			"attributes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"key":   {Type: schema.TypeString, Computed: true},
					"value": {Type: schema.TypeString, Computed: true},
				}},
				Description: "Key/value attributes attached to the asset.",
			},
			"categories": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"key":          {Type: schema.TypeString, Computed: true},
					"display_name": {Type: schema.TypeString, Computed: true},
					"value":        {Type: schema.TypeList, Computed: true, Elem: &schema.Schema{Type: schema.TypeString}},
				}},
				Description: "Categories applied to the asset.",
			},
			"custom_fields": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"key":          {Type: schema.TypeString, Computed: true},
					"display_name": {Type: schema.TypeString, Computed: true},
					"data_type":    {Type: schema.TypeString, Computed: true},
					"text_value":   {Type: schema.TypeString, Computed: true},
					"date_value":   {Type: schema.TypeString, Computed: true},
				}},
				Description: "Custom fields populated for the asset.",
			},
			"dependencies_list": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"organization_id": {Type: schema.TypeString, Computed: true},
					"group_id":        {Type: schema.TypeString, Computed: true},
					"asset_id":        {Type: schema.TypeString, Computed: true},
					"version":         {Type: schema.TypeString, Computed: true},
				}},
				Description: "Dependencies resolved by the platform.",
			},
		},
	}
}

func resourceExchangeAssetCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	pco := m.(ProviderConfOutput)

	orgId := d.Get("org_id").(string)
	groupId := d.Get("group_id").(string)
	assetId := d.Get("asset_id").(string)
	version := d.Get("version").(string)
	classifier := d.Get("classifier").(string)

	authctx := getExchangeAssetAuthCtx(ctx, &pco)

	req := pco.exchangeassetsclient.DefaultAPI.AssetsPost(authctx).
		XStrictPackage(d.Get("strict_package").(bool)).
		OrganizationId(orgId).
		AssetId(assetId).
		Version(version).
		Name(d.Get("name").(string)).
		Classifier(classifier).
		GroupId(groupId)

	if v, ok := d.GetOk("api_version"); ok {
		req = req.ApiVersion(v.(string))
	}
	if v, ok := d.GetOk("main_file"); ok {
		req = req.Main(v.(string))
	}
	if v, ok := d.GetOk("allowed_api_spec_formats"); ok {
		req = req.XAllowedApiSpecFormats(v.(string))
	}
	if v, ok := d.GetOk("dependencies"); ok {
		req = req.Dependencies(v.(string))
	}
	if v, ok := d.GetOk("original_format_version"); ok {
		req = req.OriginalFormatVersion(v.(string))
	}
	if v, ok := d.GetOk("metadata"); ok {
		req = req.Metadata(v.(string))
	}
	if v, ok := d.GetOk("tags"); ok {
		tagsJSON, err := json.Marshal(ListInterface2ListStrings(v.([]any)))
		if err != nil {
			return diag.Diagnostics{{Severity: diag.Error, Summary: "Unable to encode tags for asset " + assetId, Detail: err.Error()}}
		}
		req = req.Tags(string(tagsJSON))
	}
	if v, ok := d.GetOk("asset_link"); ok {
		req = req.AssetLink(v.(string))
	}

	if v, ok := d.GetOk("asset_file"); ok {
		f, err := os.Open(v.(string))
		if err != nil {
			return diag.Diagnostics{{Severity: diag.Error, Summary: "Unable to open asset_file for " + assetId, Detail: err.Error()}}
		}
		defer f.Close()
		req = req.Asset(f)
	}

	_, httpr, err := req.Execute()
	if err != nil {
		return exchangeAssetHTTPDiag(httpr, err, "Unable to create exchange asset "+assetId+"/"+version)
	}
	defer httpr.Body.Close()

	d.SetId(ComposeResourceId([]string{orgId, groupId, assetId, version}))

	// POST has no description field; set it via PATCH so the first apply converges in one cycle.
	if desc, ok := d.GetOk("description"); ok {
		body := exchange_assets.NewPatchAssetNameAndDescr()
		body.SetName(d.Get("name").(string))
		body.SetDescription(desc.(string))
		patchHttpr, patchErr := pco.exchangeassetsclient.DefaultAPI.
			AssetsOrgIdAssetIdPatch(authctx, orgId, assetId).
			PatchAssetNameAndDescr(*body).Execute()
		if patchErr != nil {
			return exchangeAssetHTTPDiag(patchHttpr, patchErr, "Asset "+assetId+"/"+version+" created but description PATCH failed")
		}
		defer patchHttpr.Body.Close()
	}

	return resourceExchangeAssetRead(ctx, d, m)
}

func resourceExchangeAssetRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)

	orgId, groupId, assetId, version, err := splitExchangeAssetId(d.Id())
	if err != nil {
		return diag.Diagnostics{{Severity: diag.Error, Summary: "Unable to parse exchange asset id " + d.Id(), Detail: err.Error()}}
	}

	authctx := getExchangeAssetAuthCtx(ctx, &pco)

	asset, httpr, err := pco.exchangeassetsclient.DefaultAPI.AssetsOrgIdAssetIdVersionAssetGet(authctx, orgId, assetId, version).Execute()
	if err != nil {
		if httpr != nil && httpr.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return exchangeAssetHTTPDiag(httpr, err, "Unable to read exchange asset "+assetId+"/"+version)
	}
	defer httpr.Body.Close()

	if diags := setExchangeAssetAttrs(d, asset); diags != nil {
		return diags
	}

	classifier := classifierFromAssetAttributes(asset.GetAttributes())
	if classifier == "" {
		classifier = d.Get("classifier").(string)
	}
	d.Set("classifier", classifier)
	d.Set("strict_package", d.Get("strict_package").(bool))

	d.Set("org_id", orgId)
	d.Set("group_id", groupId)
	d.Set("asset_id", assetId)
	d.Set("version", version)

	return diags
}

func resourceExchangeAssetUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)

	if !d.HasChanges("name", "description") {
		return diags
	}

	orgId, _, assetId, version, err := splitExchangeAssetId(d.Id())
	if err != nil {
		return diag.Diagnostics{{Severity: diag.Error, Summary: "Unable to parse exchange asset id " + d.Id(), Detail: err.Error()}}
	}

	authctx := getExchangeAssetAuthCtx(ctx, &pco)

	body := exchange_assets.NewPatchAssetNameAndDescr()
	body.SetName(d.Get("name").(string))
	body.SetDescription(d.Get("description").(string))

	httpr, err := pco.exchangeassetsclient.DefaultAPI.AssetsOrgIdAssetIdPatch(authctx, orgId, assetId).
		PatchAssetNameAndDescr(*body).Execute()
	if err != nil {
		return exchangeAssetHTTPDiag(httpr, err, "Unable to update exchange asset "+assetId+"/"+version)
	}
	defer httpr.Body.Close()

	d.Set("last_updated", time.Now().Format(time.RFC850))

	return resourceExchangeAssetRead(ctx, d, m)
}

func resourceExchangeAssetDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)

	orgId, _, assetId, version, err := splitExchangeAssetId(d.Id())
	if err != nil {
		return diag.Diagnostics{{Severity: diag.Error, Summary: "Unable to parse exchange asset id " + d.Id(), Detail: err.Error()}}
	}

	authctx := getExchangeAssetAuthCtx(ctx, &pco)

	httpr, err := pco.exchangeassetsclient.DefaultAPI.AssetsOrgIdAssetIdVersionDelete(authctx, orgId, assetId, version).
		XDeleteType("hard-delete").Execute()
	if err != nil {
		if httpr != nil && httpr.StatusCode == 404 {
			d.SetId("")
			return diags
		}
		return exchangeAssetHTTPDiag(httpr, err, "Unable to delete exchange asset "+assetId+"/"+version)
	}
	defer httpr.Body.Close()

	d.SetId("")
	return diags
}

func splitExchangeAssetId(id string) (orgId, groupId, assetId, version string, err error) {
	parts := DecomposeResourceId(id)
	if len(parts) != 4 {
		return "", "", "", "", fmt.Errorf("expected org_id/group_id/asset_id/version, got: %s", id)
	}
	return parts[0], parts[1], parts[2], parts[3], nil
}

func exchangeAssetHTTPDiag(httpr *http.Response, err error, summary string) diag.Diagnostics {
	details := extractAPIErrorDetail(err, httpr)
	return diag.Diagnostics{{Severity: diag.Error, Summary: summary, Detail: details}}
}

// diffSuppressExchangeAssetWriteOnly hides post-import diffs on inputs the Exchange API
// never returns. Without suppression, declaring asset_file (etc.) in config after
// `terraform import` would force replacement on the first plan because state is empty.
// On a fresh create, d.Id() is still empty so the field flows to the Create call.
func diffSuppressExchangeAssetWriteOnly(_, old, _ string, d *schema.ResourceData) bool {
	return d.Id() != "" && old == ""
}

// diffSuppressExchangeAssetTags hides post-import diffs on the tags list. The Exchange
// API embeds user-set tags into a larger attribute set with no reliable way to separate
// them from platform-generated tags, so we cannot reconstruct the original list in Read.
// Post-import the SDK reports `tags.#` going from "0" to N — suppress it so terraform
// does not force a replacement when the imported config re-declares tags.
func diffSuppressExchangeAssetTags(k, old, _ string, d *schema.ResourceData) bool {
	if d.Id() == "" {
		return false
	}
	if k == "tags.#" {
		return old == "" || old == "0"
	}
	// suppress each element only when the list was empty server-side
	prev, ok := d.GetOk("tags")
	if !ok {
		return true
	}
	if l, ok := prev.([]any); ok && len(l) == 0 {
		return true
	}
	return false
}

func getExchangeAssetAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, exchange_assets.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, exchange_assets.ContextServerIndex, pco.server_index)
}
