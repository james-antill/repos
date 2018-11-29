package repos

import (
	"bytes"
)

//  The basic way rpm works is to take a string and split it into segments
// of alpha/numeric/tilde/other chars. All the "other" parts are considered
// equal. A tilde is newer. , the others are compared
// as normal strings ... but numbers are kinda special because leading zeros
// are ignored.

const tilde = true
const caret = true

type versionT int

const (
	vT_ALP versionT = iota
	vT_NUM
	vT_TIL
	vT_CAR
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
		if tilde {
			return vT_TIL
		}
		return vT_MSC
	case id == int('^'):
		if caret {
			return vT_CAR
		}
		return vT_MSC
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

// T_MSC slices aren't useful, they just split things that don't split on their own.
// own. See rpmvercmp testcases like TestRpmvercmpOdd
// tl;dr 1.x == 1x
func nextUsefulSlice(d []byte) ([]byte, versionT, []byte) {
	cs1, t1, s1 := nextSlice(d)
	if t1 == vT_MSC {
		cs1, t1, s1 = nextSlice(s1)
	}

	return cs1, t1, s1
}

func rpmvercmpBytes(s1, s2 []byte) int {
	for len(s1) > 0 || len(s2) > 0 {
		cs1, t1, shadows1 := nextUsefulSlice(s1)
		s1 = shadows1
		cs2, t2, shadows2 := nextUsefulSlice(s2)
		s2 = shadows2

		// Tilde sections mean it's older...
		if t1 == vT_TIL || t2 == vT_TIL {
			if t1 == vT_TIL && t2 == vT_TIL {
				if len(cs1) == len(cs2) {
					continue
				}
				if len(cs1) < len(cs2) {
					cs1, t1, s1 = nextUsefulSlice(s1)
				} else {
					cs2, t2, s2 = nextUsefulSlice(s2)
				}
			}

			if t1 == vT_TIL {
				return -1
			}
			if t2 == vT_TIL {
				return 1
			}
		}
		// Caret is almost the same as tilde,
		// differs when it's the end of the string...
		if t1 == vT_CAR || t2 == vT_CAR {
			if t1 == vT_CAR && t2 == vT_CAR {
				if len(cs1) == len(cs2) {
					continue
				}
				if len(cs1) < len(cs2) {
					cs1, t1, s1 = nextUsefulSlice(s1)
				} else {
					cs2, t2, s2 = nextUsefulSlice(s2)
				}
			}

			if t1 == vT_CAR && t2 == vT_END {
				return 1
			}
			if t2 == vT_CAR && t1 == vT_END {
				return -1
			}

			if t1 == vT_CAR {
				return -1
			}
			if t2 == vT_CAR {
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
