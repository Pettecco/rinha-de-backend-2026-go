package ivf

type TopKEntry struct {
	dist    uint32
	isFraud bool
}

type TopK struct {
	entries [5]TopKEntry
	size    int
}

func (tk *TopK) Push(dist uint32, isFraud bool) {
	if tk.size < 5 {
		tk.entries[tk.size] = TopKEntry{dist: dist, isFraud: isFraud}
		tk.size++
		if tk.size == 5 {
			tk.siftUp()
		}
		return
	}

	if dist >= tk.entries[0].dist {
		return
	}

	tk.entries[0] = TopKEntry{dist: dist, isFraud: isFraud}
	tk.siftDown()
}

func (tk *TopK) FraudCount() int {
	count := 0
	for i := 0; i < tk.size; i++ {
		if tk.entries[i].isFraud {
			count++
		}
	}
	return count
}

func (tk *TopK) Reset() {
	tk.size = 0
}

func (tk *TopK) siftUp() {
	for i := tk.size/2 - 1; i >= 0; i-- {
		tk.siftDownFrom(i)
	}
}

func (tk *TopK) siftDown() {
	tk.siftDownFrom(0)
}

func (tk *TopK) siftDownFrom(i int) {
	n := tk.size
	for {
		smallest := i
		left := 2*i + 1
		right := 2*i + 2

		if left < n && tk.entries[left].dist > tk.entries[smallest].dist {
			smallest = left
		}
		if right < n && tk.entries[right].dist > tk.entries[smallest].dist {
			smallest = right
		}
		if smallest == i {
			break
		}
		tk.entries[i], tk.entries[smallest] = tk.entries[smallest], tk.entries[i]
		i = smallest
	}
}
