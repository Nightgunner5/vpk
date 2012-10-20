package vpk

import (
	"encoding/binary"
	"fmt"
	"io"
)

type header interface {
	implementsHeader()
}

type implementsHeader struct{}

func (implementsHeader) implementsHeader() {}

type headerStart struct {
	Signature uint32 // Always 0x3412aa55
	Version   uint32
}

type Header_v1 struct {
	implementsHeader
	TreeLength uint32
}

type Header_v2 struct {
	implementsHeader
	TreeLength   uint32
	unknown1     uint32
	FooterLength uint32
	unknown2     uint32
	unknown3     uint32
}

func readHeader(reader io.Reader) (header, error) {
	var start headerStart
	err := binary.Read(reader, binary.LittleEndian, &start)
	if err != nil {
		return nil, err
	}

	if start.Signature != 0x3412aa55 {
		return nil, fmt.Errorf("Invalid signature 0x%08x (expected 0x%08x)", start.Signature, 0x3412aa55)
	}

	var h header
	switch start.Version {
	case 1:
		h = Header_v1{}
	case 2:
		h = Header_v2{}
	default:
		return nil, fmt.Errorf("Unknown version %d", start.Version)
	}
	err = binary.Read(reader, binary.LittleEndian, &h)
	if err != nil {
		return nil, err
	}

	return h, nil
}
