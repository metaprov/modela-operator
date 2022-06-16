package controllers

// Modela system represent the model core system
type Monitoring struct {
}

func NewMonitoring() *Monitoring {
	return &Monitoring{}
}

// Check if the database installed
func (d Monitoring) Installed() (bool, error) {
	return false, nil
}

func (d Monitoring) Install() error {
	return nil
}

// Check if we are still installing the database
func (d Monitoring) Installing() (bool, error) {
	return false, nil
}

// Check if the database is ready
func (d Monitoring) Ready() (bool, error) {
	return false, nil
}

func (d Monitoring) Uninstall() error {
	return nil
}
