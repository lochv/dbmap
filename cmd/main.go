package main

import (
	"dbmap"
	"dbmap/internal/config"
	"dbmap/internal/receive"
	"dbmap/internal/report"
)

func main() {
	revChan := receive.New()
	rpChan := report.New()
	db := dbmap.New(config.Conf.TestIp, config.Conf.SourcePort, config.Conf.Iface, revChan, rpChan)
	db.Run(uint32(config.Conf.Workers))
	db.Wait()
}
