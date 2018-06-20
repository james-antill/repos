package repos

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
)

func hchk(data []byte, chk Checksum) bool {
	var hash hash.Hash
	switch chk.Kind {
	case "md5":
		hash = md5.New()
	case "sha":
		fallthrough
	case "sha1":
		hash = sha1.New()
	case "sha2":
		fallthrough
	case "sha256":
		hash = sha256.New()
	case "sha512":
		hash = sha512.New()
	case "":
		panic("No hash specified")
	default:
		panic(fmt.Sprintf("Unknown hash <%s> specified", chk.Kind))
	}

	hash.Write(data)
	hval := fmt.Sprintf("%x", hash.Sum(nil))
	if hval != chk.Data {
		// fmt.Println("JDBG:", chk.Kind, hval, chk.Data)
	}
	return hval == chk.Data
}

func hchks(data []byte, chks []Checksum) bool {
	for _, chk := range chks {
		if !hchk(data, chk) {
			return false
		}
	}

	return true
}
