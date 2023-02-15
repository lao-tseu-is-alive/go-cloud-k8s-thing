package objects

import (
	"github.com/lao-tseu-is-alive/go-cloud-k8s-common-libs/pkg/database"
	"log"
)

type Service struct {
	Log         *log.Logger
	dbConn      database.DB
	JwtSecret   []byte
	JwtDuration int
}
