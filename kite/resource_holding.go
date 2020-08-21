package kite

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	kiteconnect "github.com/zerodhatech/gokiteconnect"
)

func resourceHolding() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceHoldingCreateorUpdate,
		ReadContext:   resourceHoldingRead,
		UpdateContext: resourceHoldingCreateorUpdate,
		DeleteContext: resourceHoldingDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"tradingsymbol": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"order_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"exchange": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "BSE",
			},
			"quantity": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func resourceHoldingCreateorUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	kc := m.(*kiteconnect.Client)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	tradingsymbol := d.Get("tradingsymbol").(string)
	exchange := d.Get("exchange").(string)
	desiredQuantity := d.Get("quantity").(int)

	// compute quantity and transaction type
	transactionType, computedQuantity, err := computeTransation(kc, desiredQuantity, tradingsymbol)
	if err != nil {
		return diag.FromErr(err)
	}
	// manual check for 0 quantity
	if computedQuantity == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Desired quantity and Real quantity is same for %s", tradingsymbol),
			Detail:   fmt.Sprintf("Please import the holdings with terraform import kite_holding.$name %s", tradingsymbol),
		})
		return diags
	}
	// place an order.
	resp, err := sendOrder(kc, exchange, tradingsymbol, computedQuantity, transactionType)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to place order for %s", tradingsymbol),
			Detail:   fmt.Sprintf("Error while placing order for %s : %s", tradingsymbol, err),
		})
		return diags
	}

	// check for order rejections
	diags = checkforOrderIssues(kc, resp.OrderID)
	if len(diags) != 0 {
		return diags
	}

	// if no rejections then update the state file.
	d.SetId(tradingsymbol)
	d.Set("order_id", resp.OrderID)

	return diags
}

func resourceHoldingRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var (
		diags         diag.Diagnostics
		tradingsymbol = d.Id()
		quantity      int
		kc            = m.(*kiteconnect.Client)
	)
	// Fetch user holdings.
	holdings, err := kc.GetHoldings()
	if err != nil {
		return diag.FromErr(err)
	}
	for _, h := range holdings {
		// Filter holdings with the tradingsymbol.
		if h.Tradingsymbol == tradingsymbol {
			d.Set("tradingsymbol", h.Tradingsymbol)
			d.Set("exchange", h.Exchange)
			quantity += h.Quantity
		}
	}
	// Fetch user positions.
	positions, err := kc.GetPositions()
	if err != nil {
		return diag.FromErr(err)
	}
	for _, p := range positions.Net {
		// Filter holdings with the tradingsymbol.
		if p.Tradingsymbol == tradingsymbol {
			quantity += p.Quantity
		}
	}
	d.Set("quantity", quantity)

	return diags
}

func resourceHoldingDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	kc := m.(*kiteconnect.Client)
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	tradingsymbol := d.Get("tradingsymbol").(string)
	exchange := d.Get("exchange").(string)
	quantity := d.Get("quantity").(int)
	// Place a new SELL order
	_, err := sendOrder(kc, exchange, tradingsymbol, quantity, "SELL")
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to place order for %s", tradingsymbol),
			Detail:   fmt.Sprintf("Error while placing order for %s : %s", tradingsymbol, err),
		})
		return diags
	}
	return diags
}

// computeTransaction takes in the desiredQuantity for a particular stock and checks
// with the real world quantity which is decided after summing up the positions and holdings
// for that stock. The positions contain the net quantity bought/sold for that day and holdings
// contain the net total including T+1.
// If real world quantity is more than desired quantity then
// it is assumed that user wants to sell the difference. If desired quantity is more
// than real world quantity then it is assumed that user wants to buy the difference.
func computeTransation(kc *kiteconnect.Client, desiredQuantity int, tradingsymbol string) (string, int, error) {
	holdings, err := kc.GetHoldings()
	if err != nil {
		return "", 0, err
	}
	var actualQuantity int
	for _, h := range holdings {
		if h.Tradingsymbol == tradingsymbol {
			actualQuantity += h.Quantity
		}
	}
	positions, err := kc.GetPositions()
	if err != nil {
		return "", 0, err
	}
	for _, p := range positions.Net {
		if p.Tradingsymbol == tradingsymbol {
			actualQuantity += p.Quantity
		}
	}

	transactionType := "BUY"
	computedQuantity := desiredQuantity - actualQuantity

	if actualQuantity > desiredQuantity {
		transactionType = "SELL"
		computedQuantity = actualQuantity - desiredQuantity
	}

	return transactionType, computedQuantity, nil
}

// sendOrder is a small wrapper on `PlaceOrder` API calls. It constructs the order params
// with a few hardcoded parameters.
func sendOrder(kc *kiteconnect.Client, exchange string, tradingsymbol string, quantity int, transactionType string) (kiteconnect.OrderResponse, error) {
	orderParams := kiteconnect.OrderParams{
		Exchange:        exchange,
		Tradingsymbol:   tradingsymbol,
		Quantity:        quantity,
		TransactionType: transactionType,
		Product:         "CNC",
		OrderType:       "MARKET",
		Validity:        "DAY",
	}

	return kc.PlaceOrder("regular", orderParams)
}

// Since the fate of order depends on external market conditions, we need to check
// for the order status "asynchronously". Having Postbacks would be ideal but running
// a web server here was complicated for the use case, so as a workaround, the program
// halts for 10 seconds (more than enough for liquid contracts) and checks for incomplete
// or rejected orders. In case the order is not fully filled or rejected and error is thrown
// to user and `apply` call is marked with `ERROR`.
func checkforOrderIssues(kc *kiteconnect.Client, orderID string) diag.Diagnostics {
	// after placing order, wait for 10s to update the status.
	time.Sleep(10 * time.Second)
	var diags diag.Diagnostics

	// Get the order details for this orderID
	orders, err := kc.GetOrderHistory(orderID)
	if err != nil {
		return diag.FromErr(err)
	}
	// The last order entry would be the latest one and
	// can be used to check the status and pending quantity.
	order := orders[len(orders)-1]
	if order.Status != "COMPLETE" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("Unable to place %s order for %s", order.TransactionType, order.TradingSymbol),
			Detail:   fmt.Sprintf("%s - %s", order.Status, order.StatusMessage),
		})
		return diags
	}
	if order.PendingQuantity != 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  fmt.Sprintf("%s Order for %s is partially filled. Please check https://kite.zerodha.com for more information.", order.TransactionType, order.TradingSymbol),
			Detail:   fmt.Sprintf("%f/%f order filled.", order.FilledQuantity, order.Quantity),
		})
		return diags
	}
	return diags
}
