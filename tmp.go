package tddl

//func downloadFilesByID(svc *drive.Service, files *drive.FileList) error {
//	destPath := "tmp/"
//	for _, f := range files.Files {
//		dest := destPath + f.Name
//		var resp *http.Response
//		var err error
//		if f.MimeType != "text/csv" {
//			resp, err = svc.Files.Export(f.Id, "text/csv").Download()
//			dest = dest + ".csv"
//		} else {
//			resp, err = svc.Files.Get(f.Id).Download()
//		}
//		if err != nil {
//			return fmt.Errorf("unable to download file: %v", err)
//		}
//		defer resp.Body.Close()
//		out, err := os.Create(dest)
//		if err != nil {
//			return fmt.Errorf("unable to create file: %v", err)
//		}
//		defer out.Close()
//		io.Copy(out, resp.Body)
//	}
//	return nil
//}

//func archiveFiles(svc *drive.Service, driveID, folderID, archiveID string, files *drive.FileList) error {
//	for _, f := range files.Files {
//		_, err := svc.Files.Update(f.Id, &drive.File{}).SupportsTeamDrives(true).AddParents(archiveID).RemoveParents(folderID).Do()
//		if err != nil {
//			return fmt.Errorf("unable to move file: %v", err)
//		}
//	}
//	return nil
//}
