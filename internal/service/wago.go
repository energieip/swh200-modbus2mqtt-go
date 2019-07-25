package service

import (
	"github.com/energieip/common-components-go/pkg/dnanosense"
	"github.com/energieip/common-components-go/pkg/dwago"
	"github.com/energieip/common-components-go/pkg/pconst"
	"github.com/energieip/common-components-go/pkg/tools"
	"github.com/energieip/swh200-modbus2mqtt-go/internal/core"
	"github.com/goburrow/modbus"
	"github.com/romana/rlog"
)

func (s *Service) onWagoSetup(conf dwago.WagoDef) {
	var wago core.WagoDump
	d, _ := s.wagos.Get(conf.Mac)
	if d != nil {
		status, _ := core.ToWagoDump(d)
		if status == nil || status.IsConfigured {
			return
		}
		wago = *status
	}
	if conf.Cluster != nil {
		wago.Cluster = *conf.Cluster
	}
	if conf.FriendlyName != nil {
		wago.FriendlyName = *conf.FriendlyName
	} else {
		wago.FriendlyName = conf.Mac
	}
	if conf.IP != nil {
		wago.IP = *conf.IP
	}
	if conf.Label != nil {
		wago.Label = *conf.Label
	}
	freq := 1000
	if conf.DumpFrequency != nil {
		freq = *conf.DumpFrequency
	}
	wago.DumpFrequency = freq
	var nanos []core.NanoDump
	for _, v := range conf.Nanosenses {
		nano := core.NanoDump{}
		nano.Label = v.Label
		nano.Mac = v.Mac
		nano.ModbusIDCO2 = v.CO2
		nano.ModbusIDCOV = v.COV
		nano.ModbusIDHygro = v.Hygrometry
		nano.ModbusIDTemp = v.Temperature
		nano.FriendlyName = v.FriendlyName
		nano.Group = v.Group
		nano.DumpFrequency = freq
		nano.Cluster = v.Cluster
		nano.Error = 1 //info not yet available
		nanos = append(nanos, nano)
	}
	wago.Nanosenses = nanos
	wago.Mac = conf.Mac
	var progs []core.CronJobDump
	for _, v := range conf.CronJobs {
		cron := core.CronJobDump{}
		cron.ModbusID = v.ModbusID
		cron.Group = v.Group
		cron.Action = v.Action
		progs = append(progs, cron)
	}
	wago.CronJobs = progs
	if wago.Mac != "" {
		wago.IsConfigured = true
	}
	s.wagos.Set(conf.Mac, wago)
}

func (s *Service) onWagoUpdate(conf dwago.WagoDef) {
	d, _ := s.wagos.Get(conf.Mac)
	if d == nil {
		return
	}

	status, _ := core.ToWagoDump(d)
	if status == nil || !status.IsConfigured {
		return
	}
	wago := *status

	if conf.Cluster != nil {
		wago.Cluster = *conf.Cluster
	}
	if conf.FriendlyName != nil {
		wago.FriendlyName = *conf.FriendlyName
	} else {
		wago.FriendlyName = conf.Mac
	}
	if conf.IP != nil {
		wago.IP = *conf.IP
	}
	if conf.Label != nil {
		wago.Label = *conf.Label
	}
	if conf.IsConfigured != nil {
		wago.IsConfigured = *conf.IsConfigured
	}
	freq := 1000
	if conf.DumpFrequency != nil {
		freq = *conf.DumpFrequency
	}
	wago.DumpFrequency = freq
	var nanos []core.NanoDump
	for _, v := range conf.Nanosenses {
		nano := core.NanoDump{}
		nano.Label = v.Label
		nano.Mac = v.Mac
		nano.FriendlyName = v.FriendlyName
		nano.ModbusIDCO2 = v.CO2
		nano.ModbusIDCOV = v.COV
		nano.ModbusIDHygro = v.Hygrometry
		nano.ModbusIDTemp = v.Temperature
		nano.Cluster = v.Cluster
		nano.Group = v.Group
		nano.DumpFrequency = freq

		nano.Error = 1 //info not yet available
		nanos = append(nanos, nano)
	}
	wago.Nanosenses = nanos
	wago.Mac = conf.Mac
	var progs []core.CronJobDump
	for _, v := range conf.CronJobs {
		cron := core.CronJobDump{}
		cron.ModbusID = v.ModbusID
		cron.Group = v.Group
		cron.Action = v.Action
		progs = append(progs, cron)
	}
	wago.CronJobs = progs
	wago.IsConfigured = true
	wago.Error = 1
	s.wagos.Set(conf.Mac, wago)
}

