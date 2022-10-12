// https://play.golang.org/p/f46idWAVIcC
// https://play.golang.org/p/JZevbAOD1uL
// https://play.golang.org/p/XJo1iDmeEE0
// https://play.golang.org/p/LT7Zk1QfSXS
// https://play.golang.org/p/jfFrt4b2F8H
// https://play.golang.org/p/0TYNo_W-QYS
// https://play.golang.org/p/w-ZyPvKlOoI
// https://play.golang.org/p/eJj92yDwdYf
// https://play.golang.org/p/VyjAFFRVL3s
// https://play.golang.org/p/m6otphcj5Y2
// https://play.golang.org/p/3zRoISzoN1c
// https://play.golang.org/p/vt4aSwqPhvQ


package main

import (
	"fmt"
)

func main() {

	var a []int

	fmt.Println("a = ", a)

	b := []int{}

	fmt.Println("b = ", b)

	c := new([]int)

	fmt.Println("c = ", c)
	
	var d []int = []int{}

	fmt.Println("d = ", d)

}

// ----------
package main

import (
	"fmt"
)

func main() {
	a := []int{0,1,2,3}
	b := a
	
	b[1] = b[len(b)-1]
	b[len(b)-1] = 0
	b = b[:len(b)-1]
	
	fmt.Println(len(a), len(b), cap(a), cap(b))
}

// -----------
package main

import "time"

func main() {
	defer println(1)
	defer println(2)

	time.Sleep(1 * time.Second)

	go func() {
		panic("golang is dead")
	}()
}

// -----------
package main

import "time"

func main() {
	defer println(1)
	defer println(2)

	go func() {
		panic("golang is dead")
	}()
	
	time.Sleep(1 * time.Second)
}

// ---------------
package main

import (
	"fmt"
)

type MyError string

func (me MyError) Error() string {
	return string(me)
}


func main() {

	myError := func() error {
		var err *MyError
		
		fmt.Println("1. ",err == nil)
		
		return err
	}
		
	err := myError()
	
	fmt.Println("2. ",err == nil)
}

// --------------
package main

import (
	"fmt"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(1)

	done := false

	go func() {
		done = true
	}()

	for !done {
	}
	fmt.Println("finished")
}

// ----
package mainpackage main

func main() {
	var counter int

	for i := 0; i < 1000; i++ {
		go func() {
			counter++
		}()
	}
	
	println(counter)
}

// ----
package main

func main() {
	v := 5
	p := &v
	
	println(*p)

	changePointer(p)
	
	println(*p)
}

func changePointer(p *int) {
	v := 3
	p = &v
}

// ----
package main

import "time"

func worker(a int) chan int {
	ch := make(chan int)

	go func() {
		time.Sleep(1 * time.Second)
		ch <- a
	}()

	return ch
}

func main() {
	timeStart := time.Now()

	a := <-worker(1)
	b := <-worker(2)

	println(a, b)

	println("time =", int(time.Since(timeStart).Seconds()))
}

// ----
package main

import "unsafe"

type A struct {
	a bool
	b int32
	c float64
}

type B struct {
	c float64
	b int32
	a bool
}

type C struct {
    	a bool
	c float64
	b int32
}

func main() {
	a := A{}
	b := B{}
	c := C{}
	
	println(unsafe.Sizeof(a))
	println(unsafe.Sizeof(b))
	println(unsafe.Sizeof(c))
}

// -----
package main

import "fmt"

func main() {
	a := []int{1, 4, 5, 9, 2}
	b := []int{1, 2, 4, 5, 9}

	// sort a

	fmt.Println(a)
	fmt.Println(b)
}

// ------
package main

import "fmt"

func main() {
	a := []int{1, 2, 5, 90, 42}
	b := []int{1, 2, 40, 91, 900}
	c := []int{}
	r := []int{1, 2, 4, 5, 40, 42, 90, 91, 900}

	// union a and b to c
	

	fmt.Println(a)
	fmt.Println(b)
	fmt.Println(c)
	fmt.Println(r)
}
