package main

import "database/sql"

type Line struct {
	ContainerID string `json:"container_id"`
	Name        string `json:"name"`
	Hostname    string `json:"hostname"`
	Md5Name     string `json:"md5_name"`
	Type        string `json:"type"`
	Body        string `json:"body"`
}

type Container struct {
	Hostname   string `json:"Hostname"`
	Md5Name    string `json:"Md5Name"`
	Command    string `json:"Command"`
	Created    int    `json:"Created"`
	HostConfig struct {
		NetworkMode string `json:"NetworkMode"`
	} `json:"HostConfig"`
	ID      string `json:"Id"`
	Image   string `json:"Image"`
	ImageID string `json:"ImageID"`
	Labels  struct {
	} `json:"Labels"`
	Mounts []struct {
		Destination string `json:"Destination"`
		Mode        string `json:"Mode"`
		Propagation string `json:"Propagation"`
		Rw          bool   `json:"RW"`
		Source      string `json:"Source"`
		Type        string `json:"Type"`
	} `json:"Mounts"`
	Names           []string `json:"Names"`
	NetworkSettings struct {
		Networks struct {
			Bridge struct {
				Aliases             interface{} `json:"Aliases"`
				DriverOpts          interface{} `json:"DriverOpts"`
				EndpointID          string      `json:"EndpointID"`
				Gateway             string      `json:"Gateway"`
				GlobalIPv6Address   string      `json:"GlobalIPv6Address"`
				GlobalIPv6PrefixLen int         `json:"GlobalIPv6PrefixLen"`
				IPAMConfig          interface{} `json:"IPAMConfig"`
				IPAddress           string      `json:"IPAddress"`
				IPPrefixLen         int         `json:"IPPrefixLen"`
				IPv6Gateway         string      `json:"IPv6Gateway"`
				Links               interface{} `json:"Links"`
				MacAddress          string      `json:"MacAddress"`
				NetworkID           string      `json:"NetworkID"`
			} `json:"bridge"`
		} `json:"Networks"`
	} `json:"NetworkSettings"`
	Ports []struct {
		IP          string `json:"IP,omitempty"`
		PrivatePort int    `json:"PrivatePort"`
		PublicPort  int    `json:"PublicPort,omitempty"`
		Type        string `json:"Type"`
	} `json:"Ports"`
	State  string `json:"State"`
	Status string `json:"Status"`
}

type Statistic struct {
	MessageStats struct {
		DeliverGet        int `json:"deliver_get"`
		DeliverGetDetails struct {
			Rate float64 `json:"rate"`
		} `json:"deliver_get_details"`
	} `json:"message_stats"`
}

type Alert struct {
	ID           int            `db:"id" json:"id"`
	ContainerMd5 string         `db:"container_md5" json:"container_md5"`
	TelegramID   string         `db:"telegram_id" json:"telegram_id"`
	TelegramName sql.NullString `db:"telegram_name" json:"telegram_name,omitempty"`
	KeyAlert     string         `db:"key_alert" json:"key_alert"`
	DateCreate   string         `db:"date_create" json:"date_create,omitempty"`
}

type PreparedAlert map[string][]struct {
	Alert Alert
	Data  Line
}

type TelegramChat struct {
	TelegramName string `db:"telegram_name" json:"telegram_name"`
	TelegramID   string `db:"telegram_id" json:"telegram_id"`
}
