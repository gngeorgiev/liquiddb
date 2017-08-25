package liquidgo

import (
	"testing"
)

func closeConn(l *LiquidGo, t *testing.T) {
	t.Helper()
	if err := l.Close(); err != nil {
		t.Fatal(err)
	}
}

func connect(t *testing.T) *LiquidGo {
	t.Helper()

	l := NewDefault()

	err := l.Connect()
	if err != nil {
		t.Fatal(err)
	}

	return l
}

func TestLiquidGo_Connect_Basic(t *testing.T) {
	l := connect(t)

	defer closeConn(l, t)
}
