package main

import "hash"

//
// HashString calculates a 32 bit hash code for a supplied string and returns it
//
// NOTE:
// 	1. I have implemented the standard Go hash.Hash32 interface for general purpose hash
// 	calculations, then wrapped that in this helper function to simplify the strings case
//	2. This hash code is non-cryptographic, instead being optimised for hash tables use
//  3. This implements a standard FNV-1a hash algorithm for 32 bit hash codes
//
func HashString(str string) uint32 {
	hasher := CreateMyHash32()
	hasher.Write([]byte(str))
	return hasher.Sum32()
}

const (
	fvOffset = 2166136261
	fvPrime  = 16777619
)

// MyHash32 type implements the standard hash.Hash32 interface using FNV-1a
type MyHash32 struct {
	hashVal uint32
}

// CreateMyHash32 creates a new Hash32 implementation of the FNV-1a has algorithm
func CreateMyHash32() hash.Hash32 {
	hv := MyHash32{}
	hv.Reset()
	return &hv
}

// io.Writer interface methods

// Write takes a sequence of bytes to hash and accumulates the hash code
// Returns the number of bytes processed (all of them)
func (h *MyHash32) Write(b []byte) (int, error) {
	for _, next := range b {
		h.hashVal ^= uint32(next)
		h.hashVal *= fvPrime
	}
	return len(b), nil
}

// Sum appends the current hash to b and returns the resulting slice.
// It does not change the underlying hash state.
func (h *MyHash32) Sum(b []byte) []byte {
	return append(b, byte(h.hashVal>>24), byte(h.hashVal>>16), byte(h.hashVal>>8), byte(h.hashVal))
}

// Reset resets the Hash to its initial state.
func (h *MyHash32) Reset() {
	h.hashVal = fvOffset
}

// Size returns the number of bytes Sum will return.
func (h *MyHash32) Size() int {
	return 4
}

// BlockSize returns the hash's underlying block size.
// The Write method must be able to accept any amount
// of data, but it may operate more efficiently if all writes
// are a multiple of the block size.
func (h *MyHash32) BlockSize() int {
	return 1
}

// hash.Hash32 interface

// Sum32 returns the hash code
func (h *MyHash32) Sum32() uint32 {
	return h.hashVal
}
