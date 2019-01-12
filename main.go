package main

import (
	"log"
	"oynix.io/precision/prec"
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

}
