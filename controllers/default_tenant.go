package controllers

type DefaultTenant struct {
}

func NewDefaultTenant() *DefaultTenant {
	return &DefaultTenant{}
}

// Check if the database installed
func (d DefaultTenant) Installed() (bool, error) {
	return false, nil
}

func (d DefaultTenant) Install() error {
	return nil
}

// Check if we are still installing the default tenant
func (d DefaultTenant) Installing() (bool, error) {
	return false, nil
}

// Check if the default tenant is installed and ready
func (d DefaultTenant) Ready() (bool, error) {
	return false, nil
}

func (d DefaultTenant) Uninstall() error {
	return nil
}
