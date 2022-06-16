package controllers

// Modela system represent the model core system
type CertManager struct {
}

func NewCertManager() *CertManager {
	return &CertManager{}
}

// Check if the database installed
func (d CertManager) Installed() (bool, error) {
	return false, nil
}

func (d CertManager) Install() error {
	return nil
}

// Check if we are still installing the database
func (d CertManager) Installing() (bool, error) {
	return false, nil
}

// Check if the database is ready
func (d CertManager) Ready() (bool, error) {
	return false, nil
}

func (d CertManager) Uninstall() error {
	return nil
}