func (s *Service) sendHello(driver core.WagoDump) {
	driverHello := dwago.Wago{
		Mac:          driver.Mac,
		IP:           driver.IP,
		Cluster:      driver.Cluster,
		IsConfigured: false,
		Protocol:     "modbus",
		FriendlyName: driver.FriendlyName,
	}
	dump, _ := tools.ToJSON(driverHello)

	err := s.local.SendCommand("/read/wago/"+driver.Mac+"/"+pconst.UrlHello, dump)
	if err == nil {
		rlog.Info("Hello " + driver.Mac + " sent to the server")
	}
	for _, v := range driver.Nanosenses {
		nano := dnanosense.Nanosense{}
		nano.Label = v.Label
		nano.Mac = v.Mac
		nano.Group = v.Group
		nano.Cluster = v.Cluster
		nano.DumpFrequency = v.DumpFrequency
		nano.Error = v.Error
		dump, _ = tools.ToJSON(nano)
		s.local.SendCommand("/read/nano/"+driver.Mac+"/"+pconst.UrlHello, dump)
	}
}

func (s *Service) sendDump(driver core.WagoDump) {
	driverHello := dwago.Wago{
		Mac:          driver.Mac,
		IP:           driver.IP,
		Cluster:      driver.Cluster,
		IsConfigured: driver.IsConfigured,
		Protocol:     "modbus",
		FriendlyName: driver.FriendlyName,
	}
	dump, err := tools.ToJSON(driverHello)
	if err != nil {
		rlog.Errorf("Could not dump Wago %v status %v", driver.Mac, err.Error())
		return
	}

	s.local.SendCommand("/read/wago/"+driver.Mac+"/"+pconst.UrlStatus, dump)
}

func bytes2int(val []byte) int {
	result := 0
	for _, b := range val {
		result = result*256 + int(b)
	}
	return result
}

func (s *Service) updateWagoStatus(driver core.WagoDump) {
	if driver.IP == "" {
		return
	}
	handler := modbus.NewTCPClientHandler(driver.IP + ":502")
	err := handler.Connect()
	defer handler.Close()
	if err != nil {
		rlog.Errorf("Cannot connect on %v : %v", driver.IP+":502", err.Error())
		driver.Error = 1
		for _, v := range driver.Nanosenses {
			nano := dnanosense.Nanosense{}
			nano.Group = v.Group
			nano.Cluster = v.Cluster
			nano.DumpFrequency = v.DumpFrequency
			nano.Label = v.Label
			nano.Mac = v.Mac
			nano.Error = v.Error
			dump, _ := tools.ToJSON(nano)
			s.local.SendCommand("/read/nano/"+driver.Mac+"/"+pconst.UrlStatus, dump)
		}
		s.wagos.Set(driver.Mac, driver)
		return
	}
	client := modbus.NewClient(handler)

	var nanos []core.NanoDump
	for _, v := range driver.Nanosenses {
		errorRead := false
		nano := dnanosense.Nanosense{}
		nano.Label = v.Label
		nano.Mac = v.Mac
		nano.Cluster = v.Cluster
		nano.Group = v.Group
		nano.FriendlyName = v.FriendlyName
		nano.DumpFrequency = v.DumpFrequency
		quantity := uint16(1)

		results, err := client.ReadHoldingRegisters(uint16(v.ModbusIDCO2), quantity)
		if err != nil {
			rlog.Errorf("Cannot read on (%v) %v : %v", v.Label, v.ModbusIDCO2, err.Error())
			errorRead = true
		} else {
			nano.CO2 = bytes2int(results)
		}

		results, err = client.ReadHoldingRegisters(uint16(v.ModbusIDCOV), quantity)
		if err != nil {
			rlog.Errorf("Cannot read on (%v) %v : %v", v.Label, v.ModbusIDCOV, err.Error())
			errorRead = true
		} else {
			nano.COV = bytes2int(results)
		}

		results, err = client.ReadHoldingRegisters(uint16(v.ModbusIDTemp), quantity)
		if err != nil {
			rlog.Errorf("Cannot read on (%v) %v : %v", v.Label, v.ModbusIDTemp, err.Error())
			errorRead = true
		} else {
			nano.Temperature = bytes2int(results)
		}

		results, err = client.ReadHoldingRegisters(uint16(v.ModbusIDHygro), quantity)
		if err != nil {
			rlog.Errorf("Cannot read on (%v) %v : %v", v.Label, v.ModbusIDHygro, err.Error())
			errorRead = true
		} else {
			nano.Hygrometry = bytes2int(results)
		}

		if errorRead {
			nano.Error = 1
		} else {
			nano.Error = 0
		}
		dump, _ := tools.ToJSON(nano)
		s.local.SendCommand("/read/nano/"+nano.Mac+"/"+pconst.UrlStatus, dump)
		nanos = append(nanos, v)
	}
	driver.Nanosenses = nanos
	s.wagos.Set(driver.Mac, driver)
}
