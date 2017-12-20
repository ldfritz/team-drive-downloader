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

func DownloadFile(svc *drive.Service, src, dest, mime string) error {
	_, fileID, err := GetFileID(svc, src)
	if err != nil {
		return fmt.Errorf("DownloadFile() -> unable to find file: %v", err)
	}
	file, err := svc.Files.Get(fileID).SupportsTeamDrives(true).Do()
	if err != nil {
		return fmt.Errorf("DownloadFile() -> unable to get file: %v", err)
	}
	var resp *http.Response
	if mime != "" && file.MimeType != mime {
		if !GoogleMIMETypes[file.MimeType] {
			return fmt.Errorf("only Google documents can be converted")
		}
		if !ExportMIMETypes[mime] {
			return fmt.Errorf("unknown export MIME type")
		}
		resp, err = svc.Files.Export(fileID, mime).Download()
		if err != nil {
			return fmt.Errorf("unable to export file: %v", err)
		}
	} else {
		resp, err = svc.Files.Get(fileID).Download()
		if err != nil {
			return fmt.Errorf("DownloadFile() -> unable to download file: %v", err)
		}
	}
	defer resp.Body.Close()
	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("DownloadFile() -> unable to create file: %v", err)
	}
	defer out.Close()
	io.Copy(out, resp.Body)
	return nil
}

var (
	// Google MIME types: https://developers.google.com/drive/v3/web/mime-types
	GoogleMIMETypes = map[string]bool{
		"application/vnd.google-apps.audio":        true,
		"application/vnd.google-apps.document":     true,
		"application/vnd.google-apps.drawing":      true,
		"application/vnd.google-apps.file":         true,
		"application/vnd.google-apps.folder":       true,
		"application/vnd.google-apps.form":         true,
		"application/vnd.google-apps.fusiontable":  true,
		"application/vnd.google-apps.map":          true,
		"application/vnd.google-apps.photo":        true,
		"application/vnd.google-apps.presentation": true,
		"application/vnd.google-apps.script":       true,
		"application/vnd.google-apps.site":         true,
		"application/vnd.google-apps.spreadsheet":  true,
		"application/vnd.google-apps.unknown":      true,
		"application/vnd.google-apps.video":        true,
		"application/vnd.google-apps.drive-sdk":    true,
	}
	// Other MIME types: https://developers.google.com/drive/v3/web/manage-downloads
	ExportMIMETypes = map[string]bool{
		"application/epub+zip":                                                      true,
		"application/pdf":                                                           true,
		"application/rtf":                                                           true,
		"application/vnd.google-apps.script+json":                                   true,
		"application/vnd.oasis.opendocument.presentation":                           true,
		"application/vnd.oasis.opendocument.text":                                   true,
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true,
		"application/x-vnd.oasis.opendocument.spreadsheet":                          true,
		"application/zip":                                                           true,
		"image/jpeg":                                                                true,
		"image/png":                                                                 true,
		"image/svg+xml":                                                             true,
		"text/csv":                                                                  true,
		"text/html":                                                                 true,
		"text/plain":                                                                true,
		"text/tab-separated-values":                                                 true,
	}
)

func MoveFile(svc *drive.Service, src, dest string) error {
	_, fileID, err := GetFileID(svc, src)
	if err != nil {
		return fmt.Errorf("MoveFile() -> unable to find file: %v", err)
	}
	_, oldFolderID, err := GetFolderID(svc, path.Dir(src))
	if err != nil {
		return fmt.Errorf("MoveFile() -> unable to find old parent folder: %v", err)
	}
	_, newFolderID, err := GetFolderID(svc, dest)
	if err != nil {
		return fmt.Errorf("MoveFile() -> unable to find new parent folder file: %v", err)
	}

	_, err = svc.Files.Update(fileID, &drive.File{}).SupportsTeamDrives(true).AddParents(newFolderID).RemoveParents(oldFolderID).Do()
	if err != nil {
		return fmt.Errorf("MoveFile() -> unable to move file: %v", err)
	}
	return nil
}
