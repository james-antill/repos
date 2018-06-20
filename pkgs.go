package repos

import (
	"bytes"
	"crypto/md5"
	"hash"

	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"

	"path/filepath"
)

type Pkg struct {
	name    string
	epoch   int
	version string
	release string
	arch    string
	chk     Checksum
}

func (pkg *Pkg) Nevra() string {
	return fmt.Sprintf("%s-%d:%s-%s.%s", pkg.name, pkg.epoch,
		pkg.version, pkg.release, pkg.arch)
}
func (pkg *Pkg) Nvra() string {
	return fmt.Sprintf("%s-%s-%s.%s", pkg.name,
		pkg.version, pkg.release, pkg.arch)
}
func (pkg *Pkg) Nevr() string {
	return fmt.Sprintf("%s-%d:%s-%s", pkg.name, pkg.epoch,
		pkg.version, pkg.release)
}
func (pkg *Pkg) Nvr() string {
	return fmt.Sprintf("%s-%s-%s", pkg.name,
		pkg.version, pkg.release)
}
func (pkg *Pkg) Na() string {
	return fmt.Sprintf("%s.%s", pkg.name, pkg.arch)
}
func (pkg *Pkg) Name() string {
	return pkg.name
}

func (pkg *Pkg) Envra() string {
	return fmt.Sprintf("%d:%s-%s-%s.%s", pkg.epoch, pkg.name,
		pkg.version, pkg.release, pkg.arch)
}

func (pkg *Pkg) UInevra() string {
	if pkg.epoch != 0 {
		return pkg.Nevra()
	}

	return pkg.Nvra()
}
func (pkg *Pkg) UInevr() string {
	if pkg.epoch != 0 {
		return pkg.Nevr()
	}

	return pkg.Nvr()
}
func (pkg *Pkg) UIenvra() string {
	if pkg.epoch != 0 {
		return pkg.Envra()
	}

	return pkg.Nvra()
}

// String: String representation of a pkg
func (pkg *Pkg) String() string {
	return pkg.UInevra()

}

func (pkg *Pkg) Checksum() Checksum {
	return pkg.chk
}

// Less: Comparison, for sorting
func (pkg *Pkg) Cmp(o *Pkg) int {
	if pkg == o {
		return 0
	}
	ret := strings.Compare(pkg.name, o.name)
	if ret != 0 {
		return ret
	}

	if pkg.epoch != o.epoch {
		return pkg.epoch - o.epoch
	}

	ret = rpmvercmp(pkg.version, o.version)
	if ret != 0 {
		return ret
	}

	ret = rpmvercmp(pkg.release, o.release)
	if ret != 0 {
		return ret
	}

	// This is outside the scope of RPM, as far as it cares this should
	// now be: return 0

	ret = strings.Compare(pkg.arch, o.arch)
	if ret != 0 {
		return ret
	}

	if false {
		ret = strings.Compare(pkg.chk.Kind, o.chk.Kind)
		if ret != 0 { // Skip ?
			return ret
		}
	}
	ret = strings.Compare(pkg.chk.Data, o.chk.Data)
	if ret != 0 {
		return ret
	}

	return 0
}

type ByPkg []*Pkg

func (a ByPkg) Len() int           { return len(a) }
func (a ByPkg) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPkg) Less(i, j int) bool { return a[i].Cmp(a[j]) < 0 }

// Match: User input to all forms of a package
func (pkg *Pkg) Match(pattern string) bool {
	f, e := filepath.Match(pattern, pkg.UInevra())
	if e != nil {
		return false
	}
	if f {
		return true
	}
	if f, _ = filepath.Match(pattern, pkg.UIenvra()); f {
		return true
	}
	if f, _ = filepath.Match(pattern, pkg.Nvra()); f {
		return true
	}
	if f, _ = filepath.Match(pattern, pkg.Nvr()); f {
		return true
	}
	if f, _ = filepath.Match(pattern, pkg.Na()); f {
		return true
	}
	if f, _ = filepath.Match(pattern, pkg.Name()); f {
		return true
	}
	return false
}

type Pkgs struct {
	Repo *Repodata
	Pkgs []*Pkg
}

