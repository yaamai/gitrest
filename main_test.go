package main

import (
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
)

func BenchmarkCommitSpeed(b *testing.B) {
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		f2, err := fs.OpenFile("aa", os.O_RDWR, 0600)
		if err != nil {
			log.Println(err)
		}
		_, err = f2.Write([]byte(strconv.Itoa(i)))
		if err != nil {
			log.Println(err)
		}
		f2.Close()
		b.StartTimer()
		// log.Println(l)

		w.AddWithOptions(&git.AddOptions{Path: "aa"})
		opts := &git.CommitOptions{
			Author: &object.Signature{
				Name:  "John Doe",
				Email: "john@doe.org",
				When:  time.Now(),
			},
		}
		_, err = w.Commit("test", opts)
		if err != nil {
			log.Fatalln(err)
		}
		// log.Println("new-commit:", c)
	}
}
