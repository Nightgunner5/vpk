package vpk

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

func readString(reader *bufio.Reader, buf *[]byte) (string, error) {
	*buf = (*buf)[:0]

	for {
		c, err := reader.ReadByte()
		if err != nil {
			return "", err
		}

		if c == '\000' {
			return string(*buf), nil
		}

		*buf = append(*buf, c)
	}

	panic("This line is unreachable unless the universe is imploding. The fact that you reached it leads me to believe the universe is actually imploding, which should be accompanied by activities other than reading Go source code.")
}

// Returns map[extension][path][filename]FileInfo
func readTree(reader *bufio.Reader) (map[string]map[string]map[string]*FileInfo, error) {
	buf := make([]byte, 0, 64)

	fileTree := make(map[string]map[string]map[string]*FileInfo)

	var err error
	var extension, path, filename string

	for err == nil {
		extension, err = readString(reader, &buf)
		if extension == "" || err != nil {
			break
		}

		for err == nil {
			path, err = readString(reader, &buf)
			if path == "" || err != nil {
				break
			}

			for err == nil {
				filename, err = readString(reader, &buf)
				if filename == "" || err != nil {
					break
				}

				err = readFileInfo(fileTree, extension, path, filename, reader)
			}
		}
	}

	if err != nil {
		return nil, err
	}

	return fileTree, nil
}

type fileInfo struct {
	CRC           uint32
	PreloadLength uint16

	ArchiveIndex uint16 // If -1, the data is in this file with the offset starting from the end of the header.

	EntryOffset uint32

	EntryLength uint32 // Number of bytes not included in PreloadData.

	Terminator uint16 // Always 0xffff
}

type FileInfo struct {
	crc uint32

	preload []byte

	archive uint16

	offset, length uint32

	extension, path, filename string
}

func readFileInfo(fileTree map[string]map[string]map[string]*FileInfo, extension, path, filename string, reader io.Reader) error {
	var info fileInfo
	err := binary.Read(reader, binary.LittleEndian, &info)
	if err != nil {
		return err
	}

	var ok bool

	var extensionTree map[string]map[string]*FileInfo
	if extensionTree, ok = fileTree[extension]; !ok {
		extensionTree = make(map[string]map[string]*FileInfo)
		fileTree[extension] = extensionTree
	}

	var pathTree map[string]*FileInfo
	if pathTree, ok = extensionTree[path]; !ok {
		pathTree = make(map[string]*FileInfo)
		extensionTree[path] = pathTree
	}

	if _, ok = pathTree[filename]; ok {
		return fmt.Errorf("Duplicate file in same tree: %s/%s.%s", path, filename, extension)
	}

	pathTree[filename] = &FileInfo{
		crc:       info.CRC,
		preload:   make([]byte, info.PreloadLength),
		archive:   info.ArchiveIndex,
		offset:    info.EntryOffset,
		length:    info.EntryLength,
		extension: extension,
		path:      path,
		filename:  filename,
	}

	if info.PreloadLength != 0 {
		_, err := io.ReadFull(reader, (*pathTree[filename]).preload)
		if err != nil {
			return err
		}
	}

	return nil
}
