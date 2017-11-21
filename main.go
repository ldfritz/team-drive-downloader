package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/ldfritz/go-helpers/googleservices"
	"golang.org/x/net/context"
	drive "google.golang.org/api/drive/v3"
)

func main() {
	// Connection boilerplate
	ctx := context.Background()
	secretFilename := "secret"
	tokenFilename := "token"
	permissions := drive.DriveScope

	teamDriveName := os.Args[1]
	folderName := os.Args[2]

	ctx, config, token, err := googleservices.Authenticate(ctx, secretFilename, tokenFilename, permissions)
	if err != nil {
		log.Print("unable to authenticate: ", err)
		return
	}

	svc, err := googleservices.Drive(ctx, config, token)
	if err != nil {
		log.Print("unable to connect to Drive: ", err)
		return
	}
	// /Connection boilerplate

	driveID, err := getTeamDriveByName(svc, teamDriveName)
	if err != nil {
		log.Print(err)
		return
	}

	folderID, err := getTeamDriveFolderByName(svc, driveID, folderName)
	if err != nil {
		log.Print(err)
		return
	}

	files, err := listFiles(svc, driveID, folderID)
	if err != nil {
		log.Print(err)
		return
	}

	err = downloadFilesByID(svc, files)
	if err != nil {
		log.Print(err)
		return
	}

	archiveID, err := getArchiveSubFolder(svc, driveID, folderID)
	if err != nil {
		log.Print(err)
		return
	}

	err = archiveFiles(svc, driveID, folderID, archiveID, files)
	if err != nil {
		log.Print(err)
		return
	}
}

func getTeamDriveByName(svc *drive.Service, name string) (string, error) {
	var driveID string
	resp, err := svc.Teamdrives.List().Do()
	if err != nil {
		return "", fmt.Errorf("unable to list files: %v", err)
	}
	for _, f := range resp.TeamDrives {
		if f.Name == name {
			driveID = f.Id
		}
	}
	if driveID == "" {
		return "", fmt.Errorf("Team Drive not found: %s", name)
	}
	return driveID, nil
}

func getTeamDriveFolderByName(svc *drive.Service, driveID, name string) (string, error) {
	var folderID string
	resp, err := svc.Files.List().Corpora("teamDrive").TeamDriveId(driveID).IncludeTeamDriveItems(true).SupportsTeamDrives(true).Q(fmt.Sprintf("'%s' in parents and trashed = false", driveID)).Do()
	if err != nil {
		return "", fmt.Errorf("unable to get folder IDs: %v", err)
	}
	for _, f := range resp.Files {
		if f.Name == name {
			folderID = f.Id
		}
	}
	if folderID == "" {
		return "", fmt.Errorf("folder not found: %s", name)
	}
	return folderID, nil
}

func getArchiveSubFolder(svc *drive.Service, driveID, folderID string) (string, error) {
	var archiveID string
	name := "archive"
	resp, err := svc.Files.List().Corpora("teamDrive").TeamDriveId(driveID).IncludeTeamDriveItems(true).SupportsTeamDrives(true).Q(fmt.Sprintf("mimeType='application/vnd.google-apps.folder' and '%s' in parents and trashed=false and name='archive'", folderID)).Do()
	if err != nil {
		return "", fmt.Errorf("unable to get archive folder: %v", err)
	}
	for _, f := range resp.Files {
		if f.Name == name {
			archiveID = f.Id
		}
	}
	if archiveID == "" {
		return "", fmt.Errorf("folder not found: %s", name)
	}
	return archiveID, nil
}

func listFiles(svc *drive.Service, driveID, folderID string) (*drive.FileList, error) {
	resp, err := svc.Files.List().Corpora("teamDrive").TeamDriveId(driveID).IncludeTeamDriveItems(true).SupportsTeamDrives(true).Q(fmt.Sprintf("mimeType!='application/vnd.google-apps.folder' and '%s' in parents and trashed=false", folderID)).Do()
	if err != nil {
		return &drive.FileList{}, fmt.Errorf("unable to get files: %v", err)
	}
	return resp, nil
}

func downloadFilesByID(svc *drive.Service, files *drive.FileList) error {
	destPath := "tmp/"
	for _, f := range files.Files {
		dest := destPath + f.Name
		var resp *http.Response
		var err error
		if f.MimeType != "text/csv" {
			resp, err = svc.Files.Export(f.Id, "text/csv").Download()
			dest = dest + ".csv"
		} else {
			resp, err = svc.Files.Get(f.Id).Download()
		}
		if err != nil {
			return fmt.Errorf("unable to download file: %v", err)
		}
		defer resp.Body.Close()
		out, err := os.Create(dest)
		if err != nil {
			return fmt.Errorf("unable to create file: %v", err)
		}
		defer out.Close()
		io.Copy(out, resp.Body)
	}
	return nil
}

func archiveFiles(svc *drive.Service, driveID, folderID, archiveID string, files *drive.FileList) error {
	for _, f := range files.Files {
		_, err := svc.Files.Update(f.Id, &drive.File{}).SupportsTeamDrives(true).AddParents(archiveID).RemoveParents(folderID).Do()
		if err != nil {
			return fmt.Errorf("unable to move file: %v", err)
		}
	}
	return nil
}