func (repo *Repodata) Load() (*Pkgs, error) {
	var xmlData struct {
		Packages []struct {
			Name string `xml:"name"`
			V    struct {
				Epoch   int    `xml:"epoch,attr"`
				Version string `xml:"ver,attr"`
				Relase  string `xml:"rel,attr"`
			} `xml:"version"`
			Arch     string `xml:"arch"`
			Checksum struct {
				T string `xml:"type,attr"`
				D string `xml:",chardata"`
			} `xml:"checksum"`
		} `xml:"package"`
	}

	primarygz, err := url2bytes(repo.Baseurl + repo.Primary.Path)
	if err != nil {
		return nil, err
	}

	if !hchks(primarygz, repo.Primary.Chks) {
		err = fmt.Errorf("error: Checksum doesn't match for Primary")
		return nil, err
	}

	br := bytes.NewReader(primarygz)
	zr, err := autounzip(br, repo.Primary.Path)
	if err != nil {
		panic(err)
	}

	var primarybuf bytes.Buffer
	if _, err := io.Copy(&primarybuf, zr); err != nil {
		panic(err)
	}

	if err := zr.Close(); err != nil {
		panic(err)
	}

	primary := primarybuf.Bytes()
	// fmt.Println(string(primary))

	err = xml.Unmarshal(primary, &xmlData)
	if err != nil {
		fmt.Printf("error: %v", err)
		return nil, err
	}

	ret := &Pkgs{Repo: repo}
	for _, xp := range xmlData.Packages {
		p := &Pkg{}
		p.name = xp.Name
		p.arch = xp.Arch
		p.version = xp.V.Version
		p.release = xp.V.Relase
		p.epoch = xp.V.Epoch
		p.chk = Checksum{Kind: xp.Checksum.T, Data: xp.Checksum.D}
		ret.Pkgs = append(ret.Pkgs, p)
	}

	sort.Sort(ByPkg(ret.Pkgs))

	return ret, nil
}

func (snap *Repodata) MustLoad() *Pkgs {
	ret, err := snap.Load()
	if err != nil {
		panic(err)
	}

	return ret
}

func (pkgs *Pkgs) Match(pattern string) *Pkgs {
	if len(pattern) == 0 {
		return pkgs
	}

	ret := &Pkgs{Repo: pkgs.Repo}
	for _, p := range pkgs.Pkgs {
		if p.Match(pattern) {
			ret.Pkgs = append(ret.Pkgs, p)
		}
	}

	return ret
}

func (a *Pkgs) Merge(b *Pkgs) *Pkgs {
	arepo := a.Repo
	if a.Repo != b.Repo {
		arepo = nil
	}

	ret := &Pkgs{Repo: arepo}

	pas := a.Pkgs
	pbs := b.Pkgs
	for len(pas) > 0 && len(pbs) > 0 {
		c := pas[0].Cmp(pbs[0])
		switch {
		case c == 0:
			// Pick a, but remove b too
			pbs = pbs[1:]
			fallthrough
		case c < 0:
			ret.Pkgs = append(ret.Pkgs, pas[0])
			pas = pas[1:]
		case c > 0:
			ret.Pkgs = append(ret.Pkgs, pbs[0])
			pbs = pbs[1:]
		}
	}

	ret.Pkgs = append(ret.Pkgs, pas...)
	ret.Pkgs = append(ret.Pkgs, pbs...)

	return ret
}

type RPMDBV struct {
	count int
	chk   Checksum
	hash  hash.Hash
}

func (r *RPMDBV) String() string {
	return fmt.Sprintf("%d:%s", r.count, r.chk.Data)
}

func (r *RPMDBV) Count() int {
	return r.count
}
func (r *RPMDBV) Checksum() Checksum {
	return r.chk
}
func iRPMDBVmake() *RPMDBV {
	r := &RPMDBV{}
	r.hash = md5.New()

	return r
}

func (r *RPMDBV) add(pkg *Pkg) {
	r.count++

	_, _ = io.WriteString(r.hash, pkg.UIenvra())
	if pkg.chk.Kind != "" {
		_, _ = io.WriteString(r.hash, pkg.chk.Kind)
		_, _ = io.WriteString(r.hash, pkg.chk.Data)
	}
}

func (r *RPMDBV) done() {
	r.chk.Kind = "md5"
	r.chk.Data = fmt.Sprintf("%x", r.hash.Sum(nil))
	r.hash = nil
}

func (pkgs *Pkgs) RPMDBVersion() *RPMDBV {
	r := iRPMDBVmake()
	for _, p := range pkgs.Pkgs {
		r.add(p)
	}
	r.done()

	return r
}
