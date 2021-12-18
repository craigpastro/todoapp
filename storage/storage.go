package storage

type Storage interface {
	Create(userID, data string) (string, error)
	Read(userID, postID string) (string, error)
	Update(userID, postID, data string) error
	Delete(userID, postID string) error
}

func New(storageType string) (Storage, error) {
	if storageType == "memory" {
		return NewMemoryDB(), nil
	}

	return nil, UndefinedStorageType(storageType)
}
