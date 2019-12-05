package ld

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
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

// CacheEntry represents an entry in the ld cache.
type CacheEntry struct {
	Flags  int32
	Key    string
	Val    string
	OSVers uint32
	HWCap  uint64
}

type Cache struct {
	Format  CacheFormat
	Entries []CacheEntry
}

func archAlignment(arch string) int {
	return (archIntSize(arch) * 2 * 8) - 1
}

func archIntSize(arch string) int {
	if arch == "x86" || arch == "arm" {
		return 2
	}
	return 4
}

func libc5EntrySize(arch string) int {
	// 3 'ints'.
	if arch == "x86" || arch == "arm" {
		return 6
	}
	return 12
}

func libc5NumEntries(r io.Reader, arch string) (int, error) {
	var (
		buff  [4]byte
		iSize = archIntSize(arch)
	)
	if n, err := r.Read(buff[:iSize]); err != nil || n != iSize {
		return 0, fmt.Errorf("reading number of entries: %v", err)
	}
	return int(binary.LittleEndian.Uint32(buff[:])), nil
}

func glibc11EntrySize() int {
	return 4*4 + 8
}

type glibc1v1Hdr struct {
	NumLibs    uint32
	LenStrings uint32
	Unused     [5]uint32
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
			Flags: entries[i].Flags,
			HWCap: entries[i].HWCap,
			Key:   strings.TrimSuffix(string(key), "\x00"),
			Val:   strings.TrimSuffix(string(val), "\x00"),
		}
	}

	return &Cache{
		Format:  Glibc11Format,
		Entries: out,
	}, nil
}

// ParseCache parses an ld.so.cache file.
func ParseCache(r io.ReadSeeker, arch string) (*Cache, error) {
	var glibcCheckOffset int64
	var checkLibc5Magic [len(libc5Magic)]byte
	if n, err := r.Read(checkLibc5Magic[:]); err != nil || n != len(checkLibc5Magic) {
		return nil, fmt.Errorf("reading magic: %v", err)
	}

	var (
		err              error
		numLibc5Sections int
	)
	if string(checkLibc5Magic[:]) == libc5Magic {
		if numLibc5Sections, err = libc5NumEntries(r, arch); err != nil {
			return nil, err
		}
		glibcCheckOffset = int64(len(libc5Magic) + archIntSize(arch) + numLibc5Sections*libc5EntrySize(arch))
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
