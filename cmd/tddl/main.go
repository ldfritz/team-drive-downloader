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

var mainHelp = `usage: tddl COMMAND [ARGUMENTS]

Commands:
  dl SOURCE DESTINATION    Download a Team Drive file.
  ls [DRIVE]/[PATH]        List contents of Drive folder.
  help                     Display this message.
  version                  Display version information.`

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
			fmt.Printf("%s/\n", v.Name)
		}
	case cmd == "ls":
		pathname := os.Args[2]
		files, err := tddl.GetFolderContents(svc, pathname)
		if err != nil {
			log.Fatalln("unable to get path contents:", err)
		}
		for _, v := range files.Files {
			fmt.Print(v.Name)
			if v.MimeType == "application/vnd.google-apps.folder" {
				fmt.Print("/")
			}
			fmt.Print("\n")
		}
	case cmd == "dl":
		if len(os.Args) < 4 {
			fmt.Println("error: need source and destination for download\n")
			fmt.Println(mainHelp)
			return
		}
		src := os.Args[2]
		dest := os.Args[3]
		err := tddl.DownloadFile(src, dest)
		if err != nil {
			log.Fatalln("unable to download file:", err)
		}
	case cmd == "version":
		fmt.Println(tddl.Version)
	default:
		fmt.Println(mainHelp)

	}
}
