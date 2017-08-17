package liquidgo

import (
	"fmt"
	"net"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

type LiquidGo struct {
	config       LiquidGoConfig
	status       LiquidGoStatus
	disconnected chan bool
	conn         net.Conn
}

func New(config LiquidGoConfig) *LiquidGo {
	return &LiquidGo{
		config:       config,
		status:       StatusDisconnected,
		disconnected: make(chan bool, 0),
		conn:         nil,
	}
}

func NewDefault() *LiquidGo {
	return New(NewConfigBuilder().Finalize())
}

func (l *LiquidGo) Status() LiquidGoStatus {
	return l.status
}

func (l *LiquidGo) Connect() error {
	connect := func() error {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s%s", l.config.Host, l.config.Port))
		if err != nil {
			return err
		}

		l.conn = conn
		return nil
	}

	err := connect()
	if l.config.AutoConnect {
		go func() {
			for {
				reconnect := <-l.disconnected
				if !reconnect {
					break
				}

				l.status = StatusDisconnected
				go func() {
					err := connect()
					if err != nil {
						log.WithField("type", "reconnect").Error(err)
						l.disconnected <- true
					} else {
						log.WithField("type", "reconnect").Println("Success")
					}
				}()
			}
		}()
	}

	if err != nil {
		log.WithField("type", "initial connect").Error(err)
		return err
	}

	log.WithField("type", "initial connect").Println("Success")
	return nil
}

func (l *LiquidGo) Close() error {
	l.disconnected <- false

	return l.conn.Close()
}

func (l *LiquidGo) Ref(path string) *Ref {
	return &Ref{path}
}

func (l *LiquidGo) RefFromArray(path []string) *Ref {
	return l.Ref(strings.Join(path, "."))
}
