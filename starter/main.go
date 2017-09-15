package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	ev := os.Getenv("LOGSTASH_HOSTS")
	if err := udateConffile(ev); err != nil {
		fmt.Printf("Error updating configuration file: %v\n", err)
		os.Exit(1)
	}
}

// update conffile to add logstash setting (no need for elasticsearch setting)
func udateConffile(logstashHosts string) error {
	fileName := "/etc/beatconf/dbeat.yml"
	file, err := os.Create(fileName + ".new")
	if err != nil {
		fmt.Printf("Error creating new conffile for creation: %v\n", err)
		return err
	}
	filetpt, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Error opening conffile: %s : %v\n", fileName, err)
		return err
	}
	scanner := bufio.NewScanner(filetpt)
	elasticsearch := false
	logstash := false
	//nbLine := 0
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "output.elasticsearch:") {
			elasticsearch = true
			logstash = false
		} else if strings.Contains(line, "output.logstash:") {
			logstash = true
			elasticsearch = false
			if logstashHosts != "" {
				line = "output.logstash:"
			}
		} else if strings.Contains(line, "logging.level:") {
			logstash = false
			elasticsearch = false
		}
		if logstashHosts != "" {
			if elasticsearch {
				line = "#" + line
			}
			if logstash {
				if strings.Contains(line, "hosts:") {
					list := strings.Split(logstashHosts, ",")
					line = "  hosts: ['" + strings.TrimSpace(list[0]) + "'"
					for _, host := range list[1:] {
						line += "," + strings.TrimSpace(host)
					}
					line += "]"
				}
				if strings.Contains(line, "ssl.certificate_authorities:") {
					if lca := os.Getenv("LOGSTASH_CERT_AUTHS"); lca != "" {
						list := strings.Split(lca, ",")
						line = "  ssl.certificate_authorities: ['" + strings.TrimSpace(list[0]) + "'"
						for _, cert := range list[1:] {
							line += "," + strings.TrimSpace(cert)
						}
						line += "]"
					}
				}
				if strings.Contains(line, "ssl.certificate:") {
					if lc := os.Getenv("LOGSTASH_CERT"); lc != "" {
						line = "  ssl.certificate: " + lc
					}
				}
				if strings.Contains(line, "ssl.key:") {
					if lk := os.Getenv("LOGSTASH_KEY"); lk != "" {
						line = "  ssl.key: " + lk
					}
				}
			}
		}
		//nbLine++
		//fmt.Printf("%d:%s\n", nbLine, line)
		file.WriteString(line + "\n")
	}
	if err = scanner.Err(); err != nil {
		fmt.Printf("Error reading conffile: %s %v\n", fileName, err)
		file.Close()
		return err
	}
	file.Close()
	os.Remove(fileName)
	err2 := os.Rename(fileName+".new", fileName)
	if err2 != nil {
		fmt.Printf("Error renaming conffile .new: %v\n", err)
		return err
	}
	return nil
}
