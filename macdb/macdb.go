package macdb

import (
	"bufio"
	"io"
	"os"
	"strings"
)

type DB map[string]string

type MACDB struct {
	custom DB
	vendor DB
}

/*
Load loads a text file containing both Vendor MAC and custom MACs

*/
func (db *MACDB) Load(filename string) error {
	db.custom = make(map[string]string)
	db.vendor = make(map[string]string)
	file, err := os.Open(filename)
	if err != nil {
		return err
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
					db.vendor[key] = mac
				} else {
					db.custom[key] = mac
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *MACDB) Lookup(mac string) string {
	result := db.custom[mac]
	if len(result) != 0 {
		return result
	}
	return db.vendor[mac[:8]]
}

func (db *MACDB) Size() int {
	return len(db.custom) + len(db.vendor)
}
