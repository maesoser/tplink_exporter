package macdb

import (
	"bufio"
	"io"
	"os"
	"strings"
)

type DB map[string]string

func Load(filename string) (DB, DB, error) {
	var customMACs = make(map[string]string)
	var vendorMACs = make(map[string]string)
	if len(filename) == 0 {
		return customMACs, vendorMACs, nil
	}
	file, err := os.Open(filename)
	if err != nil {
		return customMACs, vendorMACs, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		line = strings.Replace(line, "\n", "", -1)
		elements := strings.Split(line, "=")
		if len(elements) == 2 {
			if key := strings.TrimSpace(elements[0]); len(key) > 0 {
				mac := strings.TrimSpace(elements[1])
				if len(key) == 8 {
					vendorMACs[key] = mac
				} else {
					customMACs[key] = mac
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return customMACs, vendorMACs, err
		}
	}
	return customMACs, vendorMACs, nil
}

func Lookup(mac string, custom, vendor DB) string {
	result := custom[mac]
	if len(result) != 0 {
		return result
	}
	return vendor[mac[:8]]
}
