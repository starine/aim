package container

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/starine/aim"
	"github.com/starine/aim/logger"
	"github.com/starine/aim/naming"
	"github.com/starine/aim/tcp"
	"github.com/starine/aim/wire"
	"github.com/starine/aim/wire/pkt"
)

const (
	stateUninitialized = iota
	stateInitialized
	stateStarted
	stateClosed
)

const (
	StateYoung = "young"
	StateAdult = "adult"
)

const (
	KeyServiceState = "service_state"
)

// Container Container
type Container struct {
	sync.RWMutex
	Naming     naming.Naming
	Srv        aim.Server
	state      uint32
	srvclients map[string]ClientMap
	selector   Selector
	dialer     aim.Dialer
	deps       map[string]struct{}
	monitor    sync.Once
}

var log = logger.WithField("module", "container")

// Default Container
var c = &Container{
	state:    0,
	selector: &HashSelector{},
	deps:     make(map[string]struct{}),
}

// Default Default
func Default() *Container {
	return c
}

// Init Init
func Init(srv aim.Server, deps ...string) error {
	if !atomic.CompareAndSwapUint32(&c.state, stateUninitialized, stateInitialized) {
		return errors.New("has Initialized")
	}
	c.Srv = srv
	for _, dep := range deps {
		if _, ok := c.deps[dep]; ok {
			continue
		}
		c.deps[dep] = struct{}{}
	}
	log.WithField("func", "Init").Infof("srv %s:%s - deps %v", srv.ServiceID(), srv.ServiceName(), c.deps)
	c.srvclients = make(map[string]ClientMap, len(deps))
	return nil
}

// SetDialer set tcp dialer
func SetDialer(dialer aim.Dialer) {
	c.dialer = dialer
}

// EnableMonitor start
func EnableMonitor(listen string) {
	c.monitor.Do(func() {
		go func() {
			http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("ok"))
			})
			// add prometheus metrics
			http.Handle("/metrics", promhttp.Handler())
			_ = http.ListenAndServe(listen, nil)
		}()
	})
}

// SetSelector set a default selector
func SetSelector(selector Selector) {
	c.selector = selector
}

// SetServiceNaming
func SetServiceNaming(nm naming.Naming) {
	c.Naming = nm
}

// Start server
func Start() error {
	if c.Naming == nil {
		return fmt.Errorf("naming is nil")
	}

	if !atomic.CompareAndSwapUint32(&c.state, stateInitialized, stateStarted) {
		return errors.New("has started")
	}

	// 1. ??????Server
	go func(srv aim.Server) {
		err := srv.Start()
		if err != nil {
			log.Errorln(err)
		}
	}(c.Srv)

	// 2. ??????????????????????????????
	for service := range c.deps {
		go func(service string) {
			err := connectToService(service)
			if err != nil {
				log.Errorln(err)
			}
		}(service)
	}

	//3. ????????????
	if c.Srv.PublicAddress() != "" && c.Srv.PublicPort() != 0 {
		err := c.Naming.Register(c.Srv)
		if err != nil {
			log.Errorln(err)
		}
	}

	// wait quit signal of system
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	log.Infoln("shutdown", <-c)
	// 4. ??????
	return shutdown()
}

// push message to server
func Push(server string, p *pkt.LogicPkt) error {
	p.AddStringMeta(wire.MetaDestServer, server)
	return c.Srv.Push(server, pkt.Marshal(p))
}

// Forward message to service
func Forward(serviceName string, packet *pkt.LogicPkt) error {
	if packet == nil {
		return errors.New("packet is nil")
	}
	if packet.Command == "" {
		return errors.New("command is empty in packet")
	}
	if packet.ChannelId == "" {
		return errors.New("ChannelId is empty in packet")
	}
	return ForwardWithSelector(serviceName, packet, c.selector)
}

// ForwardWithSelector forward data to the specified node of service which is chosen by selector
func ForwardWithSelector(serviceName string, packet *pkt.LogicPkt, selector Selector) error {
	cli, err := lookup(serviceName, &packet.Header, selector)
	if err != nil {
		return err
	}
	// add a tag in packet
	packet.AddStringMeta(wire.MetaDestServer, c.Srv.ServiceID())
	log.Debugf("forward message to %v with %s", cli.ServiceID(), &packet.Header)
	return cli.Send(pkt.Marshal(packet))
}

// shutdown Shutdown
func shutdown() error {
	if !atomic.CompareAndSwapUint32(&c.state, stateStarted, stateClosed) {
		return errors.New("has closed")
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()
	// 1. ?????????????????????
	err := c.Srv.Shutdown(ctx)
	if err != nil {
		log.Error(err)
	}
	// 2. ???????????????????????????
	err = c.Naming.Deregister(c.Srv.ServiceID())
	if err != nil {
		log.Warn(err)
	}
	// 3. ??????????????????
	for dep := range c.deps {
		_ = c.Naming.Unsubscribe(dep)
	}

	log.Infoln("shutdown")
	return nil
}

func lookup(serviceName string, header *pkt.Header, selector Selector) (aim.Client, error) {
	clients, ok := c.srvclients[serviceName]
	if !ok {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}
	// ??????????????????StateAdult?????????
	srvs := clients.Services(KeyServiceState, StateAdult)
	if len(srvs) == 0 {
		return nil, fmt.Errorf("no services found for %s", serviceName)
	}
	id := selector.Lookup(header, srvs)
	if cli, ok := clients.Get(id); ok {
		return cli, nil
	}
	return nil, fmt.Errorf("no client found")
}

func connectToService(serviceName string) error {
	clients := NewClients(10)
	c.srvclients[serviceName] = clients
	// 1. ??????Watch???????????????
	delay := time.Second * 10
	err := c.Naming.Subscribe(serviceName, func(services []aim.ServiceRegistration) {
		for _, service := range services {
			if _, ok := clients.Get(service.ServiceID()); ok {
				continue
			}
			log.WithField("func", "connectToService").Infof("Watch a new service: %v", service)
			service.GetMeta()[KeyServiceState] = StateYoung
			go func(service aim.ServiceRegistration) {
				time.Sleep(delay)
				service.GetMeta()[KeyServiceState] = StateAdult
			}(service)

			_, err := buildClient(clients, service)
			if err != nil {
				logger.Warn(err)
			}
		}
	})
	if err != nil {
		return err
	}
	// 2. ??????????????????????????????
	services, err := c.Naming.Find(serviceName)
	if err != nil {
		return err
	}
	log.Info("find service ", services)
	for _, service := range services {
		// ?????????StateAdult
		service.GetMeta()[KeyServiceState] = StateAdult
		_, err := buildClient(clients, service)
		if err != nil {
			logger.Warn(err)
		}
	}
	return nil
}

func buildClient(clients ClientMap, service aim.ServiceRegistration) (aim.Client, error) {
	c.Lock()
	defer c.Unlock()
	var (
		id   = service.ServiceID()
		name = service.ServiceName()
		meta = service.GetMeta()
	)
	// 1. ??????????????????????????????
	if _, ok := clients.Get(id); ok {
		return nil, nil
	}
	// 2. ???????????????????????????tcp??????
	if service.GetProtocol() != string(wire.ProtocolTCP) {
		return nil, fmt.Errorf("unexpected service Protocol: %s", service.GetProtocol())
	}

	// 3. ??????????????????????????????
	cli := tcp.NewClientWithProps(id, name, meta, tcp.ClientOptions{
		Heartbeat: aim.DefaultHeartbeat,
		ReadWait:  aim.DefaultReadWait,
		WriteWait: aim.DefaultWriteWait,
	})
	if c.dialer == nil {
		return nil, fmt.Errorf("dialer is nil")
	}
	cli.SetDialer(c.dialer)
	err := cli.Connect(service.DialURL())
	if err != nil {
		return nil, err
	}
	// 4. ????????????
	go func(cli aim.Client) {
		err := readLoop(cli)
		if err != nil {
			log.Debug(err)
		}
		clients.Remove(id)
		cli.Close()
	}(cli)
	// 5. ???????????????????????????
	clients.Add(cli)
	return cli, nil
}

// Receive default listener
func readLoop(cli aim.Client) error {
	log := logger.WithFields(logger.Fields{
		"module": "container",
		"func":   "readLoop",
	})
	log.Infof("readLoop started of %s %s", cli.ServiceID(), cli.ServiceName())
	for {
		frame, err := cli.Read()
		if err != nil {
			log.Trace(err)
			return err
		}
		if frame.GetOpCode() != aim.OpBinary {
			continue
		}
		buf := bytes.NewBuffer(frame.GetPayload())

		packet, err := pkt.MustReadLogicPkt(buf)
		if err != nil {
			log.Info(err)
			continue
		}
		err = pushMessage(packet)
		if err != nil {
			log.Info(err)
		}
	}
}

// ????????????????????????????????????channel???
func pushMessage(packet *pkt.LogicPkt) error {
	server, _ := packet.GetMeta(wire.MetaDestServer)
	if server != c.Srv.ServiceID() {
		return fmt.Errorf("dest_server is incorrect, %s != %s", server, c.Srv.ServiceID())
	}
	channels, ok := packet.GetMeta(wire.MetaDestChannels)
	if !ok {
		return fmt.Errorf("dest_channels is nil")
	}

	channelIds := strings.Split(channels.(string), ",")
	packet.DelMeta(wire.MetaDestServer)
	packet.DelMeta(wire.MetaDestChannels)
	payload := pkt.Marshal(packet)
	log.Debugf("Push to %v %v", channelIds, packet)

	for _, channel := range channelIds {
		messageOutFlowBytes.WithLabelValues(packet.Command).Add(float64(len(payload)))
		err := c.Srv.Push(channel, payload)
		if err != nil {
			log.Debug(err)
		}
	}
	return nil
}
