package liquidgo

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	"time"

	"github.com/gngeorgiev/liquiddb"
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
	connected    chan struct{}

	connMutex sync.Mutex
	conn      net.Conn

	errCh chan error

	dataChannelsMutex sync.Mutex
	dataChannels      []chan liquiddb.EventData
}

func New(config LiquidGoConfig) *LiquidGo {
	l := &LiquidGo{
		config:       config,
		status:       StatusDisconnected,
		disconnected: make(chan bool, 0),

		conn: nil,

		errCh: make(chan error),

		dataChannelsMutex: sync.Mutex{},
		dataChannels:      make([]chan liquiddb.EventData, 0),
	}

	return l
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

		go l.waitError()
		go l.read()

		return nil
	}

	err := connect()
	if l.config.AutoConnect {
		go func() {
			for {
				reconnect := <-l.disconnected
				l.connMutex.Lock()
				l.conn = nil
				l.connMutex.Unlock()
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
	return newRef(l, path)
}

func (l *LiquidGo) RefFromArray(path []string) *Ref {
	return l.Ref(strings.Join(path, "."))
}

func (l *LiquidGo) Write(data ClientData) error {
	b, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if err := l.conn.SetWriteDeadline(time.Now().Add(time.Second * 1)); err != nil {
		return err
	}

	_, writeErr := l.conn.Write(b)
	return writeErr
}

func (l *LiquidGo) Read(ch chan liquiddb.EventData) {
	l.dataChannelsMutex.Lock()
	defer l.dataChannelsMutex.Unlock()

	l.dataChannels = append(l.dataChannels, ch)
}

func (l *LiquidGo) ReadId(id uint64) liquiddb.EventData {
	ch := make(chan liquiddb.EventData)
	l.Read(ch)
	defer l.StopRead(ch)

	for d := range ch {
		if d.ID == id {
			return d
		}
	}

	return liquiddb.EventData{}
}

func (l *LiquidGo) StopRead(ch chan liquiddb.EventData) {
	l.dataChannelsMutex.Lock()
	defer l.dataChannelsMutex.Unlock()

	for i, dataCh := range l.dataChannels {
		if dataCh == ch {
			l.dataChannels = append(l.dataChannels[:i], l.dataChannels[i+1:]...)
			break
		}
	}
}

func (l *LiquidGo) waitError() {
	err := <-l.errCh
	log.Println(err)
	l.disconnected <- true
}

func (l *LiquidGo) read() {
	data := make([]byte, 0, 4096)
	buf := make([]byte, 0, 1024)

	for {
		n, err := l.conn.Read(buf)
		if n > 0 {
			data = append(data, buf[:n]...)
		}

		if err != nil {
			l.errCh <- err
			break
		}

		var eventData liquiddb.EventData
		if err := json.Unmarshal(data, eventData); err != nil {
			l.errCh <- err
			break
		}

		l.dataChannelsMutex.Lock()
		for _, ch := range l.dataChannels {
			select {
			case ch <- eventData:
			default:
			}
		}
		l.dataChannelsMutex.Unlock()
	}
}
