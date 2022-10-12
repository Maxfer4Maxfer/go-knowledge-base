package main

// Work scheduler

func Schedule1(
	servers []string, numTask int,
	call func(srv string, task int),
) {
	idle := make(chan string, len(servers))
	for _, srv := range servers {
		idle <- srv
	}

	for task := 0; task < numTask; task++ {
		go func(task int) {
			srv := <-idle
			call(srv, task)
			idle <- srv
		}(task)
	}
}

func Schedule2(
	servers []string, numTask int,
	call func(srv string, task int),
) {
	idle := make(chan string, len(servers))
	for _, srv := range servers {
		idle <- srv
	}

	for task := 0; task < numTask; task++ {
		task := task
		srv := <-idle
		go func() {
			call(srv, task)
			idle <- srv
		}()
	}

	for i := 0; i < len(servers); i++ {
		<-idle
	}
}

func Schedule3(
	servers chan string, numTask int,
	call func(srv string, task int) bool,
) {
	work := make(chan int, numTask)
	done := make(chan bool)
	exit := make(chan bool)

	runTasks := func(srv string) {
		for task := range work {
			if call(srv, task) {
				done <- true
			} else {
				work <- task
			}
		}
	}

	go func() {
		for {
			select {
			case srv := <-servers:
				go runTasks(srv)
			case <-exit:
				return
			}
		}
	}()

	for task := 0; task < numTask; task++ {
		work <- task
	}

	for i := 0; i < numTask; i++ {
		<-done
	}

	close(work)
	exit <- true
}

func pattern2() {
}
