package main

import "time"

func insertDB() {
	tick := time.Tick(1 * time.Second)
	go func() {
		for {
			select {
			case <-tick:
				timeInsert()
				// case <-ticks:
				algorithmInsert()
			}
		}
	}()
}
