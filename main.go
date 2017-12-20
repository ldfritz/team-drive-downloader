package tddl

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"

	drive "google.golang.org/api/drive/v3"
)

func GetAllTeamDrives(svc *drive.Service) ([]*drive.TeamDrive, error) {
	resp, err := svc.Teamdrives.List().Do()
	if err != nil {
		return []*drive.TeamDrive{}, fmt.Errorf("unable to list Team Drives: %v", err)
	}
	return resp.TeamDrives, nil
}

func GetFolderContents(svc *drive.Service, pathname string) (*drive.FileList, error) {
	driveID, folderID, err := GetFolderID(svc, pathname)
	if err != nil {
		return nil, fmt.Errorf("unable to find folder: %v", err)
	}
	files, err := ListFiles(svc, driveID, folderID)
	if err != nil {
		return nil, fmt.Errorf("unable to list contets: %v", err)
	}
	return files, nil
}

func GetTeamDriveByName(svc *drive.Service, name string) (*drive.TeamDrive, error) {
	var td *drive.TeamDrive
	drives, err := GetAllTeamDrives(svc)
	if err != nil {
		return td, fmt.Errorf("unable to list Drives: %v", err)
	}
	for _, v := range drives {
		if v.Name == name {
			td = v
		}
	}
	if td == nil {
		return td, fmt.Errorf("Team Drive not found: %s", name)
	}
	return td, nil
}

func GetFolderID(svc *drive.Service, pathname string) (string, string, error) {
	ps := slicePath(pathname)
	td, err := GetTeamDriveByName(svc, ps[0])
	if err != nil {
		return "", "", fmt.Errorf("unable to get Team Drive: %v", err)
	}
	driveID := td.Id
	folderID := driveID
	for _, v := range ps[1:] {
		resp, err := svc.Files.List().Corpora("teamDrive").TeamDriveId(driveID).IncludeTeamDriveItems(true).SupportsTeamDrives(true).Q(fmt.Sprintf("mimeType='application/vnd.google-apps.folder' and '%s' in parents and trashed=false and name='%s'", folderID, v)).Do()
		if err != nil {
			return driveID, "", fmt.Errorf("unable to get folder ID: %v", err)
		}
		if len(resp.Files) == 0 {
			return driveID, "", fmt.Errorf("unable to find folder: %v", v)
		}
		folderID = resp.Files[0].Id
	}
	return driveID, folderID, nil
}

func GetFileID(svc *drive.Service, pathname string) (string, string, error) {
	pathname = strings.TrimRight(pathname, "/")
	driveID, folderID, err := GetFolderID(svc, path.Dir(pathname))
	filename := path.Base(pathname)
	if err != nil {
		return driveID, folderID, fmt.Errorf("unable to find folder: %v", err)
	}
	resp, err := svc.Files.List().Corpora("teamDrive").TeamDriveId(driveID).IncludeTeamDriveItems(true).SupportsTeamDrives(true).Q(fmt.Sprintf("'%s' in parents and trashed=false and name='%s'", folderID, filename)).Do()
	if err != nil {
		return driveID, "", fmt.Errorf("unable to get file ID: %v", err)
	}
	if len(resp.Files) == 0 {
		return driveID, "", fmt.Errorf("unable to find folder: %v", filename)
	}
	fileID := resp.Files[0].Id
	return driveID, fileID, nil
}

func ListFiles(svc *drive.Service, driveID, folderID string) (*drive.FileList, error) {
	resp, err := svc.Files.List().Corpora("teamDrive").TeamDriveId(driveID).IncludeTeamDriveItems(true).SupportsTeamDrives(true).Q(fmt.Sprintf("'%s' in parents and trashed=false", folderID)).Do()
	if err != nil {
		return resp, fmt.Errorf("unable to get files: %v", err)
	}
	return resp, nil
}

func slicePath(pathname string) []string {
	var ps []string
	p := path.Clean(pathname)
	p = strings.Trim(p, "/.")
	for p != "" {
		d, f := path.Split(p)
		ps = append(ps, f)
		p = strings.TrimRight(d, "/")
	}
	for i := range ps[:len(ps)/2] {
		ps[i], ps[len(ps)-1-i] = ps[len(ps)-1-i], ps[i]
	}
	return ps
}

func DownloadFile(svc *drive.Service, src, dest string) error {
	// I need to re-implement the export conversions
	// Google MIME types: https://developers.google.com/drive/v3/web/mime-types
	// Other MIME types: https://developers.google.com/drive/v3/web/manage-downloads
	_, fileID, err := GetFileID(svc, src)
	if err != nil {
		return fmt.Errorf("unable to find file: %v", err)
	}
	var resp *http.Response
	resp, err = svc.Files.Get(fileID).Download()
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
	return nil
}
