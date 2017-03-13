package gtpv1u

import (
	"encoding/binary"
	"fmt"
	"net"
)

type ExtHeader struct {
	Type 		uint8
	Content 	[]byte
}

type Packet struct {
	Version           uint8
	ProtocolType      uint8
	HasSequenceNumber bool
	HasNpduNumber     bool
	MessageType       uint8
	TEID              uint32
	SequenceNumber    uint16
	NpduNumber        uint8
	ExtHeaders      	[]ExtHeader
	Content  			[]byte
}

func (a ExtHeader) Copy() ExtHeader {
	content := make([]byte, len(a.Content))
	copy(content, a.Content)
	return ExtHeader{
		Type:  a.Type,
		Content: a.Content,
	}
}

func (p *Packet) Copy() *Packet {
	outP := &Packet{
		Version:			p.Version,
		ProtocolType:  p.ProtocolType,
		HasSequenceNumber: p.HasSequenceNumber,
		HasNpduNumber: p.HasNpduNumber,
		MessageType:	p.MessageType,
		TEID:          p.TEID,
		SequenceNumber:p.SequenceNumber,
		NpduNumber:    p.NpduNumber,
	}

	outP.ExtHeaders = make([]ExtHeader, len(p.ExtHeaders))
	for i := range p.ExtHeaders {
		outP.ExtHeaders[i] = p.ExtHeaders[i].Copy()
	}
	return outP
}

func (p *Packet) Encode() (b []byte, err error) {
	totalLen := 8 + len(p.Content)
	if p.HasSequenceNumber || p.HasNpduNumber || len(p.ExtHeaders) > 0 {
		totalLen += 4
	}
	for i := range p.ExtHeaders {
		totalLen += (len(p.ExtHeaders[i].Content) + 2)
	}
	b = make([]byte, totalLen, totalLen)
	// Header Encode the GTP Packet Header
	b[0] |= (p.Version << 5)
	b[0] |= ((p.ProtocolType&0x01) << 4)
	if len(p.ExtHeaders) > 0 {
		b[0] |= 0x04
	}
	if p.HasSequenceNumber {
		b[0] |= 0x02
	}
	if p.HasNpduNumber {
		b[0] |= 0x01
	}

	b[1] = p.MessageType
	binary.BigEndian.PutUint16(b[2:4], uint16(totalLen-8))
	binary.BigEndian.PutUint32(b[4:8], uint32(p.TEID))
	bIndex := 8
	if p.HasSequenceNumber || p.HasNpduNumber || len(p.ExtHeaders) > 0 {
		if p.HasSequenceNumber {
			binary.BigEndian.PutUint16(b[8:10], uint16(p.SequenceNumber))
		}
		if p.HasNpduNumber {
			b[10] = p.NpduNumber
		}
		bIndex = 12
		if len(p.ExtHeaders) > 0 {
			b[11] = p.ExtHeaders[0].Type
			for i := range p.ExtHeaders {
				contentLen := (len(p.ExtHeaders[i].Content) + 2) / 4
				b[bIndex] = uint8(contentLen)
				copy(b[bIndex + 1 : bIndex+1+len(p.ExtHeaders[i].Content)], p.ExtHeaders[i].Content)
				if len(p.ExtHeaders) > i + 1{
					b[bIndex+1+len(p.ExtHeaders[i].Content)] = p.ExtHeaders[i+1].Type
				}
				bIndex += (1+len(p.ExtHeaders[i].Content) + 1)
			}
		}
	}

	//fmt.Printf("Let copy the content %d bytes, bindex=%d, bufferSize=%d\n", len(p.Content), bIndex, totalLen)
	copy(b[bIndex:bIndex+len(p.Content)], p.Content)
	return b, nil
}

func DecodePacket(b []byte) (p * Packet, err error) {
	if len(b) < 8 {
		return nil, fmt.Errorf("Packet too small (less than 8 bytes) ")
	}
	p = &Packet{}
	p.Version = ((b[0] >> 5) & 0x03)
	p.ProtocolType = (b[0] >> 4) & 0x01
	hasExtension := false
	if ((b[0] >> 2) & 0x01) > 0 {
		hasExtension = true
	}
	if ((b[0] >> 1) & 0x01) > 0 {
		p.HasSequenceNumber = true
	}
	if (b[0] & 0x01) > 0 {
		p.HasNpduNumber = true
	}
	headerLen := 8
	if hasExtension || p.HasSequenceNumber || p.HasNpduNumber {
		headerLen = 12
		if len(b) < 12 {
			return nil, fmt.Errorf("Packet too small (less than 12 bytes) ")
		}
	}
	p.MessageType = b[1]
	messageLen := binary.BigEndian.Uint16(b[2:4])
	if len(b) < int(messageLen + 8) {
		return nil, fmt.Errorf("Packet too small (less than %d bytes) ", messageLen + 8)
	}
	p.TEID = binary.BigEndian.Uint32(b[4:8])
	if p.HasSequenceNumber {
		p.SequenceNumber = binary.BigEndian.Uint16(b[8:10])
	}
	if p.HasNpduNumber {
		p.NpduNumber = b[10]
	}
	bIndex := headerLen
	extType := (uint8)(0)
	if hasExtension {
		if len(b) < headerLen + 1 {
			return nil, fmt.Errorf("Packet too small (less than %d bytes) ",headerLen + 1)
		}
		extType = b[headerLen-1]
		extLen := 0
		for ; hasExtension ;{
			extLen = int(b[bIndex])
			if extLen == 0 {
				return nil, fmt.Errorf("Extension Header Length is 0 ")
			}
			extLen = extLen * 4
			if len(b)-bIndex+1 < extLen  {
				return nil, fmt.Errorf("Packet too small for Extension Header %x %d %d %d %d", extType, headerLen, len(b), bIndex, extLen)
			}
			eh := ExtHeader{}
			eh.Type = extType
			eh.Content = make([]byte, extLen, extLen)
			copy(eh.Content, b[bIndex+1:bIndex+1+extLen-1])
			p.ExtHeaders = append(p.ExtHeaders, eh)
			bIndex = bIndex + extLen - 1
			extType = b[bIndex]
			bIndex = bIndex + 1
			//fmt.Printf("Has new Header type=%x len=%d contentLent=%d\n", extType, extLen, len(eh.Content))
			if extType == 0 {
				hasExtension = false
			} else {
				hasExtension = true
			}
		}
	}
	//fmt.Printf("GTP Version=%d ProtocolTyp=%d MessageType=%d MessageLength=%d, bIndex=%d len=%d\n",
	//			p.Version, p.ProtocolType, p.MessageType, messageLen, bIndex, len(b))

	extraHeader := uint16(bIndex - 8)
	if messageLen < extraHeader {
		return nil, fmt.Errorf("Extension Header Message Length is not correct ")
	} else {
		if messageLen > extraHeader {
			p.Content = make([]byte, messageLen - extraHeader, messageLen - extraHeader)
			copy(p.Content, b[bIndex:bIndex + int(messageLen - extraHeader)])
		}
	}
	return p, nil
}

func (p *Packet) Send(c net.PacketConn, addr net.Addr) error {
	buf, err := p.Encode()
	if err != nil {
		return err
	}

	_, err = c.WriteTo(buf, addr)
	return err
}

