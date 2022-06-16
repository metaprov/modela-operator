package controllers

// Modela system represent the model core system
type ObjectStorage struct {
}

func NewObjectStorage() *ObjectStorage {
	return &ObjectStorage{}
}

// Check if the database installed
func (d ObjectStorage) Installed() (bool, error) {
	return false, nil
}

func (d ObjectStorage) Install() error {
	return nil
}

// Check if we are still installing the database
func (d ObjectStorage) Installing() (bool, error) {
	return false, nil
}

// Check if the database is ready
func (d ObjectStorage) Ready() (bool, error) {
	return false, nil
}

func (d ObjectStorage) Uninstall() error {
	return nil
}
