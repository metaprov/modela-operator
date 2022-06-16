package controllers

import "testing"

func TestDatabase_Install(t *testing.T) {
	database := NewDatabase()
	installed, err := database.Installed()

}
