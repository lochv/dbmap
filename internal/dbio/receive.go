package dbio

import (
	"bufio"
	"dbmap/internal/config"
	"os"
)

func NewRecv() chan string {
	outChan := make(chan string, 16)
	switch config.Conf.ReceiveMode {
	case "file":
		go func() {
			file, err := os.Open(config.Conf.InputFile)
			if err != nil {
				panic(err.Error())
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				outChan <- scanner.Text()
			}
		}()
	case "remote":
		//todo impl
	}
	return outChan
}
