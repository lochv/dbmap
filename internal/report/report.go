package report

import (
	"dbmap/internal/config"
	"fmt"
)

func New() chan string {
	in := make(chan string, 16)
	switch config.Conf.ReportMode {
	case "console":
		go func() {
			for {
				mess := <-in
				fmt.Println(mess)
			}
		}()
	case "remote":
		//todo impl
	}
	return in
}
