package core

import (
	"encoding/json"

	"github.com/energieip/common-components-go/pkg/dnanosense"
)

//WagoDump configuration by the switch
type WagoDump struct {
	Mac           string        `json:"mac"`
	Cluster       int           `json:"cluster,omitempty"`
	IsConfigured  bool          `json:"isConfigured,omitempty"`
	FriendlyName  string        `json:"friendlyName,omitempty"`
	IP            string        `json:"ip"`
	Nanosenses    []NanoDump    `json:"nanosenses"`
	CronJobs      []CronJobDump `json:"cronJobs"`
	DumpFrequency int           `json:"dumpFrequency,omitempty"`
	Label         string        `json:"label,omitempty"`
	Error         int           `json:"error"`
}

type CronJobDump struct {
	Group    int    `json:"group"`
	Action   string `json:"action"`
	ModbusID int    `json:"modbusID"`
	Content  int    `json:"content"`
}

type NanoDump struct {
	dnanosense.Nanosense
	ModbusIDCO2   int `json:"modbusIDco2"`
	ModbusIDCOV   int `json:"modbusIDcov"`
	ModbusIDTemp  int `json:"modbusIDtemp"`
	ModbusIDHygro int `json:"modbusIDhygro"`
}

//ToWagoDump convert map interface to wago object
func ToWagoDump(val interface{}) (*WagoDump, error) {
	var driver WagoDump
	inrec, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(inrec, &driver)
	return &driver, err
}

// ToJSON dump hvac setup struct
func (driver WagoDump) ToJSON() (string, error) {
	inrec, err := json.Marshal(driver)
	if err != nil {
		return "", err
	}
	return string(inrec), err
}
