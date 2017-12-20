package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/ldfritz/go-helpers/googleservices"
	"github.com/ldfritz/team-drive-downloader"
	"golang.org/x/net/context"
	drive "google.golang.org/api/drive/v3"
)

var mainHelp = `usage: tddl [OPTIONS] COMMAND [ARGUMENTS]

Commands:
  dl SOURCE DESTINATION    Download a Team Drive file.
  ls [DRIVE]/[PATH]        List contents of Drive folder.
  mv SOURCE DESTINATION    Move a Team Drive file to a new folder.
  help                     Display this message.
  version                  Display version information.`

func main() {
	flagHelp := flag.Bool("help", false, "display help")
	flagH := flag.Bool("h", false, "display help")
	flagMIME := flag.String("mime", "", "set export MIME type")
	flagFiles := flag.Bool("files", false, "target files exclusively")
	flagF := flag.Bool("f", false, "target files exclusively")
	flag.Parse()

	opts := tddl.Options{}
	opts.Help = *flagHelp || *flagH
	opts.MIME = *flagMIME
	opts.Files = *flagFiles || *flagF

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

	if len(flag.Args()) == 0 || opts.Help {
		fmt.Println(mainHelp)
		return
	}

	cmd := flag.Arg(0)
	switch {
	case cmd == "ls" && len(flag.Args()) == 1: // only arg is cmd
		drives, err := tddl.GetAllTeamDrives(svc)
		if err != nil {
			log.Fatalln("unable to get Team Drives:", err)
		}
		for _, v := range drives {
			fmt.Printf("%s/\n", v.Name)
		}
	case cmd == "ls":
		pathname := flag.Arg(1)
		files, err := tddl.GetFolderContents(svc, pathname, opts)
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
		if len(flag.Args()) < 3 { // need arg, src, dest
			fmt.Println("error: need source and destination for download\n")
			fmt.Println(mainHelp)
			return
		}
		src := flag.Arg(1)
		dest := flag.Arg(2)
		err := tddl.DownloadFile(svc, src, dest, opts)
		if err != nil {
			log.Fatalln("unable to download file:", err)
		}
	case cmd == "mv":
		if len(flag.Args()) < 3 { // need arg, src, dest
			fmt.Println("error: need source and destination for move\n")
			fmt.Println(mainHelp)
			return
		}
		src := flag.Arg(1)
		dest := flag.Arg(2)
		err := tddl.MoveFile(svc, src, dest)
		if err != nil {
			log.Fatalln("unable to move file:", err)
		}
	case cmd == "version":
		fmt.Println(tddl.Version)
	default:
		fmt.Println(mainHelp)
	}
}
