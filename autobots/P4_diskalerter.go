package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const thresholdPercentage = 85 // Set the threshold percentage here

func sendAlert(location string, current int64, min int64, threshold int64) {
	currentG := current / (1024 * 1024 * 1024)
	minG := min / (1024 * 1024 * 1024)
	thresholdG := threshold / (1024 * 1024 * 1024)

	fmt.Println("+----------------+--------------+-------------+--------------+---------------+")
	fmt.Println("| LOCATION       | CURRENT (G)  | MINIMUM (G) | THRESHOLD(G) | THRESHOLD (%) |")
	fmt.Println("+----------------+--------------+-------------+--------------+---------------+")
	fmt.Printf("| %-14s | %12d | %11d | %12d | %13d |\n", location, currentG, minG, thresholdG, thresholdPercentage)
	fmt.Println("+----------------+--------------+-------------+--------------+---------------+")
}

func parseDiskSize(sizeStr string) (int64, error) {
	sizeStr = strings.TrimSpace(sizeStr)

	// Check if the input string is just the size (e.g., "3G"), and handle it directly
	if strings.HasSuffix(sizeStr, "G") {
		size, err := strconv.ParseInt(sizeStr[:len(sizeStr)-1], 10, 64)
		if err != nil {
			return 0, err
		}
		return size * 1024 * 1024 * 1024, nil
	} else if strings.HasSuffix(sizeStr, "M") {
		size, err := strconv.ParseInt(sizeStr[:len(sizeStr)-1], 10, 64)
		if err != nil {
			return 0, err
		}
		return size * 1024 * 1024, nil
	} else {
		return 0, fmt.Errorf("unexpected output format: %s", sizeStr)
	}
}
func extractAvailableSpace(output string) (int64, error) {
	lines := strings.Split(output, "\n")
	if len(lines) < 2 {
		return 0, fmt.Errorf("unexpected output format: %s", output)
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 4 {
		return 0, fmt.Errorf("unexpected column count in df output: %s", lines[1])
	}
	return parseDiskSize(fields[3])
}

func main() {
	output, err := exec.Command("p4", "configure", "show").Output()
	if err != nil {
		fmt.Println("Error executing p4 configure show:", err)
		return
	}
	lines := strings.Split(string(output), "\n")

	var p4rootMin, p4journalMin, depotMin int64

	for _, line := range lines {
		if strings.HasPrefix(line, "filesys.P4ROOT.min=") {
			p4rootMinStr := strings.Split(line, "=")[1]
			p4rootMinStr = strings.TrimSpace(p4rootMinStr)
			p4rootMinStr = strings.Split(p4rootMinStr, " ")[0]
			p4rootMin, err = parseDiskSize(p4rootMinStr)
			if err != nil {
				fmt.Println("Error parsing p4rootMin:", err)
				return
			}
		} else if strings.HasPrefix(line, "filesys.P4JOURNAL.min=") {
			p4journalMinStr := strings.Split(line, "=")[1]
			p4journalMinStr = strings.TrimSpace(p4journalMinStr)
			p4journalMinStr = strings.Split(p4journalMinStr, " ")[0]
			p4journalMin, err = parseDiskSize(p4journalMinStr)
			if err != nil {
				fmt.Println("Error parsing p4journalMin:", err)
				return
			}
		} else if strings.HasPrefix(line, "filesys.depot.min=") {
			depotMinStr := strings.Split(line, "=")[1]
			depotMinStr = strings.TrimSpace(depotMinStr)
			depotMinStr = strings.Split(depotMinStr, " ")[0]
			depotMin, err = parseDiskSize(depotMinStr)
			if err != nil {
				fmt.Println("Error parsing depotMin:", err)
				return
			}
		}
	}

	//      fmt.Println("Parsed values:", p4rootMin, p4journalMin, depotMin) // Print parsed values

	// Get available disk spaces
	p4rootOutput, err := exec.Command("df", "-BG", "/p4/1/root").Output()
	if err != nil {
		fmt.Println("Error executing df command for P4ROOT:", err)
		return
	}
	p4rootAvailable, err := extractAvailableSpace(string(p4rootOutput))
	if err != nil {
		fmt.Println("Error fetching or parsing available space for P4ROOT:", err)
		return
	}

	p4journalOutput, err := exec.Command("df", "-BG", "/p4/1/logs/journal").Output()
	if err != nil {
		fmt.Println("Error executing df command for P4JOURNAL:", err)
		return
	}
	p4journalAvailable, err := extractAvailableSpace(string(p4journalOutput))
	if err != nil {
		fmt.Println("Error fetching or parsing available space for P4JOURNAL:", err)
		return
	}

	depotOutput, err := exec.Command("df", "-BG", "/p4/1/depots").Output()
	if err != nil {
		fmt.Println("Error executing df command for filesys.depot:", err)
		return
	}
	depotAvailable, err := extractAvailableSpace(string(depotOutput))
	if err != nil {
		fmt.Println("Error fetching or parsing available space for filesys.depot:", err)
		return
	}

	// Calculate threshold values based on configured minimums and threshold percentage
	p4rootThreshold := p4rootMin * (100 + thresholdPercentage) / 100
	p4journalThreshold := p4journalMin * (100 + thresholdPercentage) / 100
	depotThreshold := depotMin * (100 + thresholdPercentage) / 100

	if p4rootAvailable < p4rootThreshold {
		sendAlert("P4ROOT", p4rootAvailable, p4rootMin, p4rootThreshold)
	}
	if p4journalAvailable < p4journalThreshold {
		sendAlert("P4JOURNAL", p4journalAvailable, p4journalMin, p4journalThreshold)

	}
	if depotAvailable < depotThreshold {
		sendAlert("depot", depotAvailable, depotMin, depotThreshold)
	}
}
