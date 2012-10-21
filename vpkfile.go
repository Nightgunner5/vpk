package vpk

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

type VPKFile struct {
	treeOffset uint32 // Offset to the byte after the file tree

	fileTree map[string]map[string]map[string]*FileInfo
}

func ReadVPKFile(reader io.Reader) (*VPKFile, error) {
	r := bufio.NewReader(reader)

	header, err := readHeader(r)
	if err != nil {
		return nil, err
	}

	fileTree, err := readTree(r)
	if err != nil {
		return nil, err
	}

	return &VPKFile{
		treeOffset: header.treeOffset(),
		fileTree:   fileTree,
	}, nil
}

// NB: Empty filename components are converted to spaces. This seems to be the
// same way the official Valve VPK tool does it. This does have a side effect
// of making an empty filename equal to the filename " / . "
func (vpk VPKFile) GetFileInfo(filename string) *FileInfo {
	filenameRunes := []rune(strings.ToLower(filename))

	var index int
	for index = len(filenameRunes) - 1; index >= 0; index-- {
		if filenameRunes[index] == '.' {
			break
		}
		if filenameRunes[index] == '/' {
			index = -1
			break
		}
	}

	extension := " "
	if index >= 0 {
		extension = string(filenameRunes[index+1:])
		filenameRunes = filenameRunes[:index]
	}

	for index = len(filenameRunes) - 1; index >= 0; index-- {
		if filenameRunes[index] == '/' {
			break
		}
	}

	name := " "
	if index >= 0 {
		name = string(filenameRunes[index+1:])
		filenameRunes = filenameRunes[:index]
	} else {
		name = string(filenameRunes)
		filenameRunes = filenameRunes[:0]
	}

	path := " "
	if len(filenameRunes) > 0 {
		path = string(filenameRunes)
	}

	if extTree, ok := vpk.fileTree[extension]; ok {
		if pathTree, ok := extTree[path]; ok {
			if file, ok := pathTree[name]; ok {
				return file
			}
		}
	}
	return nil
}

func (vpk VPKFile) ListFiles() []string {
	var filenames []string
	for ext, extTree := range vpk.fileTree {
		if ext == " " {
			ext = ""
		} else {
			ext = "." + ext
		}
		for path, pathTree := range extTree {
			if path == " " {
				path = ""
			} else {
				path = path + "/"
			}
			for name, _ := range pathTree {
				filenames = append(filenames, path+name+ext)
			}
		}
	}

	return filenames
}

type readCloser struct {
	io.Reader
	io.Closer
}

func (vpk VPKFile) GetReader(info *FileInfo, filename string) (io.ReadCloser, error) {
	if info.archive != 0x7fff {
		if !strings.HasSuffix(filename, "_dir.vpk") {
			return nil, fmt.Errorf("Filename %q does not end with %q", filename, "_dir.vpk")
		}
		filename = fmt.Sprintf("%s_%03d.vpk", filename[:len(filename)-len("_dir.vpk")], info.archive)
	}
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	if info.archive == 0x7fff {
		_, err = file.Seek(int64(vpk.treeOffset)+int64(info.offset), 0)
	} else {
		_, err = file.Seek(int64(info.offset), 0)
	}
	if err != nil {
		file.Close()
		return nil, err
	}

	return readCloser{
		Reader: io.MultiReader(bytes.NewReader(info.preload), io.LimitReader(file, int64(info.length))),
		Closer: file,
	}, nil
}
