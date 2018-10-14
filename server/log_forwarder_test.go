package server

// TODO: convert into test
// time.Sleep(1 * time.Second)

// conn, err := pgxpool.Acquire()
// defer pgxpool.Release(conn)
// if err != nil {
// 	logger.Fatalln(err)
// }

// findOrCreateProjectID("manveru/scylla")
// buildID, _ := insertBuild(conn, 1, &githubJob{Hook: &GithubHook{}})

// for n := 0; n < 5; n++ {
// 	go func(n int) {
// 		c := make(chan *logLine)
// 		ll := &logListener{buildID: int64(buildID), recv: c}
// 		logListenerRegister <- ll
// 		m := 0
// 		for line := range c {
// 			m++
// 			pp.Println(n, m, line.Line)
// 			if m > 100 {
// 				logListenerUnregister <- ll
// 			}
// 		}
// 	}(n)
// }

// time.Sleep(1 * time.Second)

// for n := 0; n < 15; n++ {
// 	_, err := runCmdForBuild(int64(buildID), exec.Command("./sleepy.sh"))
// 	pp.Println(err)
// 	// forwardLogToDB(conn, 1, fmt.Sprintf("example line %d", n))
// }

// os.Exit(0)
