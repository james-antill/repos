package repos

import (
	"bytes"
	"testing"
)

func TestNextSlice(t *testing.T) {
	data := []struct {
		s1  []byte
		cs1 []byte
		st  versionT
		ns1 []byte
	}{
		{[]byte{'!'}, []byte{'!'}, vT_MSC, []byte{}},
		{[]byte("1.1"), []byte{'1'}, vT_NUM, []byte(".1")},
		{[]byte("10.3"), []byte("10"), vT_NUM, []byte(".3")},
		{[]byte("abcd.4"), []byte("abcd"), vT_ALP, []byte(".4")},
		{[]byte("!!abcd.5"), []byte("!!"), vT_MSC, []byte("abcd.5")},
		{[]byte("~!!abcd.6"), []byte("~"), vT_TIL, []byte("!!abcd.6")},
		{[]byte("Abcd.7"), []byte("Abcd"), vT_ALP, []byte(".7")},
		{[]byte("ABCD~XYZ"), []byte("ABCD"), vT_ALP, []byte("~XYZ")},
	}

	for i := range data {
		s1 := data[i].s1
		tcs1, tst, tns1 := nextSlice(s1)
		if bytes.Compare(data[i].cs1, tcs1) != 0 {
			t.Errorf("Fail: %s\n  rcs1=%s\n  tcs1=%s\n", string(s1),
				string(data[i].cs1), string(tcs1))
		}
		if data[i].st != tst {
			t.Errorf("Fail: %s\n  rst=%v\n  tst=%v\n", string(s1),
				data[i].st, tst)
		}
		if bytes.Compare(data[i].ns1, tns1) != 0 {
			t.Errorf("Fail: %s\n  rns1=%s\n  tns1=%s\n", string(s1),
				string(data[i].ns1), string(tns1))
		}
	}
}

// All of the tests that start with "TestRpmC"" are from:
// tests/rpmvercmp.at from https://github.com/rpm-software-management/rpm/
func tRPMVERCMP(t *testing.T, a, b string, res int) {
	ret := rpmvercmp(a, b)
	if res != ret {
		t.Errorf("rpmvercmp(%s, %s)\n  res = %d\n  ret = %d\n",
			a, b, res, ret)
	}
}
func RPMVERCMP(t *testing.T, a, b string, res int) {
	tRPMVERCMP(t, a, b, res)
	swap := res
	if res > 0 {
		swap = -1
	}
	if res < 0 {
		swap = 1
	}
	tRPMVERCMP(t, b, a, swap)
}

