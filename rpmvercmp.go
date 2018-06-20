package repos

import (
	"bytes"
)

//  The basic way rpm works is to take a string and split it into segments
// of alpha/numeric/tilda/other chars. All the "other" parts are considered
// equal. A tilda is newer. , the others are compared
// as normal strings ... but numbers are kinda special because leading zeros
// are ignored.

const tilda = true

type versionT int

const (
	vT_ALP versionT = iota
	vT_NUM
	vT_TIL
	vT_MSC
	vT_END
)

func getTbyte(d byte) versionT {
	id := int(d)
	switch {
	case id >= int('0') && id <= int('9'):
		return vT_NUM
	case id >= int('a') && id <= int('z'):
		return vT_ALP
	case id >= int('A') && id <= int('Z'):
		return vT_ALP
	case id == int('~'):
		if tilda {
			return vT_TIL
		}
		fallthrough
	default:
		return vT_MSC
	}
}

func nextSlice(d []byte) ([]byte, versionT, []byte) {
	if len(d) <= 0 {
		return d, vT_END, d
	}

	t := getTbyte(d[0])

	for num := 1; num < len(d); num++ {
		if getTbyte(d[num]) != t {
			return d[:num], t, d[num:]
		}
	}
	return d, t, []byte{}
}

func rpmvercmpBytes(s1, s2 []byte) int {
	for len(s1) > 0 || len(s2) > 0 {
		cs1, t1, ns1 := nextSlice(s1)
		if t1 == vT_MSC {
			cs1, t1, ns1 = nextSlice(ns1)
		}
		cs2, t2, ns2 := nextSlice(s2)
		if t2 == vT_MSC {
			cs2, t2, ns2 = nextSlice(ns2)
		}

		s1 = ns1
		s2 = ns2

		if t1 == vT_TIL || t2 == vT_TIL {
			if t1 == vT_TIL && t2 == vT_TIL {
				if len(cs1) < len(cs2) {
					return 1
				}
				if len(cs1) > len(cs2) {
					return -1
				}
				continue
			}

			if t1 == vT_TIL {
				return -1
			}
			if t2 == vT_TIL {
				return 1
			}
		}

		if t1 != t2 {
			if t1 == vT_END {
				return -1
			}
			if t2 == vT_END {
				return 1
			}
			if t1 == vT_ALP && t2 == vT_NUM {
				return -1
			}
			if t2 == vT_ALP && t1 == vT_NUM {
				return 1
			}
			if t1 == vT_MSC {
				return -1
			}
			return 1
		}

		if t1 == vT_MSC {
			continue
		}

		if t1 == vT_NUM {

			for len(cs1) > 0 && cs1[0] == '0' {
				cs1 = cs1[1:]
			}
			for len(cs2) > 0 && cs2[0] == '0' {
				cs2 = cs2[1:]
			}
			if len(cs1) < len(cs2) {
				return -1
			}
			if len(cs1) > len(cs2) {
				return 1
			}
		}

		ret := bytes.Compare(cs1, cs2)
		if ret != 0 {
			return ret
		}
	}

	return 0
}

func rpmvercmp(s1, s2 string) int {
	return rpmvercmpBytes([]byte(s1), []byte(s2))
}
