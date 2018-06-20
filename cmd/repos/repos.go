package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/james-antill/repos"
)

const defScheme = "https"
const defHost = "mirrors.fedoraproject.org/metalink"
const defArch = "x86_64"
const defRepo = "fedora-28"

func main() {
	var repo string
	flag.StringVar(&repo, "repo", defRepo, "Set repo")
	flag.Parse()

	url := fmt.Sprintf("%s://%s?repo=%s&arch=%s", defScheme, defHost, repo, defArch)
	fmt.Println("URL:", url)
	snap, err := repos.Metalink(url)
	if err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
	}

	repomd, err := snap.RepoMD()
	if err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
	}
	// fmt.Println(repomd)

	pkgs, err := repomd.Load()

	cmd := "list"
	args := flag.Args()
	if len(args) > 0 {
		cmd = args[0]
		args = args[1:]
	}

	if len(args) > 0 {
		mpkgs := &repos.Pkgs{Repo: pkgs.Repo}
		for _, arg := range args {
			mpkgs = mpkgs.Merge(pkgs.Match(arg))
		}
		pkgs = mpkgs
	}

	switch cmd {
	case "list":
		for _, pkg := range pkgs.Pkgs {
			fmt.Println(pkg)
		}
	case "info":
		for _, pkg := range pkgs.Pkgs {
			fmt.Println(pkg)
			fmt.Println("", pkg.Checksum())
		}

	case "rpmdbversion":
		fmt.Println(pkgs.RPMDBVersion())
	}
}
