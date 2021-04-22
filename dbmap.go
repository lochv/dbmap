package dbmap

import (
	"dbmap/internal/dbio"
)

type dbmap struct {
	kill chan int
	*engine
}

func New(ip string, port int, iName string, in chan string, out chan dbio.Report) *dbmap {
	return &dbmap{
		engine: newEngine(ip, port, iName, in, out),
		kill:   make(chan int),
	}
}

func (this *dbmap) Run(worker uint32) {
	go func() {
		for i := uint32(0); i < worker; i++ {
			this.worker(this.seq + i*3000000)
		}
	}()
}

func (this *dbmap) Wait() {
	<-this.kill
}
