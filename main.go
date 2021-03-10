package main

import (
	"fmt"
	"log"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
)

func main() {

	fs := memfs.New()
	r, err := git.Init(memory.NewStorage(), fs)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("repo:", r)

	w, err := r.Worktree()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("worktree", w)

	f, err := fs.Create("aa")
	if err != nil {
		log.Fatalln(err)
	}
	f.Write([]byte("test"))
	f.Close()

	w.AddWithOptions(&git.AddOptions{Path: "aa"})
	_, err = w.Commit("test", &git.CommitOptions{})
	if err != nil {
		log.Fatalln(err)
	}

	ref, err := r.Head()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(ref)

	commit, err := r.CommitObject(ref.Hash())
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("commit:", commit)

	obj, err := r.Storer.EncodedObject(plumbing.CommitObject, commit.Hash)
	if err != nil {
		log.Fatalln(err)
	}
	objr, err := obj.Reader()
	if err != nil {
		log.Fatalln(err)
	}

	buf := make([]byte, 1024)
	l, err := objr.Read(buf)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("commit-obj:", buf[:l], string(buf[:l]))
	log.Println("commit-encode:", obj)

	iter, err := commit.Files()
	if err != nil {
		log.Fatalln(err)
	}
	iter.ForEach(func(f *object.File) error {
		log.Println("file:", f)
		lines, err := f.Lines()
		log.Println(lines, err)
		return nil
	})
	// /refs/master/path/to/file
	// /commits/hash/path/to/file
	// /tags?branches?/master/path/to/file

	tree, err := commit.Tree()
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("tree:", tree)

	file, err := tree.File("aa")
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("file:", file)

	fr, err := file.Reader()
	fmt.Println(err)

	l, err = fr.Read(buf)
	fmt.Println(l, err)
	log.Println(string(buf[:l]))
}
