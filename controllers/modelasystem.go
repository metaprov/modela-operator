package controllers

// Modela system represent the model core system
type ModelaSystem struct {
}

func NewModelaSystem() *ModelaSystem {
	return &ModelaSystem{}
}

// Check if the database installed
func (d ModelaSystem) Installed() (bool, error) {
	return false, nil
}

func (d ModelaSystem) Install() error {
	return nil
}

// Check if we are still installing the database
func (d ModelaSystem) Installing() (bool, error) {
	return false, nil
}

// Check if the database is ready
func (d ModelaSystem) Ready() (bool, error) {
	return false, nil
}

func (d ModelaSystem) Uninstall() error {
	return nil
}
