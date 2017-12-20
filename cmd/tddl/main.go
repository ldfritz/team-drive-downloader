package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ldfritz/go-helpers/googleservices"
	"github.com/ldfritz/team-drive-downloader"
	"golang.org/x/net/context"
	drive "google.golang.org/api/drive/v3"
)

func main() {
	// Connection boilerplate
	ctx := context.Background()
	secretFilename := "secret"
	tokenFilename := "token"
	permissions := drive.DriveScope

	ctx, config, token, err := googleservices.Authenticate(ctx, secretFilename, tokenFilename, permissions)
	if err != nil {
		log.Fatalln("unable to authenticate:", err)
	}

	svc, err := googleservices.Drive(ctx, config, token)
	if err != nil {
		log.Fatalln("unable to connect to Drive:", err)
	}
	// /Connection boilerplate

	mainHelp := `usage: tddl COMMAND [ARGUMENTS]

Commands:
  ls [DRIVE]/[PATH]    List contents of Drive folder.
  help                 Display help message and exit.
  version              Display version and exit.
`

	if len(os.Args) < 2 {
		fmt.Println(mainHelp)
		return
	}

	cmd := os.Args[1]
	switch {
	case cmd == "ls" && len(os.Args) == 2:
		drives, err := tddl.GetAllTeamDrives(svc)
		if err != nil {
			log.Fatalln("unable to get Team Drives:", err)
		}
		for _, v := range drives {
			fmt.Println(v.Id, v.Name)
		}
	case cmd == "ls":
		pathname := os.Args[2]
		files, err := tddl.GetFolderContents(svc, pathname)
		if err != nil {
			log.Fatalln("unable to get path contents:", err)
		}
		for _, v := range files.Files {
			fmt.Println(v.Id, v.Name)
		}
	case cmd == "version":
		fmt.Println(tddl.Version)
	default:
		fmt.Println(mainHelp)

	}
}
