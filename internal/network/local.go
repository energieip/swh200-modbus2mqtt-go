package network

import (
	"encoding/json"
	"time"

	"github.com/energieip/common-components-go/pkg/pconst"

	"github.com/energieip/common-components-go/pkg/dwago"

	genericNetwork "github.com/energieip/common-components-go/pkg/network"
	pkg "github.com/energieip/common-components-go/pkg/service"
	"github.com/romana/rlog"
)

var (
	EventSetup  = "setup"
	EventUpdate = "update"
)

//ServerNetwork network object
type ServerNetwork struct {
	Iface  genericNetwork.NetworkInterface
	Events chan map[string]dwago.WagoDef
}

//CreateServerNetwork create network server object
func CreateServerNetwork() (*ServerNetwork, error) {
	serverBroker, err := genericNetwork.NewNetwork(genericNetwork.MQTT)
	if err != nil {
		return nil, err
	}
	serverNet := ServerNetwork{
		Iface:  serverBroker,
		Events: make(chan map[string]dwago.WagoDef),
	}
	return &serverNet, nil

}

//Connect service to server broker
func (net ServerNetwork) Connect(conf pkg.ServiceConfig) error {
	cbkServer := make(map[string]func(genericNetwork.Client, genericNetwork.Message))
	cbkServer["/write/wago/+/"+pconst.UrlSetting] = net.onUpdateConf
	cbkServer["/write/wago/+/"+pconst.UrlSetup] = net.onSetup

	confServer := genericNetwork.NetworkConfig{
		IP:        conf.LocalBroker.IP,
		Port:      conf.LocalBroker.Port,
		Callbacks: cbkServer,
		LogLevel:  conf.LogLevel,
		User:      conf.LocalBroker.Login,
		Password:  conf.LocalBroker.Password,
		CaPath:    conf.LocalBroker.CaPath,
		Secure:    conf.LocalBroker.Secure,
	}

	for {
		rlog.Info("Try to connect to " + conf.LocalBroker.IP)
		err := net.Iface.Initialize(confServer)
		if err == nil {
			rlog.Info("Connected to server broker " + conf.LocalBroker.IP)
			return err
		}
		timer := time.NewTicker(time.Second)
		rlog.Error("Cannot connect to broker " + conf.LocalBroker.IP + " error: " + err.Error())
		rlog.Error("Try to reconnect " + conf.LocalBroker.IP + " in 1s")

		select {
		case <-timer.C:
			continue
		}
	}
}

func (net ServerNetwork) onUpdateConf(client genericNetwork.Client, msg genericNetwork.Message) {
	payload := msg.Payload()
	rlog.Info(msg.Topic() + " : " + string(payload))
	var conf dwago.WagoDef
	err := json.Unmarshal(payload, &conf)
	if err != nil {
		rlog.Error("Cannot parse config ", err.Error())
		return
	}
	event := make(map[string]dwago.WagoDef)
	event[EventUpdate] = conf
	net.Events <- event
}

func (net ServerNetwork) onSetup(client genericNetwork.Client, msg genericNetwork.Message) {
	payload := msg.Payload()
	rlog.Info(msg.Topic() + " : " + string(payload))
	var setup dwago.WagoDef
	err := json.Unmarshal(payload, &setup)
	if err != nil {
		rlog.Error("Cannot parse config ", err.Error())
		return
	}
	event := make(map[string]dwago.WagoDef)
	event[EventSetup] = setup
	net.Events <- event
}

//Disconnect from server
func (net ServerNetwork) Disconnect() {
	net.Iface.Disconnect()
}

//SendCommand to server
func (net ServerNetwork) SendCommand(topic, content string) error {
	err := net.Iface.SendCommand(topic, content)
	if err != nil {
		rlog.Error("Cannot send : " + content + " on: " + topic + " Error: " + err.Error())
	} else {
		rlog.Info("Sent : " + content + " on: " + topic)
	}
	return err
}
