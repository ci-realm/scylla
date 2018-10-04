package queue

import (
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx"
	. "github.com/smartystreets/goconvey/convey"
)

type testLogger struct{}

func (l testLogger) Log(lvl pgx.LogLevel, msg string, data map[string]interface{}) {
	// fmt.Println(msg)
	// pp.Println(data)
}

func makeTestQueue(t *testing.T) Queue {
	cfg, err := pgx.ParseURI(os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	cfg.LogLevel = pgx.LogLevelTrace
	cfg.Logger = testLogger{}

	pool, err := pgx.NewConnPool(pgx.ConnPoolConfig{
		ConnConfig:     cfg,
		MaxConnections: 30,
	})
	if err != nil {
		t.Fatal(err)
	}

	tx, err := pool.Begin()
	if err != nil {
		t.Fatal(err)
	}
	err = prepareDatabase(tx)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.Commit()
	if err != nil {
		t.Fatal(err)
	}

	return Queue{
		Timeout: 100 * time.Millisecond,
		Pool:    pool,
		Name:    "test",
	}
}

func TestEnqueue(t *testing.T) {
	queue := makeTestQueue(t)
	firstJob := &Item{
		Args: map[string]interface{}{"Hello": "There"},
	}

	Convey("Successful job", t, func() {
		Convey("Insert a job in another connection", func() {
			So(queue.Insert(firstJob), ShouldBeNil)
		})

		Convey("Reserve one", func() {
			err := queue.Reserve(func(i *Item) error {
				So(i.QueueName, ShouldEqual, queue.Name)
				So(i.Args, ShouldResemble, firstJob.Args)
				return nil
			})
			So(err, ShouldBeNil)
		})

		Convey("Check that no job is left behind", func() {
			err := queue.Reserve(func(i *Item) error { return nil })
			So(err, ShouldBeError, "no rows in result set")
		})
	})

	Convey("Timeout job", t, func() {
		Convey("Insert a job in another connection", func() {
			So(queue.Insert(firstJob), ShouldBeNil)
		})

		Convey("Timeout during reservation", func() {
			err := queue.Reserve(func(i *Item) error {
				So(i.QueueName, ShouldEqual, queue.Name)
				So(i.Args, ShouldResemble, map[string]interface{}{"Hello": "There"})
				time.Sleep(200 * time.Millisecond)
				return nil
			})
			So(err, ShouldBeError, "context deadline exceeded")
		})

		Convey("Check that the job is left behind", func() {
			err := queue.Reserve(func(i *Item) error {
				So(i.QueueName, ShouldEqual, queue.Name)
				So(i.Args, ShouldResemble, map[string]interface{}{"Hello": "There"})
				return nil
			})
			So(err, ShouldBeNil)
		})

		Convey("Check that the job is now gone", func() {
			err := queue.Reserve(func(i *Item) error { return nil })
			So(err, ShouldBeError, "no rows in result set")
		})
	})

	Convey("Panic job", t, func() {
		Convey("Insert a job in another connection", func() {
			So(queue.Insert(firstJob), ShouldBeNil)
		})

		Convey("Panic during reservation", func() {
			err := queue.Reserve(func(i *Item) error { panic("oh noes") })
			So(err, ShouldBeError, "oh noes")
		})

		Convey("Check that the job has the error next time", func() {
			err := queue.Reserve(func(i *Item) error {
				So(i.QueueName, ShouldEqual, queue.Name)
				So(i.Args, ShouldResemble, map[string]interface{}{"Hello": "There"})
				So(i.Errors, ShouldResemble, []string{"oh noes"})
				return nil
			})
			So(err, ShouldBeNil)
		})

		Convey("Check that the job is now gone", func() {
			err := queue.Reserve(func(i *Item) error { return nil })
			So(err, ShouldBeError, "no rows in result set")
		})
	})
}

func BenchmarkInsert(b *testing.B) {
	firstJob := &Item{Args: map[string]string{"Hello": "There"}}

	queue := makeTestQueue(&testing.T{})
	for n := 0; n < b.N; n++ {
		_ = queue.Insert(firstJob)
	}
}

func BenchmarkReserve(b *testing.B) {
	queue := makeTestQueue(&testing.T{})

	for n := 0; n < b.N; n++ {
		_ = queue.Reserve(func(i *Item) error { return nil })
	}
}

func TestStart(t *testing.T) {
	queue := makeTestQueue(t)
	done := make(chan bool)

	const workCount = 50

	wg := sync.WaitGroup{}
	wg.Add(workCount)

	counter := int64(0)

	failures := make(chan error, workCount)

	go func() {
		err := queue.Start(workCount/2, func(i *Item) error {
			args := i.Args.([]interface{})
			atomic.AddInt64(&counter, int64(args[0].(float64)))
			wg.Done()
			return nil
		})
		if err != nil {
			failures <- err
		}
	}()

	go func() {
		wg.Wait()
		close(failures)
		done <- true
	}()

	// time.Sleep(time.Millisecond)

	for i := 1; i <= workCount; i++ {
		n := i
		_ = queue.Insert(&Item{Args: []int{n}})
	}

	select {
	case <-done:
		expected := ((workCount * workCount) + workCount) / 2
		if int64(counter) != int64(expected) {
			t.Fatalf("Counter should be %d but is %d", expected, counter)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("waiting for worker timed out")
	}

	for failure := range failures {
		t.Fatal(failure)
	}
}
