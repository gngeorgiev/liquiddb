package liquidgo

type LiquidGoStatus string

const (
	StatusConnected    = LiquidGoStatus("connected")
	StatusDisconnected = LiquidGoStatus("disconnected")
)
