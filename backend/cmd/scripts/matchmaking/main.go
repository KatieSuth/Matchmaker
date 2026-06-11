// Command matchmaking seeds development event groups for manual matchmaking testing.
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/KatieSuth/MatchmakerAPI/cmd/scripts/common"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "scenarios", "all":
		runScenariosCommand(os.Args[1:])
	case "cleanup":
		runCleanupCommand(os.Args[2:])
	default:
		printUsage()
		os.Exit(1)
	}
}

func runScenariosCommand(args []string) {
	fs := flag.NewFlagSet("scenarios", flag.ExitOnError)
	host := fs.String("host", "", "Discord name of the event host (required)")
	asJSON := fs.Bool("json", false, "Emit results as JSON")
	if err := fs.Parse(args[1:]); err != nil {
		common.Fatal("failed parsing flags", "error", err)
	}
	if *host == "" {
		common.Fatal("--host is required")
	}

	seed := common.NewSeedContext()
	defer seed.Close()

	games := loadSystemGames(seed)
	ensureNoPartialState(seed, games)
	runScenarios(seed, *host, *asJSON)
}

func runCleanupCommand(args []string) {
	fs := flag.NewFlagSet("cleanup", flag.ExitOnError)
	if err := fs.Parse(args); err != nil {
		common.Fatal("failed parsing flags", "error", err)
	}

	seed := common.NewSeedContext()
	defer seed.Close()
	runCleanup(seed)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `Usage:
  go run ./cmd/scripts/matchmaking all --host=YourDiscordName [--json]
  go run ./cmd/scripts/matchmaking scenarios --host=YourDiscordName [--json]
  go run ./cmd/scripts/matchmaking cleanup

Makefile:
  make seed-matchmaking-all HOST=YourDiscordName
  make seed-matchmaking-cleanup
`)
}
