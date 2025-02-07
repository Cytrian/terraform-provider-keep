package keep

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceApiKey() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceReadApiKey,
		Schema: map[string]*schema.Schema{
			"reference_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Reference of the ApiKey",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation time of the ApiKey",
			},
			"created_by": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creator of the ApiKey",
			},
			"secret": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Secret of the ApiKey",
			},
			"role": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Secret of the ApiKey",
			},
			"is_deleted": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Deletion status of the ApiKey",
			},
		},
	}
}

func dataSourceReadApiKey(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	name := d.Get("name").(string)

	apikeys, errResp, err := client.GetApiKeys()

	if err != nil {
		if errResp != nil {
			return diag.Errorf("API Error: %s. Details: %s", errResp.Error, errResp.Details)
		}
		return diag.Errorf("error reading apikeys: %s", err)
	}

	for _, apikeyMap := range apikeys {
		apikey := apikeyMap.(map[string]interface{})
		tflog.Debug(ctx, "GetApiKeys", map[string]interface{}{
			"apikey compared": apikey,
		})
		if apikey["reference_id"].(string) == name {
			d.SetId(apikey["reference_id"].(string))
			d.Set("name", apikey["reference_id"])
			d.Set("reference_id", apikey["reference_id"])
			d.Set("created_at", apikey["created_at"])
			d.Set("created_by", apikey["created_by"])
			d.Set("secret", apikey["secret"])
			d.Set("is_deleted", apikey["is_deleted"])
			return nil
		}
	}

	return diag.Errorf("API key with reference ID %s not found", name)
}
