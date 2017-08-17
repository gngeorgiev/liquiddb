package liquidgo

type LiquidGoConfigBuilder struct {
	autoConnect *bool
	host        *string
	port        *string
}

type LiquidGoConfig struct {
	AutoConnect bool
	Host        string
	Port        string
}

func NewConfigBuilder() *LiquidGoConfigBuilder {
	return &LiquidGoConfigBuilder{}
}

func (c *LiquidGoConfigBuilder) Host(host string) *LiquidGoConfigBuilder {
	c.host = &host
	return c
}

func (c *LiquidGoConfigBuilder) Port(port string) *LiquidGoConfigBuilder {
	c.port = &port
	return c
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

	if c.host != nil {
		config.Host = *c.host
	} else {
		config.Host = "localhost"
	}

	if c.port != nil {
		config.Port = *c.port
	} else {
		config.Port = ":8083"
	}

	return config
}
