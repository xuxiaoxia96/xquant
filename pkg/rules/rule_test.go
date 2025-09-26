package rules

import (
	"fmt"
	"testing"

	"github.com/bits-and-blooms/bitset"
)

func TestRule(t *testing.T) {
	fmt.Printf("Hello from BitSet!\n")
	var b bitset.BitSet
	//// play some Go Fish
	//for i := 0; i < 100; i++ {
	//	card1 := uint(rand.Intn(52))
	//	card2 := uint(rand.Intn(52))
	//	b.Set(card1)
	//	if b.Test(card2) {
	//		fmt.Println("Go Fish!")
	//	}
	//	b.Clear(card1)
	//}

	// Chaining
	b.Set(10).Set(11).Set(1000)

	for i, e := b.NextSet(0); e; i, e = b.NextSet(i + 1) {
		fmt.Println("The following bit is set:", i)
	}
	if b.Intersection(bitset.New(100).Set(10)).Count() == 1 {
		fmt.Println("Intersection works.")
	} else {
		fmt.Println("Intersection doesn't work???")
	}
	text := b.String()
	fmt.Println(text)
}
