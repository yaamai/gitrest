


struct Item {
	id u64
}

struct List {
mut:
	data []Item
}

fn new_list() List {
	return List{}
}

fn (l List) at(pos int) Item {
	return l.data[pos]
}

fn concat(a []Item, b []Item, c []Item) []Item {
	mut r := []Item{len: a.len + b.len + c.len}
	mut idx := 0
	for i in a {
		r[idx] = i
		idx++
	}
	for i in b {
		r[idx] = i
		idx++
	}
	for i in c {
		r[idx] = i
		idx++
	}
	return r
}

fn (mut l List) insert(pos int, item Item) {
	l.data = concat(l.data[..pos], [item], l.data[pos..])
}

fn main() {
	mut l := new_list()

	l.insert(0, Item{id: 100})
	assert l.at(0).id == 100
	println("---")

	l.insert(0, Item{id: 101})
	assert l.at(0).id == 101
	assert l.at(1).id == 100
	println("---")

	l.insert(1, Item{id: 102})
	assert l.at(0).id == 101
	assert l.at(1).id == 102
	assert l.at(2).id == 100
}