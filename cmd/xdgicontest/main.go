package main

import (
	"fmt"
	"log"
	"strconv"

	// "log"
	"os"
	"time"

	"github.com/codelif/xdgicons"
)

func main() {
	size, _ := strconv.Atoi(os.Args[2])
	scale, _ := strconv.Atoi(os.Args[3])
	t := time.Now()
	il := xdgicons.NewIconLookup()
	icon, _ := il.FindIcon(os.Args[1], size, scale)
	fmt.Printf("First Lookup: %d\n", time.Now().Sub(t).Milliseconds())

	t = time.Now()
	icon, _ = il.FindIcon(os.Args[1], size, scale)
	fmt.Printf("Second Lookup: %d\n", time.Now().Sub(t).Milliseconds())

	time.Sleep(6 * time.Second)
	t = time.Now()
	icon, _ = il.FindIcon(os.Args[1], size, scale)
	fmt.Printf("Third Lookup: %d\n", time.Now().Sub(t).Milliseconds())

	t = time.Now()
	icon, err := il.FindIcon(os.Args[1], size, scale)
	fmt.Printf("Fourth Lookup: %d\n", time.Now().Sub(t).Milliseconds())

	if err != nil {
		log.Fatalf("%v", err)
	}
	fmt.Printf("%s", icon.Path)
}