func TestRpmCMain(t *testing.T) {
	RPMVERCMP(t, "1.0", "1.0", 0)
	RPMVERCMP(t, "1.0", "2.0", -1)
	RPMVERCMP(t, "2.0", "1.0", 1)

	RPMVERCMP(t, "2.0.1", "2.0.1", 0)
	RPMVERCMP(t, "2.0", "2.0.1", -1)
	RPMVERCMP(t, "2.0.1", "2.0", 1)

	RPMVERCMP(t, "2.0.1a", "2.0.1a", 0)
	RPMVERCMP(t, "2.0.1a", "2.0.1", 1)
	RPMVERCMP(t, "2.0.1", "2.0.1a", -1)

	RPMVERCMP(t, "5.5p1", "5.5p1", 0)
	RPMVERCMP(t, "5.5p1", "5.5p2", -1)
	RPMVERCMP(t, "5.5p2", "5.5p1", 1)

	RPMVERCMP(t, "5.5p10", "5.5p10", 0)
	RPMVERCMP(t, "5.5p1", "5.5p10", -1)
	RPMVERCMP(t, "5.5p10", "5.5p1", 1)

	RPMVERCMP(t, "10xyz", "10.1xyz", -1)
	RPMVERCMP(t, "10.1xyz", "10xyz", 1)

	RPMVERCMP(t, "xyz10", "xyz10", 0)
	RPMVERCMP(t, "xyz10", "xyz10.1", -1)
	RPMVERCMP(t, "xyz10.1", "xyz10", 1)

	RPMVERCMP(t, "xyz.4", "xyz.4", 0)
	RPMVERCMP(t, "xyz.4", "8", -1)
	RPMVERCMP(t, "8", "xyz.4", 1)
	RPMVERCMP(t, "xyz.4", "2", -1)
	RPMVERCMP(t, "2", "xyz.4", 1)

	RPMVERCMP(t, "5.5p2", "5.6p1", -1)
	RPMVERCMP(t, "5.6p1", "5.5p2", 1)

	RPMVERCMP(t, "5.6p1", "6.5p1", -1)
	RPMVERCMP(t, "6.5p1", "5.6p1", 1)

	RPMVERCMP(t, "6.0.rc1", "6.0", 1)
	RPMVERCMP(t, "6.0", "6.0.rc1", -1)

	RPMVERCMP(t, "10b2", "10a1", 1)
	RPMVERCMP(t, "10a2", "10b2", -1)

	RPMVERCMP(t, "1.0aa", "1.0aa", 0)
	RPMVERCMP(t, "1.0a", "1.0aa", -1)
	RPMVERCMP(t, "1.0aa", "1.0a", 1)

	RPMVERCMP(t, "10.0001", "10.0001", 0)
	RPMVERCMP(t, "10.0001", "10.1", 0)
	RPMVERCMP(t, "10.1", "10.0001", 0)
	RPMVERCMP(t, "10.0001", "10.0039", -1)
	RPMVERCMP(t, "10.0039", "10.0001", 1)

	RPMVERCMP(t, "4.999.9", "5.0", -1)
	RPMVERCMP(t, "5.0", "4.999.9", 1)

	RPMVERCMP(t, "20101121", "20101121", 0)
	RPMVERCMP(t, "20101121", "20101122", -1)
	RPMVERCMP(t, "20101122", "20101121", 1)

	RPMVERCMP(t, "2_0", "2_0", 0)
	RPMVERCMP(t, "2.0", "2_0", 0)
	RPMVERCMP(t, "2_0", "2.0", 0)
}
func TestRpmCBZ178798(t *testing.T) {
	RPMVERCMP(t, "a", "a", 0)
	RPMVERCMP(t, "a+", "a+", 0)
	RPMVERCMP(t, "a+", "a_", 0)
	RPMVERCMP(t, "a_", "a+", 0)
	RPMVERCMP(t, "+a", "+a", 0)
	RPMVERCMP(t, "+a", "_a", 0)
	RPMVERCMP(t, "_a", "+a", 0)
	RPMVERCMP(t, "+_", "+_", 0)
	RPMVERCMP(t, "_+", "+_", 0)
	RPMVERCMP(t, "_+", "_+", 0)
	RPMVERCMP(t, "+", "_", 0)
	RPMVERCMP(t, "_", "+", 0)
}
func TestRpmCTilde(t *testing.T) {
	if !tilde {
		t.SkipNow()
	}
	RPMVERCMP(t, "1.0~rc1", "1.0~rc1", 0)
	RPMVERCMP(t, "1.0~rc1", "1.0", -1)
	RPMVERCMP(t, "1.0", "1.0~rc1", 1)
	RPMVERCMP(t, "1.0~rc1", "1.0~rc2", -1)
	RPMVERCMP(t, "1.0~rc2", "1.0~rc1", 1)
	RPMVERCMP(t, "1.0~rc1~git123", "1.0~rc1~git123", 0)
	RPMVERCMP(t, "1.0~rc1~git123", "1.0~rc1", -1)
	RPMVERCMP(t, "1.0~rc1", "1.0~rc1~git123", 1)
}

func TestRpmCCaret(t *testing.T) {
	if !caret {
		t.SkipNow()
	}

	RPMVERCMP(t, "1.0^", "1.0^", 0)
	RPMVERCMP(t, "1.0^", "1.0", 1)
	RPMVERCMP(t, "1.0", "1.0^", -1)
	RPMVERCMP(t, "1.0^git1", "1.0^git1", 0)
	RPMVERCMP(t, "1.0^git1", "1.0", 1)
	RPMVERCMP(t, "1.0", "1.0^git1", -1)
	RPMVERCMP(t, "1.0^git1", "1.0^git2", -1)
	RPMVERCMP(t, "1.0^git2", "1.0^git1", 1)
	RPMVERCMP(t, "1.0^git1", "1.01", -1)
	RPMVERCMP(t, "1.01", "1.0^git1", 1)
	RPMVERCMP(t, "1.0^20160101", "1.0^20160101", 0)
	RPMVERCMP(t, "1.0^20160101", "1.0.1", -1)
	RPMVERCMP(t, "1.0.1", "1.0^20160101", 1)
	RPMVERCMP(t, "1.0^20160101^git1", "1.0^20160101^git1", 0)
	RPMVERCMP(t, "1.0^20160102", "1.0^20160101^git1", 1)
	RPMVERCMP(t, "1.0^20160101^git1", "1.0^20160102", -1)
}

