package queue

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gammazero/workerpool"
	"github.com/jackc/pgx"
)

const (
	sqlCreateQueue = `
DROP TABLE IF EXISTS queue;

CREATE TABLE queue(
	id          BIGSERIAL   NOT NULL UNIQUE PRIMARY KEY,
	name        TEXT        NOT NULL DEFAULT 'default',
	created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
	run_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
	args        JSONB       NOT NULL DEFAULT '{}'::json,
	errors      TEXT[]      DEFAULT '{}'
);

CREATE UNIQUE INDEX queue_name ON queue (id, name);
CREATE OR REPLACE FUNCTION notify_queue_inserted()
  RETURNS trigger AS $$
DECLARE
BEGIN
  PERFORM pg_notify(CAST('scylla_queue' AS TEXT), CAST(NEW.name AS text) || ' ' || CAST(NEW.id AS text));
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER queue_insert_notify
	AFTER INSERT ON queue
	FOR EACH ROW EXECUTE PROCEDURE notify_queue_inserted();
`

	sqlReserveItem = `
DELETE FROM queue
	WHERE id = (
		SELECT id
		FROM queue
    WHERE name = $1
		ORDER by id
		FOR UPDATE SKIP LOCKED
		LIMIT 1
	)
RETURNING *;
`

	sqlInsertItem = `
INSERT INTO queue (name, args) VALUES ($1, $2);
`

	sqlSetError = `
UPDATE queue SET errors = errors || ARRAY[$1] WHERE id = $2;
`
)

var logger = log.New(os.Stderr, "[queue] ", log.Lshortfile|log.Ldate|log.Ltime)

type Queue struct {
	Timeout     time.Duration
	CheckEvery  time.Duration // check for unprocessed items (default 3s)
	Pool        *pgx.ConnPool
	Name        string
	Retries     int
	StopWorkers context.CancelFunc
}

// Start runs the given number of goroutines to process work.
// It will use LISTEN to check for new items, but also runs in minute intervals
// to find any orphaned items just to make sure nothing gets left behind.
// This function does not return, so run it in a goroutine if you don't want to
// wait for it.
func (q *Queue) Start(numberOfWorkers int, fun func(*Item) error) error {
	if q.CheckEvery == 0 {
		q.CheckEvery = time.Second * 3
	}

	conn, err := q.Pool.Acquire()
	if err != nil {
		return err
	}
	defer func() {
		q.Pool.Release(conn)
	}()

	// TODO: use a separate channel for each queue
	err = conn.Listen("scylla_queue")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	q.StopWorkers = cancel

	wp := workerpool.New(numberOfWorkers)

	go func() {
		for {
			select {
			case <-ctx.Done():
				wp.Stop()
				return
			case <-time.After(q.CheckEvery):
				wp.SubmitWait(func() {
					_ = q.Reserve(fun)
				})
			}
		}
	}()

	for {
		notification, err := conn.WaitForNotification(ctx)
		if err != nil {
			logger.Println("failed waiting for notification", err)
			return err
		}

		words := strings.Split(notification.Payload, " ")
		if words[0] != q.Name {
			continue
		}

		wp.Submit(func() {
			err := q.Reserve(fun)
			if err != nil {
				if err.Error() == "no rows in result set" {
					return
				}
				logger.Println("reservation failed", err)
			}
		})
	}

	wp.StopWait()
	return nil
}

func (q Queue) Insert(i *Item) error {
	ctx, cancel := context.WithTimeout(context.Background(), q.Timeout)
	tx, err := q.Pool.BeginEx(ctx, nil)
	if err != nil {
		cancel()
		return err
	}
	defer func() {
		_ = tx.RollbackEx(ctx)
		cancel()
	}()

	_, err = tx.ExecEx(ctx, sqlInsertItem, nil, q.Name, i.Args)
	if err != nil {
		return err
	}
	return tx.CommitEx(ctx)
}

func (q Queue) Reserve(fun func(*Item) error) (err error) {
	i := &Item{}

	ctx, cancel := context.WithTimeout(context.Background(), q.Timeout)
	var tx *pgx.Tx
	tx, err = q.Pool.BeginEx(ctx, nil)
	if err != nil {
		cancel()
		return err
	}
	defer func() {
		_ = tx.RollbackEx(ctx)
		cancel()
		if r := recover(); r != nil {
			switch rt := r.(type) {
			case error:
				err = rt
			case string:
				err = errors.New(rt)
			default:
				err = fmt.Errorf("%t %#v", rt, rt)
			}
		}

		if err != nil {
			_ = q.setError(err, i)
		}
	}()

	row := tx.QueryRowEx(ctx, sqlReserveItem, nil, q.Name)
	err = row.Scan(&i.ID, &i.QueueName, &i.CreatedAt, &i.RunAt, &i.Args, &i.Errors)
	if err != nil {
		return err
	}

	err = fun(i)
	if err != nil {
		return err
	}

	err = tx.CommitEx(ctx)
	return err
}

func (q Queue) setError(e error, i *Item) error {
	ctx, cancel := context.WithTimeout(context.Background(), q.Timeout)
	defer func() { cancel() }()
	tx, err := q.Pool.BeginEx(ctx, nil)
	if err != nil {
		logger.Println(err)
		return err
	}

	_, err = tx.ExecEx(ctx, sqlSetError, nil, e.Error(), i.ID)
	if err != nil {
		logger.Println("While storing", e, ":", err)
		return err
	}

	err = tx.CommitEx(ctx)
	if err != nil {
		logger.Println(err)
	}

	return err
}

type Item struct {
	ID        int64       `json:"id"`
	QueueName string      `json:"name"`
	CreatedAt time.Time   `json:"created_at"`
	RunAt     time.Time   `json:"run_at"`
	Args      interface{} `json:"args"`
	Errors    []string    `json:"errors"`
}

func prepareDatabase(db *pgx.Tx) error {
	_, err := db.Exec(sqlCreateQueue)
	return err
}
