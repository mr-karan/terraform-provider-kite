package kite

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kiteconnect "github.com/zerodhatech/gokiteconnect"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("KITE_API_KEY", nil),
			},
			"api_secret": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("KITE_API_SECRET", nil),
			},
			"request_token": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("KITE_API_REQUEST_TOKEN", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"kite_holding": resourceHolding(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	apiKey := d.Get("api_key").(string)
	apiSecret := d.Get("api_secret").(string)
	requestToken := d.Get("request_token").(string)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	// Create a new Kite connect instance
	kc := kiteconnect.New(apiKey)

	var accessToken string
	// Check if access token exists.
	data, err := ioutil.ReadFile(".tf-kite-secret")
	// If it doesn't exist, proceed with login flow.
	if err != nil {
		// Get user details and access token
		data, err := kc.GenerateSession(requestToken, apiSecret)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create Kite client",
				Detail:   "Unable to authenticate user for Kite API",
			})

			return nil, diags
		}
		file, err := os.Create(".tf-kite-secret")

		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to save access token",
				Detail:   "Unable to save access token from Kite API for future use",
			})
			return nil, diags
		}
		defer file.Close()
		_, err = file.WriteString(data.AccessToken)

		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to save access token",
				Detail:   "Unable to save access token from Kite API for future use",
			})
			return nil, diags
		}
		accessToken = data.AccessToken
	} else {
		// check if accessToken is valid
		accessToken = string(data)
	}
	// Set access token
	kc.SetAccessToken(accessToken)

	// check if access token is valid
	_, err = kc.GetUserProfile()
	if err != nil {
		// Get user details and access token
		data, err := kc.GenerateSession(requestToken, apiSecret)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create Kite client",
				Detail:   "Unable to authenticate user for Kite API",
			})

			return nil, diags
		}
		file, err := os.Create(".tf-kite-secret")

		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to save access token",
				Detail:   "Unable to save access token from Kite API for future use",
			})
			return nil, diags
		}
		defer file.Close()
		_, err = file.WriteString(data.AccessToken)

		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to save access token",
				Detail:   "Unable to save access token from Kite API for future use",
			})
			return nil, diags
		}
		accessToken = data.AccessToken
		// Set access token
		kc.SetAccessToken(accessToken)
	}
	return kc, diags
}
