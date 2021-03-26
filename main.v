

/*
struct ID {
	site int
	counter int
}
*/

fn new_id(site u16, counter u16) u32 {
	return u32(site) << 16 | counter
}

struct Op {
mut:
	id u32
	origin u32
	left u32
	right u32
	deleted bool
	data int
}

const (
	left_delim = u32(100)
	right_delim = u32(101)
)

struct Store {
mut:
	data map[u32]Op
}

fn new_store() Store {
	mut s := Store{}
	s.data[left_delim] = Op{id: left_delim, left: left_delim, right: right_delim}
	s.data[right_delim] = Op{id: right_delim, left: left_delim, right: right_delim}

	return s
}

fn (mut s Store) insert(pos int, id u32, data int) {
	mut left := left_delim
	mut right := s.data[left].right

	for idx := 0;; idx++ {
		if idx == pos {
			break
		}
	}
	s.data[id] = Op{id: id, origin: left, left: left, right: right, data: data}
	s.data[left].right = id
	s.data[right].left = id
}

interface Identifiable {
	get_id() u64
}

struct ListItem {
	right u64
	left u64
	data any
}

struct List<T> {
	data map[u64]T
}

fn new_list<T>() List<T> {
	return List<T>{}
}

fn (mut l List<T>) insert(at u64, item T) {}
fn (mut l List<T>) insert_last() {}
fn (mut l List<T>) insert_head(item T) {}
fn (mut l List<T>) at(pos int) T {
	return l.data[0]
}

struct TempItem {
	id u64
}

fn (i TempItem) get_id() u64 {
	return i.id
}

fn main() {
	mut l := new_list<ListItem>()
	l.insert_head(TempItem{id: 100})
	assert l.at(0).get_id() == 100
}