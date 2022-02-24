package dns

import (
	"fmt"
	"github.com/rock-go/rock/auxlib"
	"github.com/rock-go/rock/buffer"
	"github.com/rock-go/rock/lua"
	"github.com/rock-go/rock/packet"
	"github.com/rock-go/rock/pipe"
	"github.com/rock-go/rock/region"
	"gopkg.in/tomb.v2"
	"net"
	"reflect"
)

var typeof = reflect.TypeOf((*monitor)(nil)).String()

type monitor struct {
	lua.Super
	cfg  *config
	tom  *tomb.Tomb
	conn net.PacketConn
}

func newM(cfg *config) *monitor {
	m := &monitor{cfg: cfg}
	m.V(typeof, lua.INIT)
	return m
}

func (m *monitor) Name() string {
	return m.cfg.name
}

func (m *monitor) Type() string {
	return typeof
}

func (m *monitor) Region(addr net.Addr) *region.Info {
	if m.cfg.region == nil {
		return nil
	}

	ip, ver := auxlib.ParseAddr(addr)
	if ver != 4 {
		return nil
	}

	info, e := m.cfg.region.Search(ip)
	if e != nil {
		return nil
	}

	return info
}

func (m *monitor) newTx(code, host string, addr net.Addr, udp *packet.UDPHeader) *Tx {
	msg, err := packet.Dns(udp)
	if err != nil {
		xEnv.Infof("%s tx parse dns fail %v", m.Name(), err)
		return nil
	}

	return &Tx{
		msg:    msg,
		code:   code,
		host:   host,
		addr:   addr,
		name:   m.Name(),
		src:    udp.Source,
		dst:    udp.Destination,
		region: m.Region(addr),
	}
}

func (m *monitor) pipe(tx *Tx) {
	defer func() {
		if tx.buf != nil {
			buffer.Put(tx.buf)
		}
	}()

	pipe.Do(m.cfg.pipe, tx, m.cfg.co, func(err error) {
		xEnv.Errorf("%s pipe call fail %v", m.Name(), err)
	})
}

func (m *monitor) acl(udp *packet.UDPHeader) bool {
	bp := m.cfg.bind.Port()
	if bp != 0 {
		return bp == int(udp.Source)
	}
	ps := m.cfg.bind.Ports()

	n := len(ps)
	if n == 0 {
		return false
	}

	for i := 0; i < n; i++ {
		if int(udp.Source) == ps[i] {
			return true
		}
	}

	return false
}

func (m *monitor) accept() {
	buf := make([]byte, 4096)
	host := m.cfg.bind.Hostname()
	code := m.cfg.co.CodeVM()

	for {
		select {

		case <-m.tom.Dying():
			xEnv.Errorf("%s accept %v", m.tom.Err())
			return

		default:
			n, addr, err := m.conn.ReadFrom(buf)
			if err != nil {
				continue
			}

			udp := packet.NewUDPHeader(buf[:n])
			if !m.acl(udp) {
				continue
			}

			tx := m.newTx(code, host, addr, udp)
			if tx == nil {
				continue
			}
			m.pipe(tx)
		}
	}
}

func (m *monitor) Listen() error {

	conn, err := net.ListenPacket(m.cfg.net(), m.cfg.bind.Hostname())
	if err != nil {
		return err
	}

	m.conn = conn
	return nil
}

func (m *monitor) Start() error {
	if e := m.Listen(); e != nil {
		return e
	}

	m.tom = new(tomb.Tomb)
	m.tom.Go(func() error {
		m.accept()
		return nil
	})

	return nil
}

func (m *monitor) Close() error {
	m.tom.Kill(fmt.Errorf("close"))
	return m.conn.Close()
}
