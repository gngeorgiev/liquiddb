package liquidgo

type LiquidGoConfigBuilder struct {
	autoConnect *bool
}

func NewConfigBuilder() *LiquidGoConfigBuilder {
	return &LiquidGoConfigBuilder{}
}

func (c *LiquidGoConfigBuilder) AutoConnect(val bool) *LiquidGoConfigBuilder {
	c.autoConnect = &val
	return c
}

func (c *LiquidGoConfigBuilder) Finalize() LiquidGoConfig {
	config := LiquidGoConfig{}

	if c.autoConnect != nil {
		config.AutoConnect = *c.autoConnect
	} else {
		config.AutoConnect = true
	}

	return config
}

type LiquidGoConfig struct {
	AutoConnect bool
}
