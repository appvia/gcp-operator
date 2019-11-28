package controller

import (
	"github.com/appvia/gcp-operator/pkg/controller/gcpproject"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, gcpproject.Add)
}
