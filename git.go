package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport/client"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
)

func (pkgs PackagesStr) install() {
	for i := 0; i < len(pkgs); i++ {
		for j := 0; j < len(pkgs[i].Packages); j++ {
			pkg := pkgs[i].Packages[j]
			pkg.cloneOrUpdate()
			pkg.install()
		}
	}
}

func (pkg *Package) cloneOrUpdate() {
	client.InstallProtocol("https", githttp.NewClient(CLIENT))

	pkgDir, _ := filepath.Abs(filepath.Join(RIME_DIR, "package", pkg.User, strings.TrimPrefix(pkg.Repo, "rime-")))
	pkg.WorkingDirectory = pkgDir
	localBranch := plumbing.NewBranchReferenceName(pkg.Branch)

	if _, err := os.Stat(pkgDir); os.IsNotExist(err) {
		fmt.Printf("Cloning %s to %s\n", pkg.Repo, pkgDir)
		_, err := git.PlainClone(pkgDir, false, &git.CloneOptions{
			URL:               pkg.URL,
			ReferenceName:     localBranch,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			Depth:             1})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Updating %s\n", pkgDir)
		repo, err := git.PlainOpen(pkgDir)
		if err != nil {
			fmt.Printf("%s is not a valid git repository\n", pkgDir)
			os.Exit(1)
		}

		// validate r.Branch
		remotes, err := repo.Remotes()
		if err != nil {
			fmt.Printf("failed to get remotes %s\n", err)
			os.Exit(1)
		}

		remoteName := remotes[0].Config().Name

		_, err = repo.Branch(pkg.Branch)
		if err != nil {
			refs, _ := remotes[0].List(&git.ListOptions{})
			var hasBranch bool
			for _, ref := range refs {
				if ref.Name() == localBranch {
					hasBranch = true
					break
				}
			}

			if !hasBranch {
				fmt.Printf("branch %s can not be found on local and remote\n", pkg.Branch)
				os.Exit(1)
			}

			err = repo.CreateBranch(&config.Branch{Name: pkg.Branch, Remote: remoteName, Merge: localBranch})
			if err != nil {
				fmt.Printf("failed to create branch %s\n", pkg.Branch)
				os.Exit(1)
			}
			remoteBranch := plumbing.NewRemoteReferenceName(remoteName, pkg.Branch)
			newReference := plumbing.NewSymbolicReference(localBranch, remoteBranch)
			err = repo.Storer.SetReference(newReference)
			if err != nil {
				fmt.Printf("failed to set reference %s->%s\n", localBranch, remoteBranch)
				os.Exit(1)
			}
		}

		w, err := repo.Worktree()
		if err != nil {
			fmt.Printf("%s doesn't contains any worktree\n", pkgDir)
			os.Exit(1)
		}

		err = w.Checkout(&git.CheckoutOptions{
			Branch: localBranch,
			Force:  true,
		})
		if err != nil {
			fmt.Printf("failed to checkout branch %s\n", pkg.Branch)
		}
		err = w.Pull(&git.PullOptions{
			RemoteName:    remoteName,
			ReferenceName: localBranch,
		})
		if err != nil {
			switch err.Error() {
			case "already up-to-date":
				fmt.Printf("%s already up-to-date\n", pkg.URL)
			case "empty git-upload-pack given":
				fmt.Printf("empty git-upload-pack given, %s already up-to-date\n", pkg.URL)
			default:
				fmt.Printf("failed to pull repository %s: %s\n", pkg.URL, err.Error())
				os.Exit(1)
			}
		}
	}
}

func commit(dir, file string) {
	repo, err := git.PlainOpen(dir)
	if err != nil {
		fmt.Printf("%s is not a valid git repository\n", dir)
		os.Exit(1)
	}
	w, err := repo.Worktree()
	if err != nil {
		fmt.Printf("%s doesn't contains any worktree\n", dir)
		os.Exit(1)
	}
	_, err = w.Add(filepath.Base(file))
	if err != nil {
		fmt.Printf("can not git add %s\n", file)
		os.Exit(1)
	}
	_, err = w.Commit(fmt.Sprintf("keep downloaded %s.%s", file, strconv.Itoa(int(time.Now().Unix()))), &git.CommitOptions{})
	if err != nil {
		fmt.Printf("failed to commit: %s", err)
		os.Exit(1)
	}
}
