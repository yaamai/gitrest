package main

import (
	"io"
	"log"
	"net/http"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/gorilla/mux"
)

type GitRestServer struct {
	repo     *git.Repository
	worktree *git.Worktree
}

func NewGitRestServer(opts *git.CloneOptions) (*GitRestServer, error) {
	fs := memfs.New()
	stor := memory.NewStorage()
	repo, err := git.Clone(stor, fs, opts)
	if err != nil {
		return nil, err
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}

	return &GitRestServer{
		repo:     repo,
		worktree: worktree,
	}, nil
}

func (s *GitRestServer) Sync() {
	opts := &git.PushOptions{
		RemoteName: "origin",
	}
	err := s.repo.Push(opts)
	if err != nil {
		log.Println(err)
	}
}

func (s *GitRestServer) handleReadFile(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)["path"]
	f, err := s.worktree.Filesystem.Open(p)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	io.Copy(w, f)
}

func (s *GitRestServer) handleUpdateFile(w http.ResponseWriter, r *http.Request) {
	p := mux.Vars(r)["path"]

	f, err := s.worktree.Filesystem.Create(p)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	io.Copy(f, r.Body)

	err = s.worktree.AddWithOptions(&git.AddOptions{Path: p})
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	_, err = s.worktree.Commit("test", &git.CommitOptions{})
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	go s.Sync()
}

func (s *GitRestServer) Router() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/{path}", s.handleReadFile).Methods("GET")
	r.HandleFunc("/{path}", s.handleUpdateFile).Methods("POST", "PUT")
	return r
}

func main() {
	opts := &git.CloneOptions{
		URL: "test",
	}

	s, err := NewGitRestServer(opts)
	if err != nil {
		log.Fatalln(err)
	}

	r := mux.NewRouter()
	r.PathPrefix("/api").Handler(http.StripPrefix("/api", s.Router()))
	r.PathPrefix("/static").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("."))))
	http.ListenAndServe(":13223", r)
}

/*
import (
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sync"

	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// RandStringBytesMaskImpr generates random string
func RandStringBytesMaskImpr(n int) string {
	b := make([]byte, n)
	// A rand.Int63() generates 63 random bits, enough for letterIdxMax letters!
	for i, cache, remain := n-1, rand.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rand.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func initServer(repo *git.Repository) {
	worktree, err := repo.Worktree()
	if err != nil {
		log.Fatalln(err)
	}

	notifyMap := map[string]map[string]*websocket.Conn{}
	upgrader := websocket.Upgrader{}
	notifyMutex := sync.Mutex{}
	mapMutex := sync.Mutex{}

	r := mux.NewRouter()

	r.HandleFunc("/{path}", func(w http.ResponseWriter, r *http.Request) {
		f, err := worktree.Filesystem.Open(r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		ref, err := repo.Head()
		w.Header().Add("X-Commit", ref.Hash().String())
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

		err = worktree.AddWithOptions(&git.AddOptions{Path: p})
		if err != nil {
			log.Fatalln(err)
		}
		_, err = worktree.Commit("test", &git.CommitOptions{})
		if err != nil {
			log.Fatalln(err)
		}
		ref, err := repo.Head()
		w.Header().Add("X-Commit", ref.Hash().String())

		notifyID := r.Header.Get("X-Notify-Id")

		go func() {
			if notifyTo, ok := notifyMap[p]; ok {
				notifyMutex.Lock()
				defer notifyMutex.Unlock()
				for id, ws := range notifyTo {
					if id == notifyID {
						continue
					}
					ws.WriteJSON(map[string]string{"type": "update", "id": ref.Hash().String()})
				}
			}
		}()
	}).Methods("POST")

	r.HandleFunc("/commits/{path}", func(w http.ResponseWriter, r *http.Request) {
		p := mux.Vars(r)["path"]
		commitIter, _ := repo.Log(&git.LogOptions{FileName: &p})
		commitIter.ForEach(func(c *object.Commit) error {
			log.Println(c.Hash)
			return nil
		})
	})

	r.HandleFunc("/notifies/{path}", func(w http.ResponseWriter, r *http.Request) {
		p := mux.Vars(r)["path"]
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatalln(err)
		}
		id := RandStringBytesMaskImpr(6)
		ws.WriteJSON(map[string]string{"type": "id", "id": id})
		mapMutex.Lock()
		if _, ok := notifyMap[p]; !ok {
			notifyMap[p] = map[string]*websocket.Conn{}
		}
		if _, ok := notifyMap[p][id]; !ok {
			notifyMap[p][id] = ws
		}
		mapMutex.Unlock()

		log.Println("notify registered")
		for {
		}
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

*/
