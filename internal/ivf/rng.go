package ivf

type LCG struct {
	state uint64
}

func NewLCG(seed uint64) *LCG {
	return &LCG{state: seed}
}

func (l *LCG) Next() uint64 {
	l.state = l.state*6364136223846793005 + 1442695040888963407
	return l.state
}

func (l *LCG) Float64() float64 {
	return float64(l.Next()>>11) / float64(1<<53)
}

func (l *LCG) Intn(n int) int {
	return int(l.Next() % uint64(n))
}

func (l *LCG) Shuffle(n int, swap func(i, j int)) {
	for i := n - 1; i > 0; i-- {
		j := l.Intn(i + 1)
		swap(i, j)
	}
}
