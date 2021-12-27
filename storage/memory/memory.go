package memory

import (
	"context"
	"time"

	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
)

type MemoryDB struct {
	store map[string]map[string]*storage.Record
}

func New() storage.Storage {
	return &MemoryDB{store: map[string]map[string]*storage.Record{}}
}

func (m *MemoryDB) Create(ctx context.Context, userID, data string) (string, time.Time, error) {
	if m.store[userID] == nil {
		m.store[userID] = map[string]*storage.Record{}
	}

	postID := myid.New()
	now := time.Now()
	m.store[userID][postID] = storage.NewRecord(userID, postID, data, now, now)

	return postID, now, nil
}

func (m *MemoryDB) Read(ctx context.Context, userID, postID string) (*storage.Record, error) {
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

func (m *MemoryDB) ReadAll(ctx context.Context, userID string) ([]*storage.Record, error) {
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

func (m *MemoryDB) Update(ctx context.Context, userID, postID, data string) (time.Time, error) {
	posts, ok := m.store[userID]
	if !ok {
		return time.Time{}, storage.ErrUserDoesNotExist
	}

	post, ok := posts[postID]
	if !ok {
		return time.Time{}, storage.ErrPostDoesNotExist
	}

	now := time.Now()
	posts[postID] = storage.NewRecord(post.UserID, post.PostID, data, post.CreatedAt, now)

	return now, nil
}

func (m *MemoryDB) Delete(ctx context.Context, userID, postID string) error {
	posts, ok := m.store[userID]
	if !ok {
		return storage.ErrUserDoesNotExist
	}

	delete(posts, postID)

	return nil
}
