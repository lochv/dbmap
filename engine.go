package dbmap

import (
	"dbmap/internal/config"
	"fmt"
	"github.com/coocood/freecache"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"net"
	"strconv"
	"time"
)

type pkInfo struct {
	ip   net.IP
	port int
	seq  uint32
}

type engine struct {
	ipTestOutBound string
	ring           *pcap.Handle
	ringSend       *pcap.Handle
	ipSrc          net.IP
	portSrc        layers.TCPPort
	macSrc         net.HardwareAddr
	macGw          net.HardwareAddr
	in             chan string
	out            chan string
	domainCache    *freecache.Cache
	seq            uint32
	iName          string
	options        gopacket.SerializeOptions
	ethernetLayer  *layers.Ethernet
	sendChan       chan pkInfo
}

func newEngine(ip string, port int, iName string, in chan string, out chan string) *engine {

	p := &engine{
		ipTestOutBound: ip,
		portSrc:        layers.TCPPort(port),
		in:             in,
		out:            out,
		domainCache:    freecache.NewCache(1024 * 1024),
		seq:            1090024978,
		iName:          iName,
		sendChan:       make(chan pkInfo, 10),
	}

	fmt.Println("[+] running on ", iName)
	ring, err := pcap.OpenLive(iName, 65536, false, 10)
	ringSend, err := pcap.OpenLive(iName, 65536, false, 0)
	if err != nil {
		panic(err.Error())
	}
	p.ring = ring
	p.ringSend = ringSend
	p.updateInfo()
	go p.listen()
	go p.run()
	time.Sleep(time.Second * 1)
	return p
}

func (this *engine) updateInfo() {
	var (
		ethLayer layers.Ethernet
		ipLayer  layers.IPv4
		tcpLayer layers.TCP
	)

	ring, err := pcap.OpenLive(this.iName, 65536, false, 3)

	if err != nil {
		panic(err.Error())
	}
	err = ring.SetBPFFilter("tcp and dst port 1337")
	if err != nil {
		panic(err.Error())
	}

	packetSource := gopacket.NewPacketSource(ring, layers.LinkTypeIPv4)
	go func() {
		for packet := range packetSource.Packets() {
			parser := gopacket.NewDecodingLayerParser(
				layers.LayerTypeEthernet,
				&ethLayer,
				&ipLayer,
				&tcpLayer,
			)
			foundLayerTypes := []gopacket.LayerType{}
			parser.DecodeLayers(packet.Data(), &foundLayerTypes)
			for _, layerType := range foundLayerTypes {
				if layerType == layers.LayerTypeEthernet {
					this.macSrc = ethLayer.SrcMAC
					this.macGw = ethLayer.DstMAC
				}
				if layerType == layers.LayerTypeIPv4 {
					this.ipSrc = ipLayer.SrcIP
				}
			}
			return
		}
	}()
	net.DialTimeout("tcp", this.ipTestOutBound+":1337", 2*time.Second)
	ring.Close()
	this.options = gopacket.SerializeOptions{
		FixLengths:       true,
		ComputeChecksums: true,
	}

	this.ethernetLayer = &layers.Ethernet{
		SrcMAC:       this.macSrc,
		DstMAC:       this.macGw,
		EthernetType: layers.EthernetTypeIPv4,
	}
}

func (this *engine) listen() {

	err := this.ring.SetBPFFilter("tcp[tcpflags] & tcp-syn == tcp-syn and tcp[tcpflags] & tcp-ack == tcp-ack and dst port " + strconv.Itoa(int(this.portSrc)))
	if err != nil {
		panic(err.Error())
	}

	var (
		ethLayer layers.Ethernet
		ipLayer  layers.IPv4
		tcpLayer layers.TCP
	)

	packetSource := gopacket.NewPacketSource(this.ring, layers.LinkTypeIPv4)
	for packet := range packetSource.Packets() {
		parser := gopacket.NewDecodingLayerParser(
			layers.LayerTypeEthernet,
			&ethLayer,
			&ipLayer,
			&tcpLayer,
		)
		foundLayerTypes := []gopacket.LayerType{}
		err := parser.DecodeLayers(packet.Data(), &foundLayerTypes)
		if err != nil {
			fmt.Println("Trouble decoding layers: ", err)
		}

		for _, layerType := range foundLayerTypes {
			if layerType == layers.LayerTypeTCP {
				host, _ := this.domainCache.Get(i32tob(tcpLayer.Ack - 1))
				if len(host) != 0 {
					res :=  fmt.Sprintf("Host %s open %d", string(host), int(tcpLayer.SrcPort))
					this.out <- res
				}
			}
		}
	}
}

func (this *engine) sendSyn(ip net.IP, port int, seq uint32) {

	ipLayer := &layers.IPv4{
		SrcIP:    this.ipSrc,
		DstIP:    ip,
		Version:  4,
		IHL:      5,
		Flags:    layers.IPv4DontFragment,
		TTL:      64,
		Protocol: layers.IPProtocolTCP,
	}

	tcpLayer := &layers.TCP{
		SrcPort: this.portSrc,
		DstPort: layers.TCPPort(port),
		Seq:     seq,
		SYN:     true,
		Window:  1024,
	}

	err := tcpLayer.SetNetworkLayerForChecksum(ipLayer)
	if err != nil {
		fmt.Println(err.Error())
	}

	buffer := gopacket.NewSerializeBuffer()
	err = gopacket.SerializeLayers(
		buffer,
		this.options,
		this.ethernetLayer,
		ipLayer,
		tcpLayer,
	)

	outgoingPacket := buffer.Bytes()

	err = this.ringSend.WritePacketData(outgoingPacket)
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (this *engine) worker(seq uint32) {
	go func() {
		startSeq := seq
		for {
			//reset seq number
			if seq-startSeq > 200000 {
				seq = startSeq
			}
			host := <-this.in
			dstaddrs, err := net.LookupIP(host)
			if err != nil {
				//fmt.Println(err.Error())
				continue
			}
			if len(dstaddrs) == 0 {
				continue
			}
			ip := dstaddrs[0].To4()
			seq += 1
			this.domainCache.Set(i32tob(seq), append([]byte(host)), 90)
			for _, port := range config.Conf.Ports {
				this.sendChan <- pkInfo{
					ip:   ip,
					port: port,
					seq:  seq,
				}
			}
		}
	}()
}

func (this *engine) run() {
	for {
		pk := <-this.sendChan
		this.sendSyn(pk.ip, pk.port, pk.seq)
	}
}

func i32tob(val uint32) []byte {
	r := make([]byte, 4)
	for i := uint32(0); i < 4; i++ {
		r[i] = byte((val >> (8 * i)) & 0xff)
	}
	return r
}
