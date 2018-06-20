package main

import (
	"flag"
	"fmt"
	"sort"
	"sync"

	"github.com/james-antill/repos"
)

const defScheme = "https"
const defHost = "mirrors.fedoraproject.org/metalink"
const defArch = "x86_64"
const defRepo = "fedora-28"

type repoData struct {
	name       string
	url        string
	mirrorlist bool
	plain      bool
}
type res struct {
	name string
	pkgs *repos.Pkgs
	err  error
}

func main() {
	flag.Parse()

	d := []repoData{}
	// Fedora repos...
	for _, repo := range []string{"26", "27", "28"} {
		url := fmt.Sprintf("%s://%s?repo=fedora-%s&arch=%s", defScheme, defHost,
			repo, defArch)
		d = append(d, repoData{name: "Fedora " + repo, url: url})
	}
	// Fedora Updates repos...
	for _, repo := range []string{"26", "27", "28"} {
		url := fmt.Sprintf("%s://%s?repo=updates-released-f%s&arch=%s", defScheme, defHost,
			repo, defArch)
		d = append(d, repoData{name: "Fedora Updates " + repo, url: url})
	}
	for _, repo := range []string{"26", "27", "28"} {
		url := fmt.Sprintf("%s://%s?repo=updates-testing-f%s&arch=%s", defScheme, defHost,
			repo, defArch)
		d = append(d, repoData{name: "Fedora Updates Tst " + repo, url: url})
	}
	for _, repo := range []string{"28"} {
		url := fmt.Sprintf("%s://%s?repo=updates-testing-modular-f%s&arch=%s", defScheme, defHost,
			repo, defArch)
		d = append(d, repoData{name: "Fedora Modular " + repo, url: url})
	}
	// Rawhide repo...
	if true {
		repo := "rawhide"
		url := fmt.Sprintf("%s://%s?repo=%s&arch=%s", defScheme, defHost,
			repo, defArch)
		d = append(d, repoData{name: "Fedora " + repo, url: url})
	}
	// EPEL repos...
	for _, repo := range []string{"6", "7"} {
		url := fmt.Sprintf("%s://%s?repo=epel-%s&arch=%s", defScheme, defHost,
			repo, defArch)
		d = append(d, repoData{name: "EPEL " + repo, url: url})
	}
	// CentOS repos...
	for _, repo := range []string{"6", "7"} {
		// url := fmt.Sprintf("%s://mirrorlist.centos.org/?release=%s&arch=%s&repo=os&infra=%s",
		url := fmt.Sprintf("%s://mirror.centos.org/centos/%s/%s/%s/",
			"http", repo, "os", defArch)
		d = append(d, repoData{name: "CentOS " + repo, url: url, plain: true})
	}
	// CentOS updates repos...
	for _, repo := range []string{"6", "7"} {
		url := fmt.Sprintf("%s://mirror.centos.org/centos/%s/%s/%s/",
			"http", repo, "updates", defArch)
		d = append(d, repoData{name: "CentOS Updates " + repo, url: url, plain: true})
	}
	// CentOS CR repo...
	for _, repo := range []string{"6", "7"} {
		url := fmt.Sprintf("%s://mirror.centos.org/centos/%s/%s/%s/",
			"http", repo, "cr", defArch)
		d = append(d, repoData{name: "CentOS CR " + repo, url: url, plain: true})
	}

	r := make(chan res, 4)
	var wg sync.WaitGroup
	for i := range d {
		rd := d[i]
		wg.Add(1)
		go func() {
			var snap *repos.Snapshot
			var err error

			if rd.plain {
				snap, err = repos.Baseurl(rd.url)
			} else { // Metalink...
				snap, err = repos.Metalink(rd.url)
			}
			if err != nil {
				r <- res{name: rd.name, pkgs: nil, err: err}
				wg.Done()
				return
			}

			repomd, err := snap.RepoMD()
			if err != nil {
				r <- res{name: rd.name, pkgs: nil, err: err}
				wg.Done()
				return
			}

			pkgs, err := repomd.Load()
			if err != nil {
				r <- res{name: rd.name, pkgs: nil, err: err}
				wg.Done()
				return
			}

			r <- res{name: rd.name, pkgs: pkgs}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(r)
	}()

	pkgs := []res{}
	for rv := range r {
		if rv.err != nil {
			fmt.Printf("Error: %s %v\n", rv.name, rv.err)
			continue
		}

		pkgs = append(pkgs, rv)
	}
	if true {
		sort.Slice(pkgs, func(i, j int) bool {
			return pkgs[i].name < pkgs[j].name
		})
	}

	cmd := "list"
	args := flag.Args()
	if len(args) > 0 {
		cmd = args[0]
		args = args[1:]
	}

	if len(args) > 0 {
		for i := range pkgs {
			rv := &pkgs[i]
			mpkgs := &repos.Pkgs{Repo: rv.pkgs.Repo}
			for _, arg := range args {
				mpkgs = mpkgs.Merge(rv.pkgs.Match(arg))
			}
			rv.pkgs = mpkgs
		}
	}

	switch cmd {
	case "list":
		for i := range pkgs {
			p := &pkgs[i]
			fmt.Println(p.name)
			for _, pkg := range p.pkgs.Pkgs {
				fmt.Println("", pkg)
			}
		}

	case "info":
		for i := range pkgs {
			p := &pkgs[i]
			fmt.Println(p.name)
			for _, pkg := range p.pkgs.Pkgs {
				fmt.Println("", pkg)
				fmt.Println("  ", pkg.Checksum())
			}
		}

	case "rpmdbversion":
		for i := range pkgs {
			p := &pkgs[i]
			fmt.Println(p.name)
			fmt.Println(p.pkgs.RPMDBVersion())
		}
	}
}
