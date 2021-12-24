package memory

import (
	"time"

	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
)

type MemoryDB struct {
	store map[string]map[string]*storage.Record
}

func NewMemoryDB() storage.Storage {
	return &MemoryDB{store: map[string]map[string]*storage.Record{}}
}

func (m *MemoryDB) Create(userID, data string) (string, time.Time, error) {
	if m.store[userID] == nil {
		m.store[userID] = map[string]*storage.Record{}
	}

	postID := myid.New()
	now := time.Now()
	m.store[userID][postID] = storage.NewRecord(userID, postID, data, now)

	return postID, now, nil
}

func (m *MemoryDB) Read(userID, postID string) (*storage.Record, error) {
	records, ok := m.store[userID]
	if !ok {
		return nil, storage.ErrUserDoesNotExist
	}

	record, ok := records[postID]
	if !ok {
		return nil, storage.ErrPostDoesNotExist
	}

	return record, nil
}

func (m *MemoryDB) ReadAll(userID string) ([]*storage.Record, error) {
	records, ok := m.store[userID]
	if !ok {
		return nil, storage.ErrUserDoesNotExist
	}

	res := []*storage.Record{}
	for _, record := range records {
		res = append(res, record)
	}

	return res, nil
}

func (m *MemoryDB) Update(userID, postID, data string) error {
	posts, ok := m.store[userID]
	if !ok {
		return storage.ErrUserDoesNotExist
	}

	_, ok = posts[postID]
	if !ok {
		return storage.ErrPostDoesNotExist
	}

	posts[postID] = storage.NewRecord(userID, postID, data, time.Now())

	return nil
}

func (m *MemoryDB) Delete(userID, postID string) error {
	posts, ok := m.store[userID]
	if !ok {
		return storage.ErrUserDoesNotExist
	}

	delete(posts, postID)

	return nil
}
