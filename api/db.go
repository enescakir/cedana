package api

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"github.com/cedana/cedana/api/services/task"
	bolt "go.etcd.io/bbolt"
)

type DB struct {
	conn   *bolt.DB
	dbLock sync.Mutex
	dbPath string
}

func NewDB(db *bolt.DB) *DB {
	return &DB{conn: db}
}

func (db *DB) Close() error {
	return db.conn.Close()
}

// KISS for now - but we may want to separate out into subbuckets as we add more
// checkpointing functionality (like incremental checkpointing or GPU checkpointing)
// structure is xid: pid, pid: state
func (db *DB) CreateOrUpdateCedanaProcess(id string, state *task.ProcessState) error {
	return db.conn.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte("default"))
		if err != nil {
			return err
		}

		marshaledState, err := json.Marshal(state)
		if err != nil {
			return err
		}

		pid := state.PID
		if pid == 0 {
			return fmt.Errorf("pid 0 returned from state - is process running?")
		}

		err = root.Put([]byte(id), []byte(strconv.Itoa(int(pid))))
		if err != nil {
			return err
		}

		err = root.Put([]byte(strconv.Itoa(int(pid))), marshaledState)
		if err != nil {
			return err
		}

		return nil
	})
}

func (db *DB) GetStateFromID(id string) (*task.ProcessState, error) {
	var state task.ProcessState

	err := db.conn.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("default"))
		if root == nil {
			return fmt.Errorf("could not find bucket")
		}

		pid := root.Get([]byte(id))
		if pid == nil {
			return fmt.Errorf("could not find pid")
		}

		marshaledState := root.Get(pid)
		if marshaledState == nil {
			return fmt.Errorf("could not find state")
		}

		return json.Unmarshal(marshaledState, &state)
	})

	return &state, err
}

func (db *DB) UpdateProcessStateWithID(id string, state *task.ProcessState) error {
	return db.conn.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte("default"))
		if err != nil {
			return err
		}

		marshaledState, err := json.Marshal(state)
		if err != nil {
			return err
		}

		pid := root.Get([]byte(id))
		if pid == nil {
			return fmt.Errorf("could not find pid")
		}

		return root.Put(pid, marshaledState)
	})
}

func (db *DB) UpdateProcessStateWithPID(pid int32, state *task.ProcessState) error {
	return db.conn.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte("default"))
		if err != nil {
			return err
		}

		marshaledState, err := json.Marshal(state)
		if err != nil {
			return err
		}

		return root.Put([]byte(strconv.Itoa(int(pid))), marshaledState)
	})
}

func (db *DB) setNewDbConn() error {
	// We need an in-memory lock to avoid issues around POSIX file advisory
	// locks as described in the link below:
	// https://www.sqlite.org/src/artifact/c230a7a24?ln=994-1081
	db.dbLock.Lock()

	database, err := bolt.Open(db.dbPath, 0600, nil)
	if err != nil {
		return fmt.Errorf("opening database %s: %w", db.dbPath, err)
	}

	db.conn = database

	return nil
}

func getIDBucket(tx *bolt.Tx) (*bolt.Bucket, error) {
	bkt := tx.Bucket(idRegistryBkt)
	if bkt == nil {
		return nil, fmt.Errorf("id registry bucket not found in DB")
	}
	return bkt, nil
}

func (db *DB) GetPID(id string) (int32, error) {
	var pid int32
	err := db.conn.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("default"))
		if root == nil {
			return fmt.Errorf("could not find bucket")
		}

		pidBytes := root.Get([]byte(id))
		if pidBytes == nil {
			return fmt.Errorf("could not find pid")
		}
		pid = int32(binary.BigEndian.Uint32(pidBytes))

		return nil
	})
	return pid, err
}

func (db *DB) ReturnAllEntries() ([]map[string]string, error) {
	var out []map[string]string
	err := db.conn.View(func(tx *bolt.Tx) error {
		root := tx.Bucket([]byte("default"))
		if root == nil {
			return fmt.Errorf("could not find bucket")
		}

		root.ForEach(func(k, v []byte) error {
			out = append(out, map[string]string{
				string(k): string(v),
			})
			return nil
		})
		return nil
	})
	return out, err
}
