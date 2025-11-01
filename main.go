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
	memLimit    = 80 // проценты
	diskLimit   = 90 // проценты
	netLimit    = 90 // проценты
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

		load := int(values[0])
		memTotal, memUsed := int(values[1]), int(values[2])
		diskTotal, diskUsed := int(values[3]), int(values[4])
		netTotal, netUsed := int(values[5]), int(values[6])

		if load > maxLoad {
			fmt.Printf("Load Average is too high: %d\n", load)
		}

		memUsage := memUsed * 100 / memTotal
		if memUsage > memLimit {
			fmt.Printf("Memory usage too high: %d%%\n", memUsage)
		}

		diskUsage := diskUsed * 100 / diskTotal
		if diskUsage > diskLimit {
			freeMb := (diskTotal - diskUsed) / (1024 * 1024)
			fmt.Printf("Free disk space is too low: %d Mb left\n", freeMb)
		}

		netUsage := netUsed * 100 / netTotal
		if netUsage > netLimit {
			// Переводим свободную полосу сети в мегабиты через 1024*1024
			freeMbit := int((netTotal - netUsed) * 8 / (1024 * 1024))
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", freeMbit)
		}

		time.Sleep(pollingTime)
	}
}
