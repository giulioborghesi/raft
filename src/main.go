package main

import "github.com/giulioborghesi/raft-implementation/src/app"

func Main() {
	a := app.MakeRaftApp("/root/state_file", "/root/log_file", 0, nil)
	a.ListenAndServe("5050", "5051")
}
