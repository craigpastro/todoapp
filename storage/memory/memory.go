package memory

import (
	"context"
	"errors"
	"time"

	"github.com/craigpastro/crudapp/myid"
	"github.com/craigpastro/crudapp/storage"
	"go.opentelemetry.io/otel/trace"
)

var _ storage.Storage = (*MemoryDB)(nil)

type MemoryDB struct {
	store  map[string]map[string]*storage.Record
	tracer trace.Tracer
}

func New(tracer trace.Tracer) *MemoryDB {
	return &MemoryDB{
		store:  map[string]map[string]*storage.Record{},
		tracer: tracer,
	}
}

func (m *MemoryDB) Create(ctx context.Context, userID, data string) (*storage.Record, error) {
	_, span := m.tracer.Start(ctx, "memory.Create")
	defer span.End()

	if m.store[userID] == nil {
		m.store[userID] = map[string]*storage.Record{}
	}

	postID := myid.New()
	now := time.Now()
	record := storage.NewRecord(userID, postID, data, now, now)
	m.store[userID][postID] = record

	return record, nil
}

func (m *MemoryDB) Read(ctx context.Context, userID, postID string) (*storage.Record, error) {
	_, span := m.tracer.Start(ctx, "memory.Read")
	defer span.End()

	record, ok := m.store[userID][postID]
	if !ok {
		return nil, storage.ErrPostDoesNotExist
	}

	return record, nil
}

func (m *MemoryDB) ReadAll(ctx context.Context, userID string) (storage.RecordIterator, error) {
	_, span := m.tracer.Start(ctx, "memory.ReadAll")
	defer span.End()

	records := m.store[userID]
	res := []*storage.Record{}
	for _, record := range records {
		res = append(res, record)
	}

	return &recordInterator{records: res}, nil
}

type recordInterator struct {
	records []*storage.Record
}

func (i *recordInterator) Next(_ context.Context) bool {
	return len(i.records) > 0
}

func (i *recordInterator) Get(dest *storage.Record) error {
	if len(i.records) == 0 {
		return errors.New("no more records")
	}

	*dest = *i.records[0]
	i.records = i.records[1:]
	return nil
}

func (i *recordInterator) Close(_ context.Context) {
	i.records = nil
}

func (m *MemoryDB) Update(ctx context.Context, userID, postID, data string) (time.Time, error) {
	_, span := m.tracer.Start(ctx, "memory.Update")
	defer span.End()

	post, ok := m.store[userID][postID]
	if !ok {
		return time.Time{}, storage.ErrPostDoesNotExist
	}

	now := time.Now()
	m.store[userID][postID] = storage.NewRecord(post.UserID, post.PostID, data, post.CreatedAt, now)

	return now, nil
}

func (m *MemoryDB) Delete(ctx context.Context, userID, postID string) error {
	_, span := m.tracer.Start(ctx, "memory.Delete")
	defer span.End()

	delete(m.store[userID], postID)
	return nil
}
