package keep

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceApiKey() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceCreateApiKey,
		ReadContext:   resourceReadApiKey,
		UpdateContext: resourceUpdateApiKey,
		DeleteContext: resourceDeleteApiKey,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the ApiKey",
			},
			"reference_id": {
				Type:        schema.TypeString,
				Computed:    true,
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
				Required:    true,
				Description: "Role of the ApiKey",
			},
			"is_deleted": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Deletion status of the ApiKey",
			},
		},
	}
}

func resourceCreateApiKey(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	name := d.Get("name").(string)
	role := d.Get("role").(string)

	// Prepare installation payload
	createPayload := map[string]interface{}{
		"name":               name,
		"role":               role,
		"system_description": "Hanswurst",
	}

	// Create API Key
	apiKey, errResp, err := client.CreateApiKey(createPayload)
	if err != nil {
		if errResp != nil {
			return diag.Errorf("API Error creating API key: %s. Details: %s", errResp.Error, errResp.Details)
		}
		return diag.Errorf("error creating API key: %s", err)
	}
	tflog.Debug(ctx, "Create", map[string]interface{}{
		"apikey created": apiKey,
	})
	d.SetId(apiKey["reference_id"].(string))
	d.Set("reference_id", apiKey["reference_id"].(string))
	d.Set("secret", apiKey["secret"].(string))
	d.Set("role", apiKey["role"].(string))

	// Update resource state
	return resourceReadApiKey(ctx, d, m)
}

func resourceReadApiKey(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	referenceID := d.Get("reference_id").(string)

	// Fetch API keys
	apikeys, errResp, err := client.GetApiKeys()
	if err != nil {
		if errResp != nil {
			return diag.Errorf("API Error: %s. Details: %s", errResp.Error, errResp.Details)
		}
		return diag.Errorf("error reading API keys: %s", err)
	}

	// Find the specific API key
	for _, apikeyMap := range apikeys {
		apikey := apikeyMap.(map[string]interface{})
		tflog.Debug(ctx, "GetApiKeys", map[string]interface{}{
			"apikey compared": apikey,
		})

		if apikey["reference_id"].(string) == referenceID {
			// Update resource attributes
			d.Set("reference_id", apikey["reference_id"])
			d.Set("created_at", apikey["created_at"])
			d.Set("created_by", apikey["created_by"])
			d.Set("secret", apikey["secret"])
			d.Set("is_deleted", apikey["is_deleted"])
			return nil
		}
	}

	// If API key not found, remove from state
	d.SetId("")
	return diag.Errorf("API key with reference ID %s not found", referenceID)
}

func resourceUpdateApiKey(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Check if reference_id has changed
	if d.HasChange("reference_id") {
		return diag.Errorf("cannot update API key reference ID")
	}

	// Refresh the state by calling read
	return resourceReadApiKey(ctx, d, m)
}

func resourceDeleteApiKey(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)
	referenceID := d.Get("reference_id").(string)

	// Delete API Key
	errResp, err := client.DeleteApiKey(referenceID)
	if err != nil {
		if errResp != nil {
			return diag.Errorf("API Error deleting API key: %s. Details: %s", errResp.Error, errResp.Details)
		}
		return diag.Errorf("error deleting API key: %s", err)
	}

	// Remove from state
	d.SetId("")
	return nil
}
