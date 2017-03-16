package main

import (
	//"log"
	//"net/http"
	"fmt"
	"os"
	"github.com/condor/gtpv1u"
	"github.com/docopt/docopt-go"
	//"strings"
	//"time"
	"net"
	"strconv"
	"time"
	"sync"
)
const (

	DEFAULT_PORT = 8080

	USAGE = `S1-U Stats

Usage:
   s1u-stats [options]
   s1u-stats --version

Options:
   -h --help                   show this help message and exit
   --version                   show version and exit
   --eNodeBPort PORT           UDP port that will receive GTPv1-U Packets from EPC [default: 2152]
   --eNodeBAddress IP          eNodeBAddress that all packets will be forwarded to [default: 10.1.20.2]
   --sGWPort PORT              UDP port that will receive GTPv1-U Packets from eNodeB [default: 2153]
   --sGWAddress IP             ePC address that all packets will be forwarded to [default: 10.2.20.2]
   --reportPeriod PERIOD       Number of seconds that all stats will be reseted [default: 60]`
)

var (
	version = "0.0.0"
)

type GTPv1uStatsService struct
{
	ForwardAddr   *net.UDPAddr
	Connection    *net.UDPConn
	ReportPeriod	uint16
	lock           sync.Mutex
	stats          map[string] uint32
}

func NewGTPv1uStatsService(reportPeriod uint16) (ser *GTPv1uStatsService) {
	ser = &GTPv1uStatsService{ForwardAddr:nil, Connection: nil, ReportPeriod: reportPeriod, lock: sync.Mutex{}, stats: map[string]uint32{}}
	return ser
}

func (s *GTPv1uStatsService) Dial(addr string, port int16) (err error) {
	address, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return err
	} else {
		conn, err := net.DialUDP("udp", nil, address)
		if err != nil {
			return err
		}
		s.ForwardAddr = address
		s.Connection = conn
	}
	return
}

func (s *GTPv1uStatsService) AddStats(key string, value uint32) {
	s.lock.Lock()
	s.stats[key] += value
	s.lock.Unlock()
}

func (s *GTPv1uStatsService) GTPPacketHandle(req * gtpv1u.Packet, b * []byte, conn * net.UDPConn, addr * net.Addr) {
	h, err:= ParseHeader(req.Content)
	if err ==  nil {
		protocol := fmt.Sprintf("Type_%d", h.Protocol)
		s.AddStats(protocol, 1)
		link := fmt.Sprintf("%v->%v", h.Src, h.Dst)
		s.AddStats(link, 1)
		bf, err := req.Encode()
		if err == nil {
			if s.Connection != nil {
				s.Connection.Write(bf)
			}
			s.AddStats("CorrectGTPPacket", 1)
		} else {
			if s.Connection != nil {
				s.Connection.Write(bf)
			}
			s.AddStats("IncorrectGTPPacket", 1)
		}
	}
}

func (s *GTPv1uStatsService) WrongGTPPacketHandle(b * []byte, conn * net.UDPConn, addr * net.Addr) {
	s.AddStats("WrongGTPPacket", 1)
	if s.Connection != nil {
		s.Connection.Write(*b)
	}
}

func (s *GTPv1uStatsService) PrintReport() {
	s.lock.Lock()
	for k,v := range s.stats {
		fmt.Printf("%30s : %d\n", k, v)
	}
	s.stats = map[string]uint32{}
	s.lock.Unlock()

}

func ToInt(s string, defaultValue int) int {
	i, err := strconv.ParseUint(s, 10, 0)
	if err != nil {
		return defaultValue
	}
	return int(i)
}

func startServers(eNodeBAddress string, eNodeBPort int16, sGWAddress string, sGWPort int16, reportPeriod uint16) (eNodeBService, sGWService *GTPv1uStatsService, err error) {
	sGWServiceHandle := NewGTPv1uStatsService(reportPeriod)
	err = sGWServiceHandle.Dial(eNodeBAddress, eNodeBPort)
	if err != nil {
		return nil, nil, err
	}
	eNodeBServiceHandle := NewGTPv1uStatsService(reportPeriod)
	err = eNodeBServiceHandle.Dial(sGWAddress, sGWPort)
	if err != nil {
		return nil, nil, err
	}

	ePCServer := gtpv1u.NewServer(fmt.Sprintf(":%d", eNodeBPort), sGWServiceHandle)
	eNodeServer := gtpv1u.NewServer(fmt.Sprintf(":%d", sGWPort), eNodeBServiceHandle)

	go func(){
		ePCServer.ListenAndServe()
	}()

	go func(){
		eNodeServer.ListenAndServe()
	}()

	return eNodeBServiceHandle, sGWServiceHandle, nil
}

func realMain() (ret int){
	var (
		err error
	)

	defer func() {
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			ret = 1
		}
	}()

	arguments, err := docopt.Parse(USAGE, nil, true, version, false)
	if err != nil {
		return
	}

	eNodeBPort := ToInt(arguments["--eNodeBPort"].(string), 0)
	if eNodeBPort == 0 {
		err = fmt.Errorf("invalid eNodeB  port: %s", arguments["--eNodeBPort"].(string))
		return
	}
	sGWPort := ToInt(arguments["--sGWPort"].(string), 0)
	if sGWPort == 0 {
		err = fmt.Errorf("invalid sGW port: %s", arguments["--sGWPort"].(string))
		return
	}

	eNodeBAddress := arguments["--eNodeBAddress"].(string)
	sGWAddress := arguments["--sGWAddress"].(string)

	reportPeriod := ToInt(arguments["--reportPeriod"].(string), 0)
	if reportPeriod == 0 {
		err = fmt.Errorf("invalid report period: %s", arguments["--reportPeriod"].(string))
		return
	}

	fmt.Printf(">>> Starting the S1U-Stats(ReportPeriod=%d)\n\teNodeB(%s:%d) <-> S1U-Stats <-> sGW(%s:%d)\n",
					reportPeriod, eNodeBAddress, eNodeBPort, sGWAddress, sGWPort)

	eNodeBService, sGWService, err := startServers(eNodeBAddress,  int16(eNodeBPort), sGWAddress, int16(sGWPort), uint16(reportPeriod))

	if err != nil {
		err = fmt.Errorf("Error in starting Servers")
		return
	}

	startTime := time.Now()

	for {
		time.Sleep(time.Duration(reportPeriod) * time.Second)
		fmt.Printf("\n")
		p := time.Since(startTime)
		fmt.Printf("Report eNodeB [%v]--------------------------\n", p)
		eNodeBService.PrintReport()
		fmt.Println("Report sGW ---")
		sGWService.PrintReport()

	}

	return
}

func main() {
	os.Exit(realMain())
}