func TestRpmCCaretTilde(t *testing.T) {
	if !tilde || !caret {
		t.SkipNow()
	}

	RPMVERCMP(t, "1.0~rc1^git1", "1.0~rc1^git1", 0)
	RPMVERCMP(t, "1.0~rc1^git1", "1.0~rc1", 1)
	RPMVERCMP(t, "1.0~rc1", "1.0~rc1^git1", -1)
	RPMVERCMP(t, "1.0^git1~pre", "1.0^git1~pre", 0)
	RPMVERCMP(t, "1.0^git1", "1.0^git1~pre", 1)
	RPMVERCMP(t, "1.0^git1~pre", "1.0^git1", -1)
}

func TestRpmCOdd(t *testing.T) {
	// These are included here to document current, arguably buggy behaviors
	// for reference purposes and for easy checking against  unintended
	// behavior changes.

	// RhBug:811992 case
	RPMVERCMP(t, "1b.fc17", "1b.fc17", 0)
	RPMVERCMP(t, "1b.fc17", "1.fc17", -1)
	RPMVERCMP(t, "1.fc17", "1b.fc17", 1)
	RPMVERCMP(t, "1g.fc17", "1g.fc17", 0)
	RPMVERCMP(t, "1g.fc17", "1.fc17", 1)
	RPMVERCMP(t, "1.fc17", "1g.fc17", -1)

	// Non-ascii characters are considered equal so these are all the same, eh...
	RPMVERCMP(t, "1.1.α", "1.1.α", 0)
	RPMVERCMP(t, "1.1.α", "1.1.β", 0)
	RPMVERCMP(t, "1.1.β", "1.1.α", 0)
	RPMVERCMP(t, "1.1.αα", "1.1.α", 0)
	RPMVERCMP(t, "1.1.α", "1.1.ββ", 0)
	RPMVERCMP(t, "1.1.ββ", "1.1.αα", 0)
}

func TestRpmvercmpOdd(t *testing.T) {
	RPMVERCMP(t, "1.a.1....", "1a1", 0)
	RPMVERCMP(t, "1.a.1....", "!!1!!a!!1", 0)
}

func TestRpmvercmpBZ50977(t *testing.T) {
	RPMVERCMP(t, "1.0", "1.z", 1)
	RPMVERCMP(t, "1.0", "1.z.1", 1)
	RPMVERCMP(t, "1.0.1", "1.z", 1)
}

func TestRpmvercmpTilde(t *testing.T) {
	if !tilde {
		t.SkipNow()
	}
	RPMVERCMP(t, "1.0~~rc1", "1.0~~rc1", 0)
	RPMVERCMP(t, "1.0~~~rc1", "1.0~~~rc1", 0)
	RPMVERCMP(t, "1.0~~~~rc1", "1.0~~~~rc1", 0)
	RPMVERCMP(t, "1.0~~~rc1", "1.0~~~~rc1", 1)

	RPMVERCMP(t, "1.0~~", "1.0~~", 0)
	RPMVERCMP(t, "1.0~~~", "1.0~~~", 0)
	RPMVERCMP(t, "1.0~~~~", "1.0~~~~", 0)
	RPMVERCMP(t, "1.0~~~", "1.0~~~~", 1)
	RPMVERCMP(t, "1.0~~~x", "1.0~~~~x", 1)
}

func TestRpmvercmpCaret(t *testing.T) {
	if !tilde {
		t.SkipNow()
	}
	RPMVERCMP(t, "1.0^^rc1", "1.0^^rc1", 0)
	RPMVERCMP(t, "1.0^^^rc1", "1.0^^^rc1", 0)
	RPMVERCMP(t, "1.0^^^^rc1", "1.0^^^^rc1", 0)
	RPMVERCMP(t, "1.0^^^rc1", "1.0^^^^rc1", 1)

	RPMVERCMP(t, "1.0^^", "1.0^^", 0)
	RPMVERCMP(t, "1.0^^^", "1.0^^^", 0)
	RPMVERCMP(t, "1.0^^^^", "1.0^^^^", 0)
	RPMVERCMP(t, "1.0^^^", "1.0^^^^", -1)
	RPMVERCMP(t, "1.0^^^x", "1.0^^^^x", 1)
}

func TestASCII(t *testing.T) {
	data := []struct {
		ascii byte
		val   int
	}{
		{'0', 0x30},
		{'9', 0x39},
		{'A', 0x41},
		{'Z', 0x5a},
		{'a', 0x61},
		{'z', 0x7a},
		{'~', 0x7e},
	}
	for i := range data {
		ascii := data[i].ascii
		val := data[i].val
		ia := int(ascii)
		if ia != val {
			t.Errorf("%v = %d/%#x != %d/%#x", ascii, ia, ia, val, val)
		}
	}
}
