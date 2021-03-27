import time



fn new_id(site u32, cnt u32) u64 {
	return u64(site) << 32 | u64(cnt)
}

// ok(idk, origink, leftk, rightk, isDeletedk, contentk)
struct Op {
	id u64
	origin u64
mut:
	left &Op
	right &Op
	deleted bool
	content int
}

// We represent linear data as a doubly linked list S of insertions
struct OpSet {
	left &Op
	right &Op
}

fn new_opset() OpSet {
	// delimiter
	mut left := &Op{id: new_id(0, 0), origin: 0, left: 0, right: 0, deleted: false, content: 0}
	mut right := &Op{id: new_id(0, 1), origin: 0, left: 0, right: 0, deleted: false, content: 0}
	left.right = right
	right.left = left
	// println(left)
	// println(right)

	return OpSet{
		left: left,
		right: right
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
	mut n := s.left.right
	for ; n != 0 && n.right != 0; n = n.right {
		if iter.next(n) {
			break
		}
	}
}
fn (s OpSet) each_all(iter OpMutIterator) {
	// loop left delimiter to right delimiter
	mut n := s.left
	for ; n != 0 && n.right != 0; n = n.right {
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
	content int
	id u64
	pos int
}
fn (mut a OpAdd) next(mut op &Op) bool {
	if a.pos == a.target {
		new_op := &Op{id: a.id, origin: op.id, left: op, right: op.right, deleted: false, content: a.content}
		op.right.left = new_op
		op.right = new_op
		return true
	}
	a.pos++
	return false
}
fn (mut s OpSet) add(pos int, content int) {
	add := OpAdd{target: pos, content: content, id: new_id(1, u32(time.now().unix_time()))}
	s.each_all(add)
}

fn main() {
	mut s := new_opset()
	println("empty: ${s}")
	s.add(0, 1)
	println("add: ${s}")
	s.add(0, 2)
	println("add: ${s}")
	s.add(1, 3)
	println("add: ${s}")
	at0 := s.at(0) or { panic("position 0 not found") }
	println("at 0: ${at0}")
}