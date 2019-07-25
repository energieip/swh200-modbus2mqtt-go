package service

import (
	"os"
	"time"

	"github.com/energieip/swh200-modbus2mqtt-go/internal/core"
	net "github.com/energieip/swh200-modbus2mqtt-go/internal/network"

	pkg "github.com/energieip/common-components-go/pkg/service"
	"github.com/energieip/common-components-go/pkg/tools"
	cmap "github.com/orcaman/concurrent-map"
	"github.com/romana/rlog"
)

const (
	DefaultTimerDump = 1000
)

//Service content
type Service struct {
	local     net.ServerNetwork //local broker for drivers
	Mac       string            //Switch mac address
	timerDump time.Duration     //in seconds
	wagos     cmap.ConcurrentMap
	conf      pkg.ServiceConfig
}

//Initialize service
func (s *Service) Initialize(confFile string) error {
	s.wagos = cmap.New()

	conf, err := pkg.ReadServiceConfig(confFile)
	if err != nil {
		rlog.Error("Cannot parse configuration file " + err.Error())
		return err
	}
	s.conf = *conf

	mac, _ := tools.GetNetworkInfo()
	s.Mac = mac

	os.Setenv("RLOG_LOG_LEVEL", conf.LogLevel)
	os.Setenv("RLOG_LOG_NOTIME", "yes")
	os.Setenv("RLOG_TIME_FORMAT", "2006/01/06 15:04:05.000")
	rlog.UpdateEnv()
	rlog.Info("Starting modbus2mqtt service")

	s.timerDump = DefaultTimerDump

	broker, err := net.CreateServerNetwork()
	if err != nil {
		rlog.Error("Cannot connect to broker " + conf.LocalBroker.IP + " error: " + err.Error())
		return err
	}
	s.local = *broker

	go s.local.Connect(*conf)
	rlog.Info("modbus2mqtt service started")
	return nil
}

//Stop service
func (s *Service) Stop() {
	rlog.Info("Stopping modbus2mqtt service")
	s.local.Disconnect()
	rlog.Info("modbus2mqtt service stopped")
}

func (s *Service) cronDump() {
	timerDump := time.NewTicker(s.timerDump * time.Millisecond)
	for {
		select {
		case <-timerDump.C:
			for _, v := range s.wagos.Items() {
				driver, _ := core.ToWagoDump(v)
				if driver.IsConfigured {
					s.updateWagoStatus(*driver)
					s.sendDump(*driver)
				} else {
					s.sendHello(*driver)
				}
			}
		}
	}
}

//Run service mainloop
func (s *Service) Run() error {
	//TODO manage case restart with already connected device???
	go s.cronDump()
	for {
		select {
		case evtUpdate := <-s.local.Events:
			for evtType, event := range evtUpdate {
				switch evtType {
				case net.EventSetup:
					s.onWagoSetup(event)
				case net.EventUpdate:
					s.onWagoUpdate(event)
				}
			}
		}
	}
	return nil
}
