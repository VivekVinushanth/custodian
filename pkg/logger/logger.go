package logger

import (
	"log"
	"os"

	"github.com/go-logr/logr"
	"github.com/go-logr/stdr"
)

var Log logr.Logger

func Init() {
	Log = stdr.New(log.New(os.Stderr, "", log.LstdFlags)).WithName("custodian")
}
