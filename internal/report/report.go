package report

import (
	"dbmap/internal/config"
	"encoding/json"
	"fmt"
	"os"
)

type Report struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

func New() chan Report {
	in := make(chan Report, 32)
	switch config.Conf.ReportMode {
	case "console":
		go func() {
			for {
				mess := <-in
				fmt.Printf("\nHost %s open %d", mess.Host, mess.Port)
			}
		}()
	case "remote":
	//todo impl
	case "file":
		go func() {
			for {
				mess := <-in
				f, err := os.OpenFile(config.Conf.OutputFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
				if err != nil {
					continue
				}
				bytes, _ := json.Marshal(mess)
				f.WriteString("\n")
				f.Write(bytes)
				f.Close()
			}
		}()
	}
	return in
}
