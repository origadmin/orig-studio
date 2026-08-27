package main

import (
	"os"

	"origadmin/application/origstudio/internal/migrate"
)

// main is the deployment vehicle for the offline migration batch job
// (see internal/migrate.RunCLI). It builds to origcms-migrate and runs as a
// one-shot task whose Docker entrypoint is /app/bin/origcms-migrate. EE is a
// microservice architecture with no `server` monolith, so migration ships as
// this standalone binary instead of a server subcommand.
func main() {
	os.Exit(migrate.RunCLI(os.Args[1:]))
}
