package proxmox

import (
	"crypto/sha1"
	"strconv"
)

// stores a SHA1 digest as a 40-character hexadecimal string
type digest string

func (d digest) String() string {
	return string(d)
}

func (d digest) sha1() [sha1.Size]byte {
	if len(d) != sha1.Size*2 {
		return [sha1.Size]byte{}
	}
	var digest [sha1.Size]byte
	for i := range sha1.Size {
		// Convert hex to byte
		b, _ := strconv.ParseUint(string(d[i*2])+string(d[i*2+1]), 16, 8)
		digest[i] = byte(b)
	}
	return digest
}
