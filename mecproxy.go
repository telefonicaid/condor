// authors : Yan Grunenberger <yan.grunenberger@telefonica.com>
// title : mecproxy.go
// description : a basic proxy to prototype with MEC data

package main

import (
"github.com/golang/glog"
"os"
"flag"
"fmt"
"net"
//"github.com/songgao/water"
"github.com/telefonicaid/condor/gtpv1u"
"github.com/fkgi/extnet"
)

// command line parameters
var sctpListenaddressPtr * string
var sctpPortPtr * int
var addressPortPtr * string
var listenPortPtr * int

var mmeAddressPtr *string
var mmePortPtr * int
var sgwAddressPtr *string
var sgwPortPtr * int

var enbMMEAddressPtr *string
var enbMMEPortPtr *int
var enbAddressPtr *string
var enbPortPtr * int

// handle on connections, address points
var UDPProxyAddrPtr *net.UDPAddr 
var UDPProxyConnPtr * net.UDPConn

var localMMEAddrPtr *extnet.SCTPAddr
var localMMEListenPtr *extnet.SCTPListener

var remoteMMEAddrPtr *extnet.SCTPAddr
var enbMMEAddrPtr *extnet.SCTPAddr

var SGWAddrPtr *net.UDPAddr 
var ENBAddrPtr *net.UDPAddr 
var SGWConnPtr *net.UDPConn
var ENBConnPtr *net.UDPConn

//var TUNIntPtr *water.Interface

func usage() {
	fmt.Fprintf(os.Stderr, "usage: mecproxy -stderrthreshold=[INFO|WARN|FATAL] -log_dir=[string]\n", )
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	flag.Usage = usage

	sctpListenaddressPtr = flag.String("lsctp", "0.0.0.0", "listening interface for incoming SCTP traffic")
	sctpPortPtr = flag.Int("psctp", 36412, "listening port for incoming SCTP traffic")

	addressPortPtr = flag.String("lgtp", "0.0.0.0", "listening interface for incoming GTP traffic")
	listenPortPtr = flag.Int("pgtp", 2152, "listening port for incoming GTP traffic")

	mmeAddressPtr = flag.String("mmeaddr", "192.168.42.108", "IP address of the MME")
	mmePortPtr = flag.Int("mmeport", 36412, "listening port for incoming GTP traffic")

	enbMMEAddressPtr = flag.String("enbmmeaddr", "192.168.42.10", "IP address of the eNB for SCTP")
	enbMMEPortPtr = flag.Int("enbmmeport", 36412, "destination port on the eNB for SCTP traffic")

	sgwAddressPtr = flag.String("sgwaddr", "192.168.42.108", "IP address of S-GW")
	sgwPortPtr = flag.Int("sgwport", 2152, "UDP port for GTP traffic of S-GW")

	enbAddressPtr = flag.String("enbaddr", "192.168.42.10", "IP address of eNB")
	enbPortPtr = flag.Int("enbport", 2152, "incoming S1-U port on the eNB")
	
	flag.Parse()
}

func handleGTPPacket(b []byte)(){
	gtppacket,err := gtpv1u.DecodePacket(b)
		if err != nil {
			glog.Error("Error:",err.Error())
		}

	glog.Info("Packet ", gtppacket)


//	TUNIntPtr.Write(gtppacket.Content) 

}


func setupUDPProxy() bool {

	// local udp server
	saddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", *addressPortPtr, * listenPortPtr ))
	if err != nil {
		glog.Error("Error:",err.Error())
		return false
	}

	UDPProxyAddrPtr = saddr

	udplistener, err := net.ListenUDP("udp",saddr)
	if err != nil {
		glog.Error("Error:",err.Error())
		return false
	}

	UDPProxyConnPtr = udplistener

	// sgw
	sgwaddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d",*sgwAddressPtr,*sgwPortPtr))
	if err != nil {
		glog.Error("Error:",err.Error())
		return false
	}

	SGWAddrPtr = sgwaddr

	// eNb
	enbaddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d",*enbAddressPtr,*enbPortPtr))
	if err != nil {
		glog.Error("Error:",err.Error())
		return false
	}

	ENBAddrPtr = enbaddr

	return true;
}


