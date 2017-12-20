# team-drive-downloader

Download files from a Team Drive folder.

Rather than always upload CSV files to an SFTP server, I can let internal users post files to our shared Team Drive.
They just need to post the files into the appropriate folder.
When this script is run it will download all the files from the specified folder.
If the file is not a CSV, it uses Google's export functionality to convert it into a CSV file.
It then moves the files into an archive subfolder.

## Get it

```
go get -u github.com/ldfritz/team-drive-downloader/cmd/tddl

```

## Configure it

Er... this needs to be documented.
It currently involves downloading an API token and then authenticating it the first time.
And then ensuring those are in correctly named, local files.

## Use it

```
tddl help
```
