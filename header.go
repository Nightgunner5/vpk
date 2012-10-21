package vpk

import (
	"encoding/binary"
	"fmt"
	"io"
)

type header interface {
	treeOffset() uint32
}

type headerStart struct {
	Signature uint32 // Always 0x55aa1234
	Version   uint32
}

type Header_v1 struct {
	TreeLength uint32
}

func (h Header_v1) treeOffset() uint32 {
	return h.TreeLength +
		8 /* Start */ +
		4 /* TreeLength */
}

type Header_v2 struct {
	TreeLength   uint32
	unknown1     uint32
	FooterLength uint32
	unknown2     uint32
	unknown3     uint32
}

func (h Header_v2) treeOffset() uint32 {
	return h.TreeLength +
		8 /* Start */ +
		4 /* TreeLength */ +
		4 /* unknown1 */ +
		4 /* FooterLength */ +
		4 /* unknown2 */ +
		4 /* unknown3 */
}

func readHeader(reader io.Reader) (header, error) {
	var start headerStart
	err := binary.Read(reader, binary.LittleEndian, &start)
	if err != nil {
		return nil, err
	}

	if start.Signature != 0x55aa1234 {
		return nil, fmt.Errorf("Invalid signature 0x%08x (expected 0x%08x)", start.Signature, 0x55aa1234)
	}

	var h header
	switch start.Version {
	case 1:
		h_ := Header_v1{}
		err = binary.Read(reader, binary.LittleEndian, &h_)
		h = h_
	case 2:
		h_ := Header_v2{}
		err = binary.Read(reader, binary.LittleEndian, &h_)
		h = h_
	default:
		return nil, fmt.Errorf("Unknown version %d", start.Version)
	}
	if err != nil {
		return nil, err
	}

	return h, nil
}
