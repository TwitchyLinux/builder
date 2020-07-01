package ld

import (
	"errors"
	"math/bits"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// EntryFlags represents the flags on a library entry.
type EntryFlags int32

func (f EntryFlags) IsELF() bool {
	return f&entryFlagsIsELF != 0
}

func (f EntryFlags) Arch() uint8 {
	return uint8((f & entryFlagsArchMask) >> 8)
}

func (f EntryFlags) IsX64() bool {
	return f.Arch() == entryFlagsArchX64
}

func (f EntryFlags) IsX32() bool {
	return f.Arch() == entryFlagsArchX32
}

const (
	entryFlagsTypeMask   = 0xff
	entryFlagsIsELF      = 0x1
	entryFlagsIsLibc5ELF = 0x2
	entryFlagsIsLibc6ELF = 0x3

	entryFlagsArchMask = 0xff00
	entryFlagsArchX64  = 0x03
	entryFlagsArchX32  = 0x08
)

// CacheEntry represents an entry in the ld cache.
type CacheEntry struct {
	Flags  EntryFlags
	Key    string
	Val    string
	OSVers uint32
	HWCap  uint64
}

// LessThan returns true if the base entry is less than the provided entry.
func (ie CacheEntry) LessThan(je CacheEntry) bool {
	keyCmp := libNameCompare(ie.Key, je.Key)
	switch {
	case keyCmp != 0:
		return keyCmp < 0
	case ie.Flags != je.Flags:
		return ie.Flags < je.Flags
	case ie.HWCap != je.HWCap:
		ic, jc := bits.OnesCount64(ie.HWCap), bits.OnesCount64(je.HWCap)
		if ic != jc {
			return ic < jc
		}
		return ie.HWCap < je.HWCap
	case ie.OSVers != je.OSVers:
		return ie.OSVers < je.OSVers
	}
	return false
}

// Cache represents a ld.so.cache file.
type Cache struct {
	Format  CacheFormat
	Entries []CacheEntry
}

func (c *Cache) Lookup(name string, platform Platform) *CacheEntry {
	i := sort.Search(len(c.Entries), func(i int) bool {
		return libNameCompare(name, c.Entries[i].Key) >= 0
	})
	for ; i < len(c.Entries); i++ {
		if c.Entries[i].Key == name && c.Entries[i].Flags == EntryFlags(platform) {
			return &c.Entries[i]
		}
		if !strings.HasPrefix(c.Entries[i].Key, name) {
			return nil
		}
	}
	return nil
}

func libNameCompare(s1, s2 string) (out int) {
	i := 0
	for ; i < len(s1) && i < len(s2); i++ {
		c1, c2 := s1[i], s2[i]
		c1Num, c2Num := c1 >= '0' && c1 <= '9', c2 >= '0' && c2 <= '9'

		switch {
		case c1Num && c2Num:
			n1, n2 := "", ""
			startIdx := i
			for ; i < len(s1); i++ {
				if s1[i] >= '0' && s1[i] <= '9' {
					n1 += string(s1[i])
				} else {
					break
				}
			}

			for i = startIdx; i < len(s2); i++ {
				if s2[i] >= '0' && s2[i] <= '9' {
					n2 += string(s2[i])
				} else {
					break
				}
			}
			i--
			d1, _ := strconv.Atoi(n1)
			d2, _ := strconv.Atoi(n2)
			if d1 != d2 {
				return d1 - d2
			}
		case c1Num: // c1 not c2.
			return 1
		case c2Num: // c2 not c1.
			return -1
		case c1 != c2:
			return int(c1) - int(c2)
		}
	}

	if len(s1) > len(s2) {
		return int(s1[i])
	} else if len(s2) > len(s1) {
		return int(s2[i])
	}
	return 0
}

func checkGlibc1v1Order(entries []CacheEntry) error {
	dupe := make([]CacheEntry, len(entries))
	copy(dupe, entries)
	sort.SliceStable(dupe, func(i int, j int) bool {
		ie, je := dupe[j], dupe[i] // reversed
		return ie.LessThan(je)
	})

	// fmt.Println(cmp.Diff(dupe, entries))
	if !reflect.DeepEqual(dupe, entries) {
		return errors.New("cache had invalid index order")
	}
	return nil
}
