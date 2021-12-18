package storage

import (
	"time"

	"github.com/craigpastro/crudapp/myid"
)

type MemoryDB struct {
	store map[string]map[string]*record
}

func NewMemoryDB() Storage {
	return &MemoryDB{store: map[string]map[string]*record{}}
}

func (m *MemoryDB) Create(userID, data string) (string, time.Time, error) {
	if m.store[userID] == nil {
		m.store[userID] = map[string]*record{}
	}

	postID := myid.New()
	now := time.Now()
	m.store[userID][postID] = NewRecord(userID, postID, data, now)

	return postID, now, nil
}

func (m *MemoryDB) Read(userID, postID string) (*record, error) {
	records, ok := m.store[userID]
	if !ok {
		return nil, UserDoesNotExist(userID)
	}

	record, ok := records[postID]
	if !ok {
		return nil, PostDoesNotExist(postID)
	}

	return record, nil
}

func (m *MemoryDB) ReadAll(userID string) ([]*record, error) {
	records, ok := m.store[userID]
	if !ok {
		return nil, UserDoesNotExist(userID)
	}

	res := []*record{}
	for _, record := range records {
		res = append(res, record)
	}

	return res, nil
}

func (m *MemoryDB) Update(userID, postID, data string) error {
	posts, ok := m.store[userID]
	if !ok {
		return UserDoesNotExist(userID)
	}

	_, ok = posts[postID]
	if !ok {
		return PostDoesNotExist(postID)
	}

	record := posts[postID]
	record.Data = data

	return nil
}

func (m *MemoryDB) Delete(userID, postID string) error {
	posts, ok := m.store[userID]
	if !ok {
		return UserDoesNotExist(userID)
	}

	delete(posts, postID)

	return nil
}
