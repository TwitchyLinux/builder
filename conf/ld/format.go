package ld

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unsafe"
)

const (
	libc5Magic     = "ld.so-1.7.0\x00"
	glibcMagic     = "glibc-ld.so.cache"
	glibc11Version = "1.1"
	glibc1v1Magic  = glibcMagic + glibc11Version
)

// CacheFormat describes the format of the shared library cache file.
type CacheFormat uint8

// Valid CacheFormat values.
const (
	UnknownFormat CacheFormat = iota
	Libc5Format
	Glibc11Format // Format version 1.1, introduced in glibc 2.2.
)

type libc5Entry struct {
	Flags int32
	K, V  uint32
}

type glibc1v1Hdr struct {
	NumLibs    uint32
	LenStrings uint32
	_          [5]uint32 // unused, for alignment
}

type glibc1v1Entry struct {
	Flags  int32
	KeyIdx uint32
	ValIdx uint32
	OSVers uint32
	HWCap  uint64
}

func parseGlibc1v1(r io.ReadSeeker, glibcSectionOffset int64, numLibs, lenStrings uint32) (*Cache, error) {
	entries := make([]glibc1v1Entry, int(numLibs))
	for i := 0; i < int(numLibs); i++ {
		if err := binary.Read(r, binary.LittleEndian, &entries[i]); err != nil {
			return nil, fmt.Errorf("reading glibc1v1 lib %d: %v", i, err)
		}
	}

	out := make([]CacheEntry, int(numLibs))
	for i := 0; i < int(numLibs); i++ {
		r.Seek(glibcSectionOffset+int64(entries[i].KeyIdx), io.SeekStart)
		key, err := bufio.NewReader(r).ReadBytes('\x00')
		if err != nil {
			return nil, fmt.Errorf("reading string key for entry %d: %v", i, err)
		}
		r.Seek(glibcSectionOffset+int64(entries[i].ValIdx), io.SeekStart)
		val, err := bufio.NewReader(r).ReadBytes('\x00')
		if err != nil {
			return nil, fmt.Errorf("reading string key for entry %d: %v", i, err)
		}

		out[i] = CacheEntry{
			Flags: EntryFlags(entries[i].Flags),
			HWCap: entries[i].HWCap,
			Key:   strings.TrimSuffix(string(key), "\x00"),
			Val:   strings.TrimSuffix(string(val), "\x00"),
		}
	}

	if err := checkGlibc1v1Order(out); err != nil {
		return nil, err
	}

	return &Cache{
		Format:  Glibc11Format,
		Entries: out,
	}, nil
}

// ParseCache parses an ld.so.cache file.
func ParseCache(r io.ReadSeeker) (*Cache, error) {
	var glibcCheckOffset int64
	var checkLibc5Magic [len(libc5Magic)]byte
	if n, err := r.Read(checkLibc5Magic[:]); err != nil || n != len(checkLibc5Magic) {
		return nil, fmt.Errorf("reading magic: %v", err)
	}

	var numLibc5Sections uint32
	if string(checkLibc5Magic[:]) == libc5Magic {
		if err := binary.Read(r, binary.LittleEndian, &numLibc5Sections); err != nil {
			return nil, err
		}
		glibcCheckOffset = int64(len(libc5Magic)+int(unsafe.Sizeof(numLibc5Sections))) + int64(numLibc5Sections)*int64(unsafe.Sizeof(libc5Entry{}))
	}

	var checkGlibcMagic [len(glibc1v1Magic)]byte
	r.Seek(glibcCheckOffset, io.SeekStart)
	if n, err := r.Read(checkGlibcMagic[:]); err != nil || n != len(glibc1v1Magic) {
		return nil, fmt.Errorf("reading glibc magic: %v", err)
	}
	if string(checkGlibcMagic[:]) == glibc1v1Magic {
		var hdr glibc1v1Hdr
		if err := binary.Read(r, binary.LittleEndian, &hdr); err != nil {
			return nil, fmt.Errorf("reading glibc1v1 header: %v", err)
		}
		return parseGlibc1v1(r, glibcCheckOffset, hdr.NumLibs, hdr.LenStrings)
	}

	if numLibc5Sections > 0 {
		return nil, errors.New("libc5 format parser not implemented")
	}

	return nil, os.ErrNotExist
}
