package proxmox

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/Telmate/proxmox-api-go/internal/generic"
	"github.com/Telmate/proxmox-api-go/internal/noCopy"
	"github.com/Telmate/proxmox-api-go/internal/notify"
	"github.com/coder/websocket"
)

type term struct {
	Port   string `json:"port"`
	Ticket string `json:"ticket"`
	UpID   string `json:"upid"`
	User   string `json:"user"`
}

func newTerminalSession(ctx context.Context, c *clientAPI, raw *term, node string, guestType, guestID string) (*terminalSession, error) {
	conn, err := c.session.webSocket(ctx, "/nodes/"+node+"/"+guestType+"/"+guestID+"/vncwebsocket?port="+raw.Port+"&vncticket="+url.QueryEscape(raw.Ticket), raw.User, raw.Ticket)
	if err != nil {
		return nil, err
	}

	// Setup sync machinery

	closeCH := notify.New()
	readCH := make(chan []byte)
	newCtx, cancel := context.WithCancel(ctx)

	session := &terminalSession{
		client:  c,
		closeCH: closeCH,
		conn:    conn,
		ctx:     ctx,
		err:     generic.Once[error]{},
		node:    node,
		readCH:  readCH,
		upID:    raw.UpID}

	session.closer = func() {
		var once sync.Once
		once.Do(func() {
			closeCH.Close()
			conn.CloseNow()
			cancel() // for some reason closing the connection doesn't cancel active read and writes.
		})
	}

	go func() { // Reader
		const beginOfStream = "\x1b[H\x1b[J" // Not fully sure what this code is exactly.
		defer close(readCH)
		for {
			select {
			case <-closeCH.Done():
				return
			default:
				_, msg, err := conn.Read(newCtx)
				if err != nil {
					newErr := deepUnwrap(err)
					switch newErr.Error() {
					case "EOF":
						continue
					}
					select {
					case <-closeCH.Done(): // When we got closed by something, discard error and don't close again.
					default:
						session.err.Set(err)
						session.closer()
					}
					return
				}
				if string(msg) == beginOfStream {
					continue
				}
			chanLoop:
				for {
					select {
					case readCH <- msg:
						break chanLoop
					case <-closeCH.Done():
						return
					default:
						time.After(time.Millisecond * 10)
					}
				}
			}
		}
	}()
	return session, nil
}

type TerminalSession interface {
	Close() error
	Error() error
	Write([]byte) error
	Reader() (chan []byte, error)
}

var _ TerminalSession = (*terminalSession)(nil)

type terminalSession struct {
	_       noCopy.True
	ctx     context.Context
	client  *clientAPI
	closeCH *notify.Channel
	closer  func()
	conn    *websocket.Conn
	readCH  chan []byte
	node    string
	upID    string
	err     generic.Once[error]
	reader  sync.Once
}

func (t *terminalSession) Close() error {
	select {
	case <-t.closeCH.Done():
	default:
		t.closer()
		t.err.Set(taskCancel(t.ctx, t.client, t.node, t.upID))
	}
	return t.err.Load()
}

func (t *terminalSession) Error() error { return t.err.Load() }

func (t *terminalSession) Write(b []byte) error {
	select {
	case <-t.closeCH.Done():
	default:
		if err := t.conn.Write(t.ctx, websocket.MessageBinary, append([]byte("0:"+strconv.Itoa(len(b))+":"), b...)); err != nil {
			t.err.Set(err)
			t.closer()
		}
	}
	return t.err.Load()
}

func (t *terminalSession) Reader() (chan []byte, error) {
	var returnReader bool
	t.reader.Do(func() {
		returnReader = true
	})
	if returnReader {
		return t.readCH, nil
	}
	return nil, errors.New("reader can only be instantiated once")
}
