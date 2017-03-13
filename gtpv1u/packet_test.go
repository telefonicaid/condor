package gtpv1u

import (
	"testing"
	"encoding/binary"
	"fmt"
)

func TestPacket(t *testing.T) {
	inBytes := []byte{0x30, 0xff, 0x00, 0x4b, 0x00, 0x07, 0xa1, 0x2d, 0x45, 0x00, 0x00, 0x4b,
		0x8c, 0x4a, 0x40, 0x00, 0x40, 0x11, 0x03, 0x97, 0xac, 0x15, 0x14, 0x02,
		0xc0, 0xa8, 0x2a, 0x01, 0x05, 0xcd, 0x00, 0x35, 0x00, 0x37, 0x0b, 0xc3,
		0x01, 0xab, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x11, 0x63, 0x6f, 0x6e, 0x6e, 0x65, 0x63, 0x74, 0x69, 0x76, 0x69, 0x74,
		0x79, 0x63, 0x68, 0x65, 0x63, 0x6b, 0x07, 0x67, 0x73, 0x74, 0x61, 0x74,
		0x69, 0x63, 0x03, 0x63, 0x6f, 0x6d, 0x00, 0x00, 0x1c, 0x00, 0x01}

	p, err := DecodePacket(inBytes)
	if err != nil {
		t.Fatalf("Error in parsing the packet: %s", err)
	}
	if p.Version != 1 {
		t.Fatalf("Version incorrect")
	}
	if p.ProtocolType != 1 {
		t.Fatalf("ProtocolType incorrect")
	}
	if p.MessageType != 0xff {
		t.Fatalf("ProtocolType incorrect")
	}
	if p.HasSequenceNumber {
		t.Fatalf("It should not has sequence number")
	}
	if p.HasNpduNumber {
		t.Fatalf("It should not has npdu number")
	}
	if len(p.ExtHeaders) != 0 {
		t.Fatalf("It should not has Extension headers")
	}
	teid := []byte{0x00, 0x07, 0xa1, 0x2d}
	if p.TEID !=  binary.BigEndian.Uint32(teid){
		t.Fatalf("TEID incorrect %x vs %x ", p.TEID, binary.BigEndian.Uint32(teid))
	}
	if len(p.Content) != len(inBytes)-8 {
		t.Fatalf("Content length incorrect %d vs %d ", len(p.Content), len(inBytes)-8)
	}
	for i := 0 ; i < (len(inBytes)-8) ; i++  {
		if p.Content[i] != inBytes[i+8] {
			t.Fatalf("Content is not the same")
		}
	}
	outBytes, err := p.Encode()
	if err != nil {
		t.Fatalf("Resulted packet can't be encoded again")
	}
	if len(outBytes) != len(inBytes) {
		t.Fatalf("Resulted bytes don't have the same length %d vs %d", len(outBytes), len(inBytes))
	}

	for i:= range outBytes {
		if outBytes[i] != inBytes[i] {
			t.Fatalf("Content is not the same")
		}
	}

}

func TestPacketWithSequenceNumber(t *testing.T) {
	inBytes := []byte{0x32, 0x10, 0x00, 0x97, 0x00, 0x00, 0x00, 0x00, 0x00, 0xfe, 0x00, 0x00, 0x02, 0x72, 0x02, 0xf3,
							0x01, 0x00, 0x00, 0x00, 0x00, 0x03, 0x72, 0xf2, 0x30, 0xff, 0xfe, 0xff, 0x0e, 0x5d, 0x0f, 0xfc,
							0x10, 0x37, 0x2f, 0x00, 0x00, 0x11, 0x37, 0x2f, 0x00, 0x00, 0x14, 0x05, 0x80, 0x00, 0x02, 0xf1,
							0x21, 0x83, 0x00, 0x10, 0x03, 0x6d, 0x6d, 0x73, 0x08, 0x6d, 0x79, 0x6d, 0x65, 0x74, 0x65, 0x6f,
							0x72, 0x02, 0x69, 0x65, 0x84, 0x00, 0x22, 0x80, 0xc0, 0x23, 0x0b, 0x01, 0x00, 0x00, 0x0b, 0x02,
							0x6d, 0x79, 0x03, 0x77, 0x61, 0x70, 0x80, 0x21, 0x10, 0x01, 0x00, 0x00, 0x10, 0x81, 0x06, 0x00,
							0x00, 0x00, 0x00, 0x83, 0x06, 0x00, 0x00, 0x00, 0x00, 0x85, 0x00, 0x04, 0xd4, 0x81, 0x41, 0x0d,
							0x85, 0x00, 0x04, 0xd4, 0x81, 0x41, 0x17, 0x86, 0x00, 0x07, 0x91, 0x53, 0x83, 0x00, 0x00, 0x00,
							0x00, 0x87, 0x00, 0x0c, 0x02, 0x23, 0x62, 0x1f, 0x93, 0x96, 0x58, 0x58, 0x74, 0xfb, 0xff, 0xff,
							0x97, 0x00, 0x01, 0x02, 0x9a, 0x00, 0x08, 0x53, 0x89, 0x01, 0x10, 0x58, 0x98, 0x53, 0x21}

	p, err := DecodePacket(inBytes)
	if err != nil {
		t.Fatalf("Error in parsing the packet: %s", err)
	}
	if p.Version != 1 {
		t.Fatalf("Version incorrect")
	}
	if p.ProtocolType != 1 {
		t.Fatalf("ProtocolType incorrect")
	}
	if p.MessageType != 0x10 {
		t.Fatalf("ProtocolType incorrect")
	}
	if !p.HasSequenceNumber {
		t.Fatalf("It should has sequence number")
	}
	if p.SequenceNumber != 0x00fe {
		t.Fatalf("The sequence number should be %x, but it is %x", 0xaabb, p.SequenceNumber)
	}
	if p.HasNpduNumber {
		t.Fatalf("It should not has npdu number")
	}
	if len(p.ExtHeaders) != 0 {
		t.Fatalf("It should not has Extension headers")
	}
	teid := []byte{0x00, 0x00, 0x00, 0x00}
	if p.TEID !=  binary.BigEndian.Uint32(teid){
		t.Fatalf("TEID incorrect %x vs %x ", p.TEID, binary.BigEndian.Uint32(teid))
	}
	if len(p.Content) != len(inBytes)-12 {
		t.Fatalf("Content length incorrect %d vs %d ", len(p.Content), len(inBytes)-12)
	}
	for i := 0 ; i < (len(inBytes)-12) ; i++  {
		if p.Content[i] != inBytes[i+12] {
			t.Fatalf("Content is not the same")
		}
	}
	outBytes, err := p.Encode()
	if err != nil {
		t.Fatalf("Resulted packet can't be encoded again")
	}
	if len(outBytes) != len(inBytes) {
		t.Fatalf("Resulted bytes don't have the same length %d vs %d", len(outBytes), len(inBytes))
	}

	for i:= range outBytes {
		if outBytes[i] != inBytes[i] {
			t.Fatalf("Content is not the same in pos=%d %x vs %x", i, outBytes[i], inBytes[i])
		}
	}
}

