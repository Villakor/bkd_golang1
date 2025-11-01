package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	url         = "http://srv.msk01.gigacorp.local/_stats"
	maxLoad     = 30
	memLimit    = 0.8
	diskLimit   = 0.9
	netLimit    = 0.9
	maxErrors   = 3
	pollingTime = 10 * time.Second
)

func main() {
	errorCount := 0

	for {
		resp, err := http.Get(url)
		if err != nil || resp.StatusCode != http.StatusOK {
			errorCount++
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic.")
				return
			}
			time.Sleep(pollingTime)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errorCount++
			time.Sleep(pollingTime)
			continue
		}

		data := strings.Split(strings.TrimSpace(string(body)), ",")
		if len(data) != 7 {
			errorCount++
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic.")
				return
			}
			time.Sleep(pollingTime)
			continue
		}

		errorCount = 0
		values := make([]float64, 7)
		for i, v := range data {
			values[i], _ = strconv.ParseFloat(v, 64)
		}

		load := values[0]
		memTotal, memUsed := values[1], values[2]
		diskTotal, diskUsed := values[3], values[4]
		netTotal, netUsed := values[5], values[6]

		if load > maxLoad {
			fmt.Printf("Load Average is too high: %.0f\n", load)
		}

		memUsage := memUsed / memTotal
		if memUsage > memLimit {
			fmt.Printf("Memory usage too high: %.0f%%\n", memUsage*100)
		}

		diskUsage := diskUsed / diskTotal
		if diskUsage > diskLimit {
			freeMb := (diskTotal - diskUsed) / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %.0f Mb left\n", freeMb)
		}

		netUsage := netUsed / netTotal
		if netUsage > netLimit {
			freeMbit := (netTotal - netUsed) * 8 / (1024 * 1024)
			fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", freeMbit)
		}

		time.Sleep(pollingTime)
	}
}
