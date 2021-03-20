package main

import (
	"io"
	"log"
	"net/http"
	"os"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/gorilla/mux"
)

func initServer(repo *git.Repository) {
	worktree, err := repo.Worktree()
	if err != nil {
		log.Fatalln(err)
	}

	r := mux.NewRouter()

	r.HandleFunc("/{path}", func(w http.ResponseWriter, r *http.Request) {
		f, err := worktree.Filesystem.Open(r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		io.Copy(w, f)
	}).Methods("GET")

	r.HandleFunc("/{path}", func(w http.ResponseWriter, r *http.Request) {
		p := mux.Vars(r)["path"]
		f, err := worktree.Filesystem.Create(p)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		io.Copy(f, r.Body)

		worktree.AddWithOptions(&git.AddOptions{Path: p})
		_, err = worktree.Commit("test", &git.CommitOptions{})
		if err != nil {
			log.Fatalln(err)
		}
	}).Methods("POST")

	r.HandleFunc("/commits/{path}", func(w http.ResponseWriter, r *http.Request) {
		p := mux.Vars(r)["path"]
		commitIter, _ := repo.Log(&git.LogOptions{FileName: &p})
		commitIter.ForEach(func(c *object.Commit) error {
			log.Println(c.Hash)
			return nil
		})
	})

	wd, _ := os.Getwd()
	log.Println(wd)
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("."))))

	http.ListenAndServe(":13223", r)
}

func main() {

	fs := memfs.New()
	repo, err := git.Init(memory.NewStorage(), fs)
	if err != nil {
		log.Fatalln(err)
	}

	initServer(repo)

	/*
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
	*/
}