func printByte(outBytes []byte) {
	for i:=0 ; i<len(outBytes); {
		fmt.Printf("%2d: ", i)
		for j:=0 ; j<4 && i<len(outBytes) ; j++ {
			fmt.Printf("0x%02x, ", outBytes[i])
			i ++
		}
		fmt.Println("")
	}
}

func TestPacketAddSomeExtensionHeaders(t *testing.T) {
	inBytes := []byte{0x30, 0xff, 0x00, 0x4b,
							0x00, 0x07, 0xa1, 0x2d,
							0x45, 0x00, 0x00, 0x4b,
							0x8c, 0x4a, 0x40, 0x00,
							0x40, 0x11, 0x03, 0x97,
							0xac, 0x15, 0x14, 0x02,
							0xc0, 0xa8, 0x2a, 0x01,
							0x05, 0xcd, 0x00, 0x35,
							0x00, 0x37, 0x0b, 0xc3,
							0x01, 0xab, 0x01, 0x00,
							0x00, 0x01, 0x00, 0x00,
							0x00, 0x00, 0x00, 0x00,
							0x11, 0x63, 0x6f, 0x6e,
							0x6e, 0x65, 0x63, 0x74,
							0x69, 0x76, 0x69, 0x74,
							0x79, 0x63, 0x68, 0x65,
							0x63, 0x6b, 0x07, 0x67,
							0x73, 0x74, 0x61, 0x74,
							0x69, 0x63, 0x03, 0x63,
							0x6f, 0x6d, 0x00, 0x00,
							0x1c, 0x00, 0x01}

	extHeader := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}

	p, err := DecodePacket(inBytes)
	p.ExtHeaders = append(p.ExtHeaders, ExtHeader{0x04, extHeader})
	outBytes, err := p.Encode()

	//printByte(outBytes)

	p1, err := DecodePacket(outBytes)
	if err != nil {
		t.Fatalf("Error in parsing the packet: %s", err)
	}
	if p1.Version != 1 {
		t.Fatalf("Version incorrect")
	}
	if p1.ProtocolType != 1 {
		t.Fatalf("ProtocolType incorrect")
	}
	if p1.MessageType != 0xff {
		t.Fatalf("ProtocolType incorrect")
	}
	if p1.HasSequenceNumber {
		t.Fatalf("It should not has sequence number")
	}
	if p1.HasNpduNumber {
		t.Fatalf("It should not has npdu number")
	}
	if len(p.ExtHeaders) != 1 {
		t.Fatalf("The Extension header has not correct length")
	}
	if len(p.ExtHeaders[0].Content) != len(extHeader) {
		t.Fatalf("It should not has Extension headers")
	}

	for i:=range extHeader {
		//fmt.Printf("%v vs %v %d\n", extHeader, p.ExtHeaders[0].Content, i)
		if p1.ExtHeaders[0].Content[i] != extHeader[i] {
			t.Fatalf("The Extension header is not correct")
		}
	}
	teid := []byte{0x00, 0x07, 0xa1, 0x2d}
	if p.TEID !=  binary.BigEndian.Uint32(teid){
		t.Fatalf("TEID incorrect %x vs %x ", p.TEID, binary.BigEndian.Uint32(teid))
	}
	if len(p.Content) != len(inBytes)-8 {
		t.Fatalf("Content length incorrect %d vs %d ", len(p.Content), len(inBytes)-8)
	}
	for i := 0 ; i < (len(inBytes)-8) ; i++  {
		if p.Content[i] != inBytes[i+8] {
			t.Fatalf("Content is not the same")
		}
	}
	//TODO: Test with a real packet with Extension Headers
}



