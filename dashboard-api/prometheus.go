package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type MetricStore struct {
	sync.Mutex
	Data map[string]StatsContainer
}

var metrics = MetricStore{
	Data: make(map[string]StatsContainer),
}

func metricsHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	prefix := "momo_container_"

	_, _ = fmt.Fprintf(w, "# HELP %scpu_usage_seconds_total Cumulative cpu time consumed in seconds.\n", prefix)
	_, _ = fmt.Fprintf(w, "# TYPE %scpu_usage_seconds_total counter\n", prefix)

	_, _ = fmt.Fprintf(w, "# HELP %smemory_rss Size of RSS in bytes.\n", prefix)
	_, _ = fmt.Fprintf(w, "# TYPE %smemory_rss gauge\n", prefix)

	_, _ = fmt.Fprintf(w, "# HELP %smemory_cache Number of bytes of page cache memory.\n", prefix)
	_, _ = fmt.Fprintf(w, "# TYPE %smemory_cache gauge\n", prefix)

	_, _ = fmt.Fprintf(w, "# HELP %snetwork_receive_bytes_total Cumulative count of bytes received.\n", prefix)
	_, _ = fmt.Fprintf(w, "# TYPE %snetwork_receive_bytes_total counter\n", prefix)

	_, _ = fmt.Fprintf(
		w,
		"# HELP %snetwork_transmit_bytes_total Cumulative count of bytes transmitted.\n",
		prefix,
	)
	_, _ = fmt.Fprintf(w, "# TYPE %snetwork_transmit_bytes_total counter\n", prefix)

	_, _ = fmt.Fprintf(w, "# HELP %sblkio_bytes_total Cumulative count of block I/O bytes by operation.\n", prefix)
	_, _ = fmt.Fprintf(w, "# TYPE %sblkio_bytes_total counter\n", prefix)

	_, _ = fmt.Fprintf(w, "# HELP %sinfo More information about containers.\n", prefix)
	_, _ = fmt.Fprintf(w, "# TYPE %sinfo gauge\n", prefix)

	metrics.Lock()
	defer metrics.Unlock()
	for key, j := range metrics.Data {
		var statsReady StatsReady
		if sd, ok := containerStats.Load(key); ok {
			statsReady = sd.(StatsReady)
		}
		var container Container
		if cs, ok := containersDataStat.Load(key); ok {
			container = cs.(Container)
		}

		if container.Hostname == "" {
			continue
		}

		// CPU
		cpuSeconds := float64(j.CPUStats.CPUUsage.TotalUsage) / 1e9

		// Mem
		memCache := j.MemoryStats.Stats.Cache
		if memCache == 0 {
			memCache = j.MemoryStats.Stats.File
		}
		memRSS := j.MemoryStats.Stats.RSS
		if memRSS == 0 {
			memRSS = calculateRSS(
				j.MemoryStats.Stats.Anon,
				j.MemoryStats.Stats.Shmem,
				j.MemoryStats.Stats.ActiveAnon,
				j.MemoryStats.Stats.InactiveAnon,
			)
		}

		// Network
		rx := j.Networks.Eth0.RxBytes
		tx := j.Networks.Eth0.TxBytes

		name := "noname"
		if len(container.Names) > 0 {
			name = strings.TrimPrefix(container.Names[0], "/")
		}

		labelsFull := fmt.Sprintf(
			`id=%q,image=%q,name=%q,server=%q,network_type=%q,net=%q,d=%q,uptime=%q`,
			j.ID,
			container.Image,
			name,
			container.Hostname,
			container.HostConfig.NetworkMode,
			statsReady.NetO+" / "+statsReady.NetI,
			statsReady.Dw+" / "+statsReady.Dr,
			container.Status+" / "+container.State,
		)
		labels := fmt.Sprintf(`name=%q,server=%q`, name, container.Hostname)

		_, _ = fmt.Fprintf(w, "%scpu_usage_seconds_total{%s} %f\n", prefix, labels, cpuSeconds)
		_, _ = fmt.Fprintf(w, "%smemory_rss{%s} %d\n", prefix, labels, memRSS)
		_, _ = fmt.Fprintf(w, "%smemory_cache{%s} %d\n", prefix, labels, memCache)
		if container.HostConfig.NetworkMode != "host" {
			_, _ = fmt.Fprintf(w, "%snetwork_receive_bytes_total{%s} %d\n", prefix, labels, rx)
			_, _ = fmt.Fprintf(w, "%snetwork_transmit_bytes_total{%s} %d\n", prefix, labels, tx)
		}

		// Disk I/O
		totalRead := 0
		totalWrite := 0
		for _, io := range j.BlkioStats.IoServiceBytesRecursive {
			if io.Op == "read" {
				totalRead += io.Value
			} else if io.Op == "write" {
				totalWrite += io.Value
			}
		}
		_, _ = fmt.Fprintf(w, "%sblkio_bytes_total{%s,op=\"read\"} %d\n", prefix, labels, totalRead)
		_, _ = fmt.Fprintf(w, "%sblkio_bytes_total{%s,op=\"write\"} %d\n", prefix, labels, totalWrite)

		// Info
		_, _ = fmt.Fprintf(w, "%sinfo{%s} 1\n", prefix, labelsFull)
	}
}

func calculateRSS(anon, shmem, activeAnon, inactiveAnon int64) int64 {
	rss := anon - shmem

	if rss < 0 {
		rss = anon
		if rss <= 0 {
			rss = activeAnon
		}

		if rss <= 0 && (activeAnon > 0 || inactiveAnon > 0) {
			rss = activeAnon + inactiveAnon
		}
	}

	return rss
}

func updateMetric(md5Name string, j StatsContainer) {
	metrics.Lock()
	defer metrics.Unlock()
	metrics.Data[md5Name] = j
}
