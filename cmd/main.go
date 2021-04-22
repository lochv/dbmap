package main

import (
	"dbmap"
	"dbmap/internal/config"
	"dbmap/internal/dbio"
)

func main() {
	recvChan := dbio.NewRecv()
	rpChan := dbio.NewRp()
	db := dbmap.New(config.Conf.TestIp, config.Conf.SourcePort, config.Conf.Iface, recvChan, rpChan)
	db.Run(uint32(config.Conf.Workers))
	db.Wait()
}
