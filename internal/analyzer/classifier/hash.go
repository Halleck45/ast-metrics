package classifier

// MurmurHash3 x86 32-bit (seed = 0), compatible avec sklearn HashingVectorizer.
func murmur3_32(s string) uint32 {
	const (
		c1 uint32 = 0xcc9e2d51
		c2 uint32 = 0x1b873593
	)

	data := []byte(s)
	h1 := uint32(0)

	nblocks := len(data) / 4
	// body
	for i := 0; i < nblocks; i++ {
		j := i * 4
		k1 := uint32(data[j]) | uint32(data[j+1])<<8 | uint32(data[j+2])<<16 | uint32(data[j+3])<<24

		k1 *= c1
		k1 = (k1 << 15) | (k1 >> 17)
		k1 *= c2

		h1 ^= k1
		h1 = (h1 << 13) | (h1 >> 19)
		h1 = h1*5 + 0xe6546b64
	}

	// tail
	var k1 uint32
	tail := data[nblocks*4:]
	switch len(tail) {
	case 3:
		k1 ^= uint32(tail[2]) << 16
		fallthrough
	case 2:
		k1 ^= uint32(tail[1]) << 8
		fallthrough
	case 1:
		k1 ^= uint32(tail[0])
		k1 *= c1
		k1 = (k1 << 15) | (k1 >> 17)
		k1 *= c2
		h1 ^= k1
	}

	// fmix
	h1 ^= uint32(len(data))
	h1 ^= h1 >> 16
	h1 *= 0x85ebca6b
	h1 ^= h1 >> 13
	h1 *= 0xc2b2ae35
	h1 ^= h1 >> 16

	return h1
}
