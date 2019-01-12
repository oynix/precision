package prec

import (
	"testing"
)

func TestAdd(t *testing.T) {
	f := 0.0
	b, i, i2 := float642Bytes(f)
	t.Log(b, i, i2)
}

func TestAdd2(t *testing.T) {
}

func TestAdd3(t *testing.T) {
	f1 := -21424.0173010901
	t.Log("f1", f1)
	f2 := -292.00000001
	t.Log("f2", f2)
	add := Add(f1, f2)
	t.Log("Add\t", add)
	t.Log("+\t", f1+f2)
}

func TestAdd4(t *testing.T) {
	s := "a22424234"
	t.Log("----", '2')
	l := len(s)
	for i := 0; i < l; i++ {
		t.Log(s[i], string(s[i]))
	}
}

func BenchmarkAdd(b *testing.B) {
	for i := 0; i < b.N; i ++ {
		f1 := 2424804242.242401313
		f2 := 0.0
		Add(f1, f2)
	}
}

func BenchmarkAdd3(b *testing.B) {
	for i := 0; i < b.N; i ++ {
		f := 2249423528484742.2424
		float642Uints(f)
	}
}

func BenchmarkAdd4(b *testing.B) {
	for i := 0; i < b.N; i ++ {
		bytes2Float64(false, []byte{4, 3, 14}, []byte{8, 9, 3, 12})
	}
}

func BenchmarkAdd5(b *testing.B) {
	for i := 0; i < b.N; i ++ {
		f := 224942.2424
		f2 := 2424.248294
		float642Uints(f)
		float642Uints(f2)
	}
}

func TestSub(t *testing.T) {
	//f1 := 24.244
	//f2 := 4.91118
	f1 := 423.1413000001
	f2 := 49384.28429
	sub := Sub(f1, f2)
	t.Log("Sub\t", sub)
	t.Log("-\t", f1-f2)
}

func BenchmarkSub(b *testing.B) {
	for i := 0; i < b.N; i ++ {
		f1 := 24.244
		f2 := 4.91118
		Sub(f1, f2)
	}
}
