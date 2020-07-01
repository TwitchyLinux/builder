// Package ld interprets dynamic linking configuration and caches.
package ld

// Platform describes a architecture/system combination.
type Platform uint

// Valid platform values.
const (
	PlatformX64 Platform = entryFlagsIsLibc6ELF | (entryFlagsArchX64 << 8)
)
