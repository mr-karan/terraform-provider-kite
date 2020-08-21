# Terraform Provider Kite
_Terraform plugin for managing stock portfolio with Zerodha Kite_

[![asciicast](https://asciinema.org/a/354736.png)](https://asciinema.org/a/354736)

## Why

This provider lets you declaratively manage your long term holdings with Terraform and Zerodha Kite API. You can model your entire portfolio as Terraform "resource" blocks and let it handle all the future updates to your portfolio. With Terraform you get to see how your portfolio evolved over time with point-in-time snapshots of your portfolio, you can do multiple changes to your portfolio with one click, you can easily extend it to make your own SIP scheduler with a simple cron job -  Possibilities are endless.

### Seriously why?

Well, honestly this is a fun project aimed at learning to build a Terraform Plugin. The [Plugin SDK](https://www.terraform.io/docs/extend/plugin-sdk.html) architecture is quite nice and I wanted to dig in some internals of how Terraform Provider works. This project helped me achieve that goal.

## Installing

Download `bin/terraform-provider-kite` and place it on your machine at `~/.terraform.d/plugins/mrkaran.dev/github/kite/`. Follow the [official instructions](https://www.terraform.io/docs/configuration/providers.html#third-party-plugins) for more information.

```sh
```

You can view [examples/versions.tf](examples/versions.tf) for the sample provider configuration.

## Example Usage

```hcl
# Buy 30 stocks of NSE:CASTROLIND
resource "kite_holding" "castrol" {
  tradingsymbol = "CASTROLIND"
  quantity      = 30
  exchange      = "NSE"
}

# Show the order id.
output "castrol_order_id" {
  value = kite_holding.castrol.order_id
}
```

## Argument Reference

The following arguments are supported:

- `tradingsymbol` - (Required, string) Tradingsymbol of the instrument.
- `quantity` - (Required, int) Quantity to transact.
- `exchange` - (Optional, string) Name of the exchange. Must be one of "NSE","BSE". Defaults to "BSE".


## Attributes Reference

The following attributes are exported:

- `quantity` - (int) Quantity held by the user. Includes the quantity in postion and holdings.
- `order_id` - (string) Unique order ID.

## Import

It is possible to import your existing holdings via `terraform import` command. This helps you to make your existing state file aware about the holdings bought without Terraform.

For eg, to import `INFY` from your holdings:

`terraform import kite_holding.infy INFY`

### Configuring Provider

The provider requires the following environment variables to authenticate with Kite API:

- `KITE_API_KEY=<>`
- `KITE_API_SECRET=<>`
- `KITE_API_REQUEST_TOKEN=<>`

You can generate these credentials using [Kite Developer Console](https://developers.kite.trade/).
To get `KITE_API_REQUEST_TOKEN` you need to visit the public Kite login endpoint at [https://kite.zerodha.com/connect/login?v=3&api_key=xxx](https://kite.zerodha.com/connect/login?v=3&api_key=xxx). You will get `request_token` as a URL parameter to the redirect URL registered for your app. This is a one time token and will expire after first succesful login attempt.

Please read point `3)` in [Warnings and Caveats](#warnings-and-caveats) section to know how the access token is persisted across multiple runs of `terraform`.

To initialise the provider:

```hcl
provider "kite" {
}
```

## Warnings and Caveats

- The author of this program is currently employed by Zerodha but this software isn't assosciated with Zerodha in any manner. This is a completely open and FOSS project.

- If you use this program and lose money, don't blame me. This software comes with absolutely no 
guarantees.

- Due to Indian Exchange regulations and guidelines the user(s) are expected to login to the trading platform every day before placing trades. To comply with that, the login cannot be automated. Since Terraform needs to call Kite API to get the latest state and modify the state, any kind of API call needs an access token. It is not possible to persist the Access Token across Terraform runs so this program persists it on the user's local path `.tf-kite-secret`. Future versions of this program will make the path to this file as a provider config. You are expected to keep this file private and not keep this open to shared environments. To repeat point `2)`, if you lose money while using/due to this program, don't blame me.

### Develop

#### Test sample configuration

First, build and install the provider.

```shell
$ make install
```

Then, navigate to the `examples` directory. 

```shell
$ cd examples
```

Run the following command to initialize the workspace and apply the sample configuration.

```shell
$ terraform init && terraform apply
```

## Credits

- [Kite Connect API](https://kite.trade/docs/connect/)
- [Terraform Plugin SDK](https://www.terraform.io/docs/extend/plugin-sdk.html)

## LICENSE

See [LICENSE](LICENSE) for more details.

## ⭐️ Show your support

Give a ⭐️ if this project helped you!

## Contributing

This is still an alpha release. For a full list of things to improve, see unchecked items in [TODO](TODO.md).
Contributions welcome!
