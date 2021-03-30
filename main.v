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
mut:
	data map[u64]&Op
}

fn new_opset() OpSet {
	// delimiter
	mut left := &Op{id: new_id(0, 1), origin: 0, left: 0, right: 0, deleted: false, content: 0}
	mut right := &Op{id: new_id(0, 2), origin: 0, left: 0, right: 0, deleted: false, content: 0}
	left.right = right.id
	right.left = left.id

	mut data := map[u64]&Op{}
	data[left.id] = left
	data[right.id] = right
	return OpSet{
		left: left.id,
		right: right.id,
		data: data
	}
}

fn (s &OpSet) next() ?&Op {
	// loop left delimiter to right delimiter
	mut n := s.data[s.data[s.left].right]
	for ; n != 0 && n.right != 0; n = s.data[n.right] {
		return n
	}
	return error("no more element")
}

fn (s &OpSet) iter() OpSetIter {
	start := s.data[s.left].right
	return OpSetIter{s: s, start: start, end: s.right, cur: s.data[start]}
}

fn (s &OpSet) iter_range(start u64, end u64) OpSetIter {
	return OpSetIter{s: s, start: start, end: end, cur: s.data[start]}
}

// iterate all Op includes delimiter
fn (s &OpSet) iter_all() OpSetIter {
	return OpSetIter{s: s, start: s.left, end: -1, cur: s.data[s.left]}
}

struct OpSetIter {
	s &OpSet
	start u64
	end u64
mut:
	cur &Op = 0
}

fn (mut iter OpSetIter) next() ?&Op {
	// println("iter: ${iter}")
	for ; iter.cur != 0 && iter.cur.id != iter.end; {
		ret := iter.cur
		iter.cur = iter.s.data[iter.cur.right]
		return ret
	}
	return error("no more element")
}

fn (mut iter OpSetIter) as_arr() []&Op {
	mut result := []&Op{}
	for op in iter {
		result << op
	}
	return result
}

// to fix infinit recursion at &OpSet.str()
fn (iter OpSetIter) str() string {
	return "OpSetIter{cur: ${iter.cur.id}, start: ${iter.start}, end:${iter.end}}"
}

fn (s OpSet) str() string {
	iter := s.iter()
	println("str(): ${iter}")
	mut str := ""
	for op in iter {
		str = str + op.content.str()
	}
	return str
}

fn (s &OpSet) dump() {
	// in 0.2.2 can't write `idx, op in s.iter_all()`
	iter := s.iter_all()
	//println("c")
	//println(iter)
	//println("d")
	println("=== dump ===")
	for op in iter {
		println(op)
	}
	println("===")
}

fn (s OpSet) at(pos int) ?&Op {
	mut cnt := 0
	for op in s {
		if cnt == pos {
			return op
		}
		cnt++
	}
	return error('not found')
}

fn (s OpSet) pos(id u64) int {
	iter := s.iter_all()
	mut cnt := 0
	for op in iter {
		if op.id == id {
			return cnt
		}
		cnt++
	}
	return cnt
}

fn (s OpSet) gen_insert_op(pos int, content int, id u64) ?&Op {
	// in 0.2.2 can't write `idx, op in s.iter_all()`
	mut idx := 0
	iter := s.iter_all()
	for op in iter {
		if idx == pos {
			// return heap allocated Op (&)
			return &Op{
				content: content,
				id: id,
				origin: op.id,
				left: op.id
				right: op.right
			}
		}
		idx++
	}
	return error('invalid pos')
}

fn (mut s OpSet) insert(pos int, content int, id u64) {
	op := s.gen_insert_op(pos, content, id) or { return }
	s.integrate(op)
}

fn (mut s OpSet) integrate(op &Op) {
	println("integrate")
	mut conflicts := []Op{}
	iter := s.iter_range(s.data[op.left].right, op.right)
	for elem in iter {
		conflicts << elem
	}
	println("conflicts: ${conflicts}")

	// serach insert position
	mut left := op.left
	mut right := op.right
	if conflicts.len > 0 {
		left = conflicts[0].left
		right = conflicts[0].id
		println("left: ${left}, right: ${right}")
	}
	i_origin_pos := s.pos(op.origin)
	for o in conflicts {
		o_pos := s.pos(o.id)
		o_origin_pos := s.pos(o.origin)

		// check origin reference does NOT crossing (rule1)
		println("pos: ${i_origin_pos}, ${o_pos}, ${o_origin_pos}")
		if (o_pos < i_origin_pos) || (i_origin_pos <= o_origin_pos) {
			// if difference origin, insert to right-most place
			// otherwise (same origin), smaller unique-id is left (rule3)
			if (o_origin_pos != i_origin_pos) || (op.id > o.id) {
				println("found: ${o}, ${op}")
				left = o.id
				right = o.right
			}
		} else {
			break
		}
	}
	println("integrate ${op} to left:${left} right:${right}")

	s.data[op.id] = op
	s.data[op.id].left = left
	s.data[op.id].right = right
	s.data[left].right = op.id
	s.data[right].left = op.id
	println(s.data[left])
	println(s.data[right])
	println(s.data[op.id])
}

fn test_iter() {
	s := new_opset()
	{
		mut iter := s.iter()
		assert iter.as_arr() == []
	}

	{
		mut iter := s.iter_all()
		left_delim := &Op{id: 1, origin: 0, left: 0, right: 2, deleted: false, content: 0}
		right_delim := &Op{id: 2, origin: 0, left: 1, right: 0, deleted: false, content: 0}
		assert iter.as_arr().len == 2
	}
}

fn main() {
	test_iter()
	mut id := new_id(1, u32(time.now().unix_time()))

	mut s := new_opset()
	s.dump()
	r1 := s.gen_insert_op(0, 1, id++) or { panic(err) }
	r2 := s.gen_insert_op(0, 2, id++) or { panic(err) }
	r3 := s.gen_insert_op(0, 3, id++) or { panic(err) }
	//s.insert(0, 2, id++)
	//s.insert(0, 3, id++)
	s.dump()
	s.integrate(r1)
	s.integrate(r3)
	s.integrate(r2)
	s.dump()
	println("insert: ${s}")

/*


	println("empty: ${s}")
	s.insert(0, 1)
	println("add: ${s}")
	s.insert(0, 2)
	println("add: ${s}")
	s.insert(1, 3)
	println("add: ${s}")
	at0 := s.at(0) or { panic("position 0 not found") }
	println("at 0: ${at0}")
*/
}