import time


fn new_id(site u32, cnt u32) u64 {
	return u64(site) << 32 | u64(cnt)
}

// ok(idk, origink, leftk, rightk, isDeletedk, contentk)
struct Op {
	id u64
	origin u64
mut:
	left u64
	right u64
	deleted bool
	content int
}

// We represent linear data as a doubly linked list S of insertions
struct OpSet {
	left u64
	right u64
	data map[u64]&Op
}

fn new_opset() OpSet {
	// delimiter
	mut left := &Op{id: new_id(0, 1), origin: 0, left: 0, right: 0, deleted: false, content: 0}
	mut right := &Op{id: new_id(0, 2), origin: 0, left: 0, right: 0, deleted: false, content: 0}
	left.right = right.id
	right.left = left.id
	// println(left)
	// println(right)

	mut data := map[u64]&Op{}
	data[left.id] = left
	data[right.id] = right
	return OpSet{
		left: left.id,
		right: right.id,
		data: data
	}
}

interface OpIterator {
	next(&Op) bool
}
interface OpMutIterator {
	next(mut op &Op) bool
}
fn (s OpSet) each(iter OpIterator) {
	// loop left delimiter to right delimiter
	mut n := s.data[s.data[s.left].right]
	for ; n != 0 && n.right != 0; n = s.data[n.right] {
		if iter.next(n) {
			break
		}
	}
}
fn (s OpSet) each_all(iter OpMutIterator) {
	// loop left delimiter to right delimiter
	mut n := s.data[s.left]
	for ; n != 0 && n.right != 0; n = s.data[n.right] {
		if iter.next(mut n) {
			break
		}
	}
}

// FIXME: use functor due to vlang's lambda can't capture variables
struct OpStringer {
mut:
	s string
}
fn (mut s OpStringer) next(op &Op) bool {
	s.s = s.s + op.content.str()
	return false
}
fn (s OpSet) str() string {
	ostr := OpStringer{}
	s.each(ostr)
	return ostr.s
}

// FIXME: use functor due to vlang's lambda can't capture variables
struct OpAt {
mut:
	target int
	result &Op = 0
	pos int
}
fn (mut a OpAt) next(op &Op) bool {
	if a.pos == a.target {
		a.result = op
		return true
	}
	a.pos++
	return false
}
fn (s OpSet) at(pos int) ?&Op {
	at := OpAt{target: pos}
	s.each(at)
	if at.result == 0 {
		return error('not found')
	}
	return at.result
}

// FIXME: use functor due to vlang's lambda can't capture variables
struct OpAdd {
mut:
	target int
	op &Op
	content int
	id u64
	pos int
}
fn (mut a OpAdd) next(mut op &Op) bool {
	if a.pos == a.target {
		a.op = &Op{
			...(*a.op)
			origin: op.id,
			left: op.id,
			right: op.right
		}
		return true
	}
	a.pos++
	return false
}
fn (s OpSet) gen_insert_op(pos int, content int) &Op {
	add := OpAdd{target: pos, op: &Op{content: content, id: new_id(1, u32(time.now().unix_time()))}}
	s.each_all(add)
	return add.op
}

fn (mut s OpSet) insert(pos int, content int) {
	op := s.gen_insert_op(pos, content)
	s.integrate(op)
}

fn (mut s OpSet) integrate(op &Op) {
}


struct ItTest {}

fn (t &ItTest) next() ?int {
	return 0
}

fn main() {

	iter := ItTest{}
	for i in iter {
		println(i)
	}



	mut s := new_opset()
	println("empty: ${s}")
	s.insert(0, 1)
	println("add: ${s}")
	s.insert(0, 2)
	println("add: ${s}")
	s.insert(1, 3)
	println("add: ${s}")
	at0 := s.at(0) or { panic("position 0 not found") }
	println("at 0: ${at0}")
}