package controllers

type Database struct {
}

func NewDatabase() *Database {
	return &Database{}
}

// Check if the database installed
func (d Database) Installed() (bool, error) {
	return false, nil
}

func (d Database) Install() error {
	return nil
}

// Check if we are still installing the database
func (d Database) Installing() (bool, error) {
	return false, nil
}

// Check if the database is ready
func (d Database) Ready() (bool, error) {
	return false, nil
}

func (d Database) Uninstall() error {
	return nil
}
