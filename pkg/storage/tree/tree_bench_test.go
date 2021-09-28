package tree

import (
	"bufio"
	"bytes"
	"math/big"
	"os"
	"strconv"
	"testing"
)

var (
	rawTreeA = mustParse("testdata/tree_1.txt")
	rawTreeB = mustParse("testdata/tree_2.txt") // tree.Len() == 1195
)

type line struct {
	key   []byte
	value uint64
}

func mustParse(path string) (lines []line) {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = f.Close()
	}()
	w := []byte{' '}
	s := bufio.NewScanner(bufio.NewReader(f))
	for s.Scan() {
		i := bytes.LastIndex(s.Bytes(), w)
		n, err := strconv.Atoi(s.Text()[i+1:])
		if err != nil {
			panic(err)
		}
		lines = append(lines, line{
			key:   []byte(s.Text())[:i],
			value: uint64(n),
		})
	}
	return lines
}

func createTree(b *testing.B, input []line) *Tree {
	b.StopTimer()
	tree := New()
	for _, l := range input {
		tree.Insert(l.key, l.value)
	}
	b.StartTimer()
	return tree
}

func BenchmarkInsert(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tree := New()
		for _, l := range rawTreeB {
			tree.Insert(l.key, l.value)
		}
	}
}

func BenchmarkClone(b *testing.B) {
	r := big.NewRat(1, 1)
	for i := 0; i < b.N; i++ {
		createTree(b, rawTreeB).Clone(r)
	}
}

func BenchmarkMerge(b *testing.B) {
	treeA := New()
	for _, l := range rawTreeA {
		treeA.Insert(l.key, l.value)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		createTree(b, rawTreeB).Merge(treeA)
	}
}

func BenchmarkTruncate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		createTree(b, rawTreeB).Truncate(512)
	}
}