func runUDPProxy(){
	var buffer [1500] byte

	for {

		glog.Info("Waiting for GTP-U/UDP packets on port ", * listenPortPtr)

		nbytes,clientaddr,err := UDPProxyConnPtr.ReadFromUDP(buffer[0:])
		if err != nil {
			glog.Error("Error:",err.Error())
			continue
		}

		glog.Info("Read ", nbytes, " from ", clientaddr)

		//func DecodePacket(b []byte) (p * Packet, err error) {
		go handleGTPPacket(buffer[0:nbytes])
		

		if clientaddr.String() == ENBAddrPtr.String() {
			glog.Info("Packet from the ENB")

			if SGWConnPtr == nil {


				sgwconn, err := net.DialUDP("udp", nil, SGWAddrPtr)
				if err != nil {
					glog.Error("Error:",err.Error())
				}
				SGWConnPtr = sgwconn

				glog.Info("Created connection toward SGW")
			}


			wnbytes, err := SGWConnPtr.Write(buffer[0:nbytes])
			if err!=nil{
				glog.Error("Error:",err.Error())
			}

			glog.Info("wrote ", wnbytes, " to ", SGWAddrPtr)

		}

		if clientaddr.String() == SGWAddrPtr.String() {
			glog.Info("Packet from the SGW")

			if ENBConnPtr == nil {
				enbconn, err := net.DialUDP("udp", nil, ENBAddrPtr)
				if err != nil {
					glog.Error("Error:",err.Error())
				}

				ENBConnPtr = enbconn 

				glog.Info("Created connection toward ENB")
			}

			wnbytes, err := ENBConnPtr.Write(buffer[0:nbytes])
			if err!=nil{
				glog.Error("Error:",err.Error())
			}

			glog.Info("wrote ", wnbytes, " to ", ENBAddrPtr)	

		}

	}
}
/*

func setupTUN() bool{
	ifce, err := water.New(water.Config{
		DeviceType: water.TUN,
	})
	if err != nil {
		glog.Error("Error:",err.Error())
		return false
	}

	TUNIntPtr = ifce

	return true
}

func runTUNproxy(){
	// TODO check the input size
	packet := make([] byte,2000) 
	for {
		glog.Info("Waiting for packet on tun interface ", TUNIntPtr.Name())
		n, err := TUNIntPtr.Read(packet)
		if err != nil {
			glog.Error("Error:",err.Error())
		}

		glog.Info("Packet received on TUN interface", packet[0:n])
	}
}*/

func setupSCTPProxy() bool {

	// local SCTP server
	sctpaddr, err := extnet.ResolveSCTPAddr("sctp", fmt.Sprintf("%s:%d",*sctpListenaddressPtr,*sctpPortPtr));
	if err!=nil {
		glog.Error("Error:",err.Error())
		return false
	}

	localMMEAddrPtr = sctpaddr

	sctplisten, err :=extnet.ListenSCTP("sctp", localMMEAddrPtr)
	if err!= nil {
		glog.Error("Error:",err.Error())
		return false
	}

	localMMEListenPtr = sctplisten


	// sgw
	sctpMMEaddr, err := extnet.ResolveSCTPAddr("sctp", fmt.Sprintf("%s:%d",*mmeAddressPtr,*mmePortPtr));
	if err != nil {
		glog.Error("Error:",err.Error())
		return false
	}

	remoteMMEAddrPtr = sctpMMEaddr

	// eNB
	sctpeNBaddr, err := extnet.ResolveSCTPAddr("sctp", fmt.Sprintf("%s:%d",*enbAddressPtr,*enbPortPtr));
	if err != nil {
		glog.Error("Error:",err.Error())
		return false
	}

	enbMMEAddrPtr = sctpeNBaddr

	return true;
}

func runSCTPProxy() {

	for {

		glog.Info("Waiting for SCTP packets on port ", * sctpPortPtr)
		sctpconn, err := localMMEListenPtr.AcceptSCTP()
		if err != nil {
			glog.Error("Error:",err.Error())
			continue
		}

		_ = sctpconn

	}

}

func main() {
	glog.Info("MECProxy by Yan Grunenberger <yan.grunenberger@telefonica.com>")



/*	if setupTUN(){
		//go runTUNproxy()
	}*/

	if (setupSCTPProxy()){
		go runSCTPProxy()
	}

	if (setupUDPProxy()){
		
		glog.Info("UDP address is ", * UDPProxyConnPtr)
		glog.Info("SGW address is ", * SGWAddrPtr)
		glog.Info("eNB address is ", * ENBAddrPtr)

		runUDPProxy()
	}

}
