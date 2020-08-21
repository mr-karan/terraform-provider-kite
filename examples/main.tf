provider "kite" {
}

resource "kite_holding" "castrol" {
  tradingsymbol = "CASTROLIND"
  quantity      = 30
  exchange      = "NSE"
}

output "castrol_order_id" {
  value = kite_holding.castrol.order_id
}
