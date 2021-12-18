package storage

import "github.com/craigpastro/crudapp/myid"

type MemoryDB struct {
	store map[string]map[string]string
}

func NewMemoryDB() Storage {
	return &MemoryDB{store: map[string]map[string]string{}}
}

func (m *MemoryDB) Create(userID, data string) (string, error) {
	if m.store[userID] == nil {
		m.store[userID] = map[string]string{}
	}

	postID := myid.New()
	m.store[userID][postID] = data

	return postID, nil
}

func (m *MemoryDB) Read(userID, postID string) (string, error) {
	posts, ok := m.store[userID]
	if !ok {
		return "", UserDoesNotExist(userID)
	}

	data, ok := posts[postID]
	if !ok {
		return "", PostDoesNotExist(postID)
	}

	return data, nil
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

	posts[postID] = data

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
