package liquidgo

type LiquidGo struct {
	config    LiquidGoConfig
	connected bool
}

func New(config LiquidGoConfig) *LiquidGo {
	return &LiquidGo{
		config:    config,
		connected: false,
	}
}

func (l *LiquidGo) Connect() error {
	return nil
}

func (l *LiquidGo) Ref(path string) *Ref {
	return &Ref{path}
}
