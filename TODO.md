
## todo
- [ ] restructure paths
  - file/directory access (GET /path/to/file, GET /path/to/dir, GET /files/path/to/file)
    - check path is exists
    - check target is object or tree
    - return file content if object
    - return directory info if tree
    - return plumbing object hash for tree or file
  - create file or directory (POST/PUT /path/to/file, POST/PUT /path/to/dir, POST/PUT /files/path/to/file)
    - create directory if necessary
    - return err if file already exists within request path
    - create file
  - watch file modify (GET /notifies/path/to/file, GET /notifies/path/to/dir)
    - generate notify client id
    - create websocket conn
    - create notify terminate chan
    - hold path,id,ws-conn,terminate-chan
    - send notify client id via ws-conn
    - wait terminate-chan

- [ ] refactor

- [ ] commit author support
- [ ] sync to other repository server
- [ ] more low level access (plumbing)
- [ ] branch(refs) support
  - GET/POST/PUT/DELETE /refs/HEAD/files/path/to/file
  - GET/POST/PUT/DELETE /refs/master/files/path/to/file
  - GET/POST/PUT/DELETE /blobs/hash
- [ ] merge if conflict (maybe 41usec + notify RTT (1-2ms*2) = <4ms conflict...)

