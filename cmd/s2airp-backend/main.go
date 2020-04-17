package main

import (
	"context"
	"flag"
	"log"
	"strconv"

	"github.com/Tomofiles/s2cells-airspace/pkg/backend"
	"github.com/Tomofiles/s2cells-airspace/pkg/backend/cockroach"

	"github.com/jonboulle/clockwork"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"

	_ "github.com/lib/pq"
)

var (
	address = flag.String("addr", ":8081", "address")

	cockroachParams = struct {
		host            *string
		port            *int
		sslMode         *string
		sslDir          *string
		user            *string
		applicationName *string
	}{
		host:            flag.String("cockroach_host", "", "cockroach host to connect to"),
		port:            flag.Int("cockroach_port", 26257, "cockroach port to connect to"),
		sslMode:         flag.String("cockroach_ssl_mode", "disable", "cockroach sslmode"),
		user:            flag.String("cockroach_user", "root", "cockroach user to authenticate as"),
		sslDir:          flag.String("cockroach_ssl_dir", "", "directory to ssl certificates. Must contain files: ca.crt, client.<user>.crt, client.<user>.key"),
		applicationName: flag.String("cockroach_application_name", "s2airp", "application name for tagging the connection to cockroach"),
	}
)

func main() {
	flag.Parse()

	ctx := context.Background()

	uriParams := map[string]string{
		"host":             *cockroachParams.host,
		"port":             strconv.Itoa(*cockroachParams.port),
		"user":             *cockroachParams.user,
		"ssl_mode":         *cockroachParams.sslMode,
		"ssl_dir":          *cockroachParams.sslDir,
		"application_name": *cockroachParams.applicationName,
	}
	uri, err := cockroach.BuildURI(uriParams)
	if err != nil {
		log.Fatal("Failed to build URI.", err)
	}

	store, err := cockroach.Dial(uri, clockwork.NewRealClock())
	if err != nil {
		log.Fatal("Failed to open connection to CRDB.", uri, err)
	}

	if err := store.Bootstrap(ctx); err != nil {
		log.Fatal("Failed to bootstrap CRDB instance.", err)
	}

	server := backend.Server{
		Store: store,
	}

	e := echo.New()
	e.Use(middleware.CORS())

	e.POST("/upload/did_areas", server.CreateDidAreas)
	e.POST("/upload/airport_areas", server.CreateAirportAreas)
	e.GET("/api/did_areas", server.GetDidAreas)
	e.GET("/api/airport_areas", server.GetAirportAreas)

	err = e.Start(*address)
	if err != nil {
		log.Fatal("Server error.", err)
	}
}
