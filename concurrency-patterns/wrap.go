package main

import (
	"fmt"
	"time"
)

// First incomming regex processing based on in-coded states (data states)
func isRegexFirst(readChar func() int) bool {
	state := 0
	for {
		c := readChar()
		if c == -1 {
			return false
		}
		switch state {
		case 0:
			if c != '"' {
				return false
			}
			state = 1
		case 1:
			if c == '"' {
				return true
			}
			if c == '\\' {
				state = 2
			} else {
				state = 1
			}
		case 2:
			state = 1
		}
	}
}

// Simplified version of incodded proccessed
// where data states converted into code states
func isRegexModified(readChar func() int) bool {
	if readChar() != '"' {
		return false
	}
	var c int
	for c != '"' && c != -1 {
		c = readChar()
		if c == '\\' {
			readChar()
		}
	}
	if c == -1 {
		return false
	}
	return true
}

////////////////////////////////////////////////////////////

type Status int

const (
	BadInput Status = iota
	Success
	NeedMoreInput
)

func (s Status) String() string {
	switch s {
	case BadInput:
		return "BadInput"
	case Success:
		return "Success"
	case NeedMoreInput:
		return "NeedMoreInput"
	default:
		return "UNEXPECTED_STATUS"
	}
}

type quoteReader struct {
	state int
}

func (q *quoteReader) Init() {
	q.state = 0
}

func (q *quoteReader) ProcessChar(c rune) Status {
	switch q.state {
	case 0:
		if c != '"' {
			return BadInput
		}
		q.state = 1
	case 1:
		if c == '"' {
			return Success
		}
		if c == -1 {
			return BadInput
		}
		if c == '\\' {
			q.state = 2
		} else {
			q.state = 1
		}
	case 2:
		q.state = 1
	}
	return NeedMoreInput
}

///////////////////
type quoteReaderWithGoroutine struct {
	char   chan rune
	status chan Status
	done   chan Status
}

func (q *quoteReaderWithGoroutine) Init(
	readString func(func() int) bool,
) {
	q.char = make(chan rune, 1)
	q.status = make(chan Status)

	go func() {
		r := readString(q.readChar)
		if r {
			q.status <- Success
		} else {
			q.status <- BadInput
		}
	}()
}

func (q *quoteReaderWithGoroutine) readChar() int {
	c := <-q.char

	if int(c) == -1 {
		q.status <- BadInput
	} else {
		q.status <- NeedMoreInput
	}

	return int(c)
}

func (q *quoteReaderWithGoroutine) ProcessChar(c rune) Status {
	select {
	case s := <-q.status:
		return s
	case q.char <- c:
	}

	return <-q.status
}

///////////////////
func wrap() {
	str := `"([^\"\\]|\\.)*\"\"`

	readCharFn := func(str string) func() int {
		i := -1
		return func() int {
			i++
			if i >= len(str) {
				return -1
			}
			fmt.Printf("%s ", string(str[i]))
			return int(str[i])
		}

	}

	fmt.Println(isRegexFirst(readCharFn(str)))
	fmt.Println(isRegexModified(readCharFn(str)))

	/////////////////////////////////////////////////////////////

	var (
		qr quoteReader
		s  Status
	)

	qr.Init()

	for i := range str {
		fmt.Printf("%s ", string(str[i]))
		s = qr.ProcessChar(rune(str[i]))
		if s == BadInput || s == Success {
			fmt.Println(s)
			break
		}
	}

	if s == NeedMoreInput {
		fmt.Println(qr.ProcessChar(-1))
	}

	/////////////////////////////////////////////////////////////

	var (
		qrwg quoteReaderWithGoroutine
	)

	qrwg.Init(isRegexModified)

	for i := range str {
		fmt.Printf("%s ", string(str[i]))
		s = qrwg.ProcessChar(rune(str[i]))
		if s == BadInput || s == Success {
			fmt.Println(s)
			break
		}
	}

	if s == NeedMoreInput {
		fmt.Println(qrwg.ProcessChar(-1))
	}

	time.Sleep(1000000)
}
