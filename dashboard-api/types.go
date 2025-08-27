package main

import (
	"database/sql"
	"math/big"
	"time"

	"github.com/gorilla/websocket"
)

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

type WSMess struct {
	Conn   *websocket.Conn
	Struct any
}

type StatsContainer struct {
	Read      time.Time `json:"read"`
	Preread   time.Time `json:"preread"`
	PidsStats struct {
		Current int     `json:"current"`
		Limit   float64 `json:"limit"`
	} `json:"pids_stats"`
	BlkioStats struct {
		IoServiceBytesRecursive []struct {
			Major int    `json:"major"`
			Minor int    `json:"minor"`
			Op    string `json:"op"`
			Value int    `json:"value"`
		} `json:"io_service_bytes_recursive"`
		IoServicedRecursive    interface{} `json:"io_serviced_recursive"`
		IoQueueRecursive       interface{} `json:"io_queue_recursive"`
		IoServiceTimeRecursive interface{} `json:"io_service_time_recursive"`
		IoWaitTimeRecursive    interface{} `json:"io_wait_time_recursive"`
		IoMergedRecursive      interface{} `json:"io_merged_recursive"`
		IoTimeRecursive        interface{} `json:"io_time_recursive"`
		SectorsRecursive       interface{} `json:"sectors_recursive"`
	} `json:"blkio_stats"`
	NumProcs     int `json:"num_procs"`
	StorageStats struct {
	} `json:"storage_stats"`
	CPUStats struct {
		CPUUsage struct {
			TotalUsage        int64 `json:"total_usage"`
			UsageInKernelmode int64 `json:"usage_in_kernelmode"`
			UsageInUsermode   int64 `json:"usage_in_usermode"`
		} `json:"cpu_usage"`
		SystemCPUUsage int64 `json:"system_cpu_usage"`
		OnlineCpus     int   `json:"online_cpus"`
		ThrottlingData struct {
			Periods          int `json:"periods"`
			ThrottledPeriods int `json:"throttled_periods"`
			ThrottledTime    int `json:"throttled_time"`
		} `json:"throttling_data"`
	} `json:"cpu_stats"`
	PrecpuStats struct {
		CPUUsage struct {
			TotalUsage        int64 `json:"total_usage"`
			UsageInKernelmode int64 `json:"usage_in_kernelmode"`
			UsageInUsermode   int64 `json:"usage_in_usermode"`
		} `json:"cpu_usage"`
		SystemCPUUsage int64 `json:"system_cpu_usage"`
		OnlineCpus     int   `json:"online_cpus"`
		ThrottlingData struct {
			Periods          int `json:"periods"`
			ThrottledPeriods int `json:"throttled_periods"`
			ThrottledTime    int `json:"throttled_time"`
		} `json:"throttling_data"`
	} `json:"precpu_stats"`
	MemoryStats struct {
		Usage int64 `json:"usage"`
		Stats struct {
			ActiveAnon            int64 `json:"active_anon"`
			ActiveFile            int64 `json:"active_file"`
			Anon                  int64 `json:"anon"`
			AnonThp               int64 `json:"anon_thp"`
			File                  int64 `json:"file"`
			FileDirty             int64 `json:"file_dirty"`
			FileMapped            int64 `json:"file_mapped"`
			FileWriteback         int64 `json:"file_writeback"`
			InactiveAnon          int64 `json:"inactive_anon"`
			InactiveFile          int64 `json:"inactive_file"`
			KernelStack           int64 `json:"kernel_stack"`
			Pgactivate            int64 `json:"pgactivate"`
			Pgdeactivate          int64 `json:"pgdeactivate"`
			Pgfault               int64 `json:"pgfault"`
			Pglazyfree            int64 `json:"pglazyfree"`
			Pglazyfreed           int64 `json:"pglazyfreed"`
			Pgmajfault            int64 `json:"pgmajfault"`
			Pgrefill              int64 `json:"pgrefill"`
			Pgscan                int64 `json:"pgscan"`
			Pgsteal               int64 `json:"pgsteal"`
			Shmem                 int64 `json:"shmem"`
			Slab                  int64 `json:"slab"`
			SlabReclaimable       int64 `json:"slab_reclaimable"`
			SlabUnreclaimable     int64 `json:"slab_unreclaimable"`
			Sock                  int64 `json:"sock"`
			ThpCollapseAlloc      int64 `json:"thp_collapse_alloc"`
			ThpFaultAlloc         int64 `json:"thp_fault_alloc"`
			Unevictable           int64 `json:"unevictable"`
			WorkingsetActivate    int64 `json:"workingset_activate"`
			WorkingsetNodereclaim int64 `json:"workingset_nodereclaim"`
			WorkingsetRefault     int64 `json:"workingset_refault"`
			Cache                 int64 `json:"cache"`
			RSS                   int64 `json:"rss"`
			TotalInactiveFile     int64 `json:"total_inactive_file"`
		} `json:"stats"`
		Limit int64 `json:"limit"`
	} `json:"memory_stats"`
	Name     string `json:"name"`
	ID       string `json:"id"`
	Networks struct {
		Eth0 struct {
			RxBytes   int64 `json:"rx_bytes"`
			RxPackets int64 `json:"rx_packets"`
			RxErrors  int64 `json:"rx_errors"`
			RxDropped int64 `json:"rx_dropped"`
			TxBytes   int64 `json:"tx_bytes"`
			TxPackets int64 `json:"tx_packets"`
			TxErrors  int64 `json:"tx_errors"`
			TxDropped int64 `json:"tx_dropped"`
		} `json:"eth0"`
	} `json:"networks"`
}

type StatsReady struct {
	Cpu    float64  `json:"cpu"`
	Mem    string   `json:"mem"`
	MemNum *big.Int `json:"mem_num"`
	MemMax string   `json:"mem_max"`
	NetI   string   `json:"net_i"`
	NetO   string   `json:"net_o"`
	Dr     string   `json:"d_r"`
	Dw     string   `json:"d_w"`
}
