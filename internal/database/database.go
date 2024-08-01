package database

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

type Chirp struct {
	id   int
	Body string `json:"body"`
}

type DB struct {
	path string
	mux  sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

func NewDB(path string) (*DB, error) {
	db := DB{
		path: path,
	}
	err := db.ensure()
	if err != nil {
		return &DB{}, err
	}
	return &db, nil
}

func (db *DB) ensure() error {
	db.mux.RLock()
	defer db.mux.RUnlock()
	if _, err := os.ReadFile(db.path); err == nil {
		return nil
	}
	err := os.WriteFile(db.path, nil, 0666) //file mode: permission to read write for everybody
	if err != nil {
		return errors.New("no database found and failed to create one")
	}
	return nil
}

func (db *DB) CreateChirp(body string) (Chirp, error) {
	dbStruct, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}
	id := assignId()() // clojure every call + 1
	dbStruct.Chirps[id] = Chirp{
		id:   id,
		Body: body,
	}
	db.mux.Lock()
	err2 := db.write(dbStruct)
	db.mux.Unlock()
	if err2 != nil {
		return Chirp{}, err
	}
	return dbStruct.Chirps[id], nil
}

func (db *DB) loadDB() (DBStructure, error) {
	dbStruct := DBStructure{}
	db.mux.RLock()
	defer db.mux.RUnlock()

	if dat, err := os.ReadFile(db.path); err != nil {
		return DBStructure{}, err
	} else if err := json.Unmarshal(dat, &dbStruct); err != nil {
		return DBStructure{}, err
	} else {
		return dbStruct, nil
	}
}

func (db *DB) write(dbStruct DBStructure) error {
	dat, err := json.Marshal(dbStruct)
	if err != nil {
		return err
	}
	db.mux.Lock()
	defer db.mux.Unlock()
	er := os.WriteFile(db.path, dat, 0666)
	if er != nil {
		return er
	}
	return nil
}

func (db *DB) GetChirps() ([]Chirp, error) {
	if dbStruct, err := db.loadDB(); err != nil {
		return nil, err
	} else {
		chirps := []Chirp{}
		for i := 1; i < len(dbStruct.Chirps); i++ {
			chirps = append(chirps, dbStruct.Chirps[i])
		}
		return chirps, nil
	}
}

func assignId() func() int {
	counter := 0

	return func() int {
		counter++
		return counter
	}
}
