# team-drive-downloader

Download files from a Team Drive folder.

Rather than always upload CSV files to an SFTP server, I can let internal users post files to our shared Team Drive.
They just need to post the files into the appropriate folder.
When this script is run it will download all the files from the specified folder.
If the file is not a CSV, it uses Google's export functionality to convert it into a CSV file.
It then moves the files into an archive subfolder.

## Installation

(I still need to walk through this to ensure it is accurate.)

```
go get -u github.com/ldfritz/team-drive-downloader
```

This should grab and build the script and its dependencies.

## Usage

In your bash script just add something like the following.
Just insert the correct values for your Team Drive and upload folder.

```
team-drive-downloader 'Team Drive name here' 'folder name here'
```
