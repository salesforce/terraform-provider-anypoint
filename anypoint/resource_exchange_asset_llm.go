package anypoint

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	exchange_assets "github.com/mulesoft-anypoint/anypoint-client-go/exchange_assets"
)

func resourceExchangeAssetLLM() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceExchangeAssetLLMCreate,
		ReadContext:   resourceExchangeAssetLLMRead,
		UpdateContext: resourceExchangeAssetLLMUpdate,
		DeleteContext: resourceExchangeAssetLLMDelete,
		Description: "Publishes and manages an Anypoint Exchange asset of type `llm` — " +
			"the metadata-only Exchange artifact required by the Anypoint AI Gateway. " +
			"Anypoint generates the LLM metadata files server-side; no spec upload is needed.",
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
			"platform": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Description: "The LLM vendor identifier. Free string — Anypoint does not enforce an enum. " +
					"Known values: `openai`, `bedrock`, `anthropic`, `azure-openai`, `gemini`, `other`. " +
					"Defaults server-side to `other` when omitted.",
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "published",
				ValidateFunc: validation.StringInSlice([]string{"published", "draft"}, false),
				Description:  "Publication status. One of `published` or `draft`. Defaults to `published`.",
			},
			"tags": {
				Type:             schema.TypeList,
				Optional:         true,
				ForceNew:         true,
				Elem:             &schema.Schema{Type: schema.TypeString},
				DiffSuppressFunc: diffSuppressExchangeAssetTags,
				Description:      "Tags to associate with the asset. The provider stringifies the list before sending. Write-only.",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Always `llm` for this resource. Returned by the platform.",
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
			"is_public": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the asset is published publicly.",
			},
			"labels": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Labels assigned by the platform.",
			},
		},
	}
}

func resourceExchangeAssetLLMCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	pco := m.(ProviderConfOutput)

	orgId := d.Get("org_id").(string)
	groupId := d.Get("group_id").(string)
	assetId := d.Get("asset_id").(string)
	version := d.Get("version").(string)

	authctx := getExchangeAssetLLMAuthCtx(ctx, &pco)

	req := pco.exchangeassetsclient.DefaultAPI.PostLLMAsset(authctx, orgId, groupId, assetId, version).
		Name(d.Get("name").(string)).
		Type_("llm").
		Status(d.Get("status").(string))

	if v, ok := d.GetOk("platform"); ok {
		req = req.PropertiesPlatform(v.(string))
	}
	if v, ok := d.GetOk("tags"); ok {
		tagsJSON, err := json.Marshal(ListInterface2ListStrings(v.([]any)))
		if err != nil {
			return diag.Diagnostics{{Severity: diag.Error, Summary: "Unable to encode tags for llm asset " + assetId, Detail: err.Error()}}
		}
		req = req.Tags(string(tagsJSON))
	}

	_, httpr, err := req.Execute()
	if err != nil {
		return exchangeAssetHTTPDiag(httpr, err, "Unable to publish llm asset "+assetId+"/"+version)
	}
	defer httpr.Body.Close()

	d.SetId(ComposeResourceId([]string{orgId, groupId, assetId, version}))

	if err := waitForLLMAssetReadable(authctx, &pco, orgId, assetId, version); err != nil {
		return diag.Diagnostics{{Severity: diag.Error, Summary: "Llm asset " + assetId + "/" + version + " published but did not become readable", Detail: err.Error()}}
	}

	if desc, ok := d.GetOk("description"); ok {
		body := exchange_assets.NewPatchAssetNameAndDescr()
		body.SetName(d.Get("name").(string))
		body.SetDescription(desc.(string))
		patchHttpr, patchErr := pco.exchangeassetsclient.DefaultAPI.
			AssetsOrgIdAssetIdPatch(authctx, orgId, assetId).
			PatchAssetNameAndDescr(*body).Execute()
		if patchErr != nil {
			return exchangeAssetHTTPDiag(patchHttpr, patchErr, "Llm asset "+assetId+"/"+version+" created but description PATCH failed")
		}
		defer patchHttpr.Body.Close()
	}

	return resourceExchangeAssetLLMRead(ctx, d, m)
}

// waitForLLMAssetReadable polls the GET endpoint after POST. Exchange publishes the
// llm asset asynchronously despite returning 202 — name/description PATCH fires too
// soon and gets 404 ASSET_NOT_FOUND. Poll with backoff up to ~30s.
func waitForLLMAssetReadable(authctx context.Context, pco *ProviderConfOutput, orgId, assetId, version string) error {
	delays := []time.Duration{500 * time.Millisecond, 1 * time.Second, 2 * time.Second, 3 * time.Second, 5 * time.Second, 8 * time.Second, 10 * time.Second}
	var lastErr error
	for _, d := range delays {
		_, httpr, err := pco.exchangeassetsclient.DefaultAPI.AssetsOrgIdAssetIdVersionAssetGet(authctx, orgId, assetId, version).Execute()
		if err == nil {
			if httpr != nil {
				httpr.Body.Close()
			}
			return nil
		}
		lastErr = err
		if httpr != nil {
			httpr.Body.Close()
		}
		time.Sleep(d)
	}
	return fmt.Errorf("llm asset not readable after polling: %w", lastErr)
}

func resourceExchangeAssetLLMRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)

	orgId, groupId, assetId, version, err := splitExchangeAssetLLMId(d.Id())
	if err != nil {
		return diag.Diagnostics{{Severity: diag.Error, Summary: "Unable to parse llm asset id " + d.Id(), Detail: err.Error()}}
	}

	authctx := getExchangeAssetLLMAuthCtx(ctx, &pco)

	asset, httpr, err := pco.exchangeassetsclient.DefaultAPI.AssetsOrgIdAssetIdVersionAssetGet(authctx, orgId, assetId, version).Execute()
	if err != nil {
		if httpr != nil && httpr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return exchangeAssetHTTPDiag(httpr, err, "Unable to read llm asset "+assetId+"/"+version)
	}
	defer httpr.Body.Close()

	if diags := setExchangeAssetLLMAttrs(d, asset); diags != nil {
		return diags
	}

	d.Set("org_id", orgId)
	d.Set("group_id", groupId)
	d.Set("asset_id", assetId)
	d.Set("version", version)

	return diags
}

func resourceExchangeAssetLLMUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)

	if !d.HasChanges("name", "description") {
		return diags
	}

	orgId, _, assetId, version, err := splitExchangeAssetLLMId(d.Id())
	if err != nil {
		return diag.Diagnostics{{Severity: diag.Error, Summary: "Unable to parse llm asset id " + d.Id(), Detail: err.Error()}}
	}

	authctx := getExchangeAssetLLMAuthCtx(ctx, &pco)

	body := exchange_assets.NewPatchAssetNameAndDescr()
	body.SetName(d.Get("name").(string))
	body.SetDescription(d.Get("description").(string))

	httpr, err := pco.exchangeassetsclient.DefaultAPI.AssetsOrgIdAssetIdPatch(authctx, orgId, assetId).
		PatchAssetNameAndDescr(*body).Execute()
	if err != nil {
		return exchangeAssetHTTPDiag(httpr, err, "Unable to update llm asset "+assetId+"/"+version)
	}
	defer httpr.Body.Close()

	d.Set("last_updated", time.Now().Format(time.RFC850))

	return resourceExchangeAssetLLMRead(ctx, d, m)
}

func resourceExchangeAssetLLMDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	var diags diag.Diagnostics
	pco := m.(ProviderConfOutput)

	orgId, _, assetId, version, err := splitExchangeAssetLLMId(d.Id())
	if err != nil {
		return diag.Diagnostics{{Severity: diag.Error, Summary: "Unable to parse llm asset id " + d.Id(), Detail: err.Error()}}
	}

	authctx := getExchangeAssetLLMAuthCtx(ctx, &pco)

	httpr, err := pco.exchangeassetsclient.DefaultAPI.AssetsOrgIdAssetIdVersionDelete(authctx, orgId, assetId, version).
		XDeleteType("hard-delete").Execute()
	if err != nil {
		if httpr != nil && httpr.StatusCode == http.StatusNotFound {
			d.SetId("")
			return diags
		}
		return exchangeAssetHTTPDiag(httpr, err, "Unable to delete llm asset "+assetId+"/"+version)
	}
	defer httpr.Body.Close()

	d.SetId("")
	return diags
}

func setExchangeAssetLLMAttrs(d *schema.ResourceData, asset *exchange_assets.Asset) diag.Diagnostics {
	d.Set("name", asset.GetName())
	d.Set("description", asset.GetDescription())
	d.Set("type", asset.GetType())
	d.Set("status", asset.GetStatus())
	d.Set("minor_version", asset.GetMinorVersion())
	d.Set("version_group", asset.GetVersionGroup())
	d.Set("organization_id", asset.GetOrganizationId())
	d.Set("created_by_id", asset.GetCreatedById())
	d.Set("created_at", asset.GetCreatedAt())
	d.Set("is_public", asset.GetIsPublic())
	d.Set("labels", asset.GetLabels())
	d.Set("platform", platformFromAssetAttributes(asset.GetAttributes()))
	return nil
}

func platformFromAssetAttributes(attrs []exchange_assets.Attribute) string {
	for _, a := range attrs {
		if a.GetKey() == "platform" {
			return a.GetValue()
		}
	}
	return ""
}

func splitExchangeAssetLLMId(id string) (orgId, groupId, assetId, version string, err error) {
	parts := DecomposeResourceId(id)
	if len(parts) != 4 {
		return "", "", "", "", fmt.Errorf("expected org_id/group_id/asset_id/version, got: %s", id)
	}
	return parts[0], parts[1], parts[2], parts[3], nil
}

func getExchangeAssetLLMAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
	tmp := context.WithValue(ctx, exchange_assets.ContextAccessToken, pco.access_token)
	return context.WithValue(tmp, exchange_assets.ContextServerIndex, pco.server_index)
}
