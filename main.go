package main

import (
	"log"
	"oynix.io/precision/prec"
	"unsafe"
)

func main() {
	f1 := 423.1413
	f2 := 49384.28429
	add := prec.Add(f1, f2)
	log.Println(f1, "+", f2, "=", add, "系统", f1+f2)

	f3 := 423.1413000001
	f4 := 49384.28429
	sub := prec.Sub(f3, f4)
	log.Println(f3, "-", f4, "=", sub, "系统", f3-f4)

	ft := float64(49807.242)
	log.Printf("%f", ft*1000)

	printF(0.99999)
}

func printF(f float64) {
	i := (*uint64)(unsafe.Pointer(&f))
	log.Printf("%v, %b", f+0.000001, i)
}