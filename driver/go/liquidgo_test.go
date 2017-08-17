package liquidgo

import (
	"testing"
)

func closeConn(l *LiquidGo, t *testing.T) {
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestLiquidGo_Connect_Basic(t *testing.T) {
	l := NewDefault()

	err := l.Connect()
	if err != nil {
		t.Fatal(err)
	}

	defer closeConn(l, t)
}
