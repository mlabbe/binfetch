package main

/*
   binhost presents builds to a list of users. Builds are expected in the format:

   /$projname/$branch/$epoch__$any_tag_at_all/...

   - $projname is the consistent id for all projects with a given name

   - $epoch is a unix timestapm in utc

   - $any_tag_at_all is separated from epoch by a double underscore,
   and can represent any label. underscores are converted into spaces.

   Builds are hosted in s3.

*/

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"frogtoss.com/binfetch/internal/pkg/objstore"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	app = kingpin.New("binfetch", "Fetch private build artefacts")

	argLs = app.Command("ls", "List available artefacts")

	argGet        = app.Command("get", "Download an artefact")
	argGetProject = argGet.Arg("project", "Project to download").Required().String()
	argGetBranch  = argGet.Flag("branch", "Branch to get").Default("master").String()

	ArgProject = kingpin.Arg("project", "Project name").Required().String()
)

// getHostOSName returns a name string that matches the archive names on binhost
func getHostOSName() string {
	goos := runtime.GOOS
	if goos == "darwin" {
		return "macos"
	}

	if goos == "windows" {
		return "win32"
	}
	return goos
}

func getHostArch() string {
	arch := runtime.GOARCH

	if arch == "386" {
		return "x86"
	}

	return arch
}

func printProjects(projects []objstore.Project) {
	fmt.Printf("Available projects: \n")
	for _, project := range projects {
		fmt.Printf("\t%s\n", project.Name)

		for _, branch := range project.Branches {
			fmt.Printf("\t\t- %s\n", branch)
		}
	}

}

func getDstFileHandleFromSrcPath(srcPath string) (*os.File, error) {
	_, filename := filepath.Split(srcPath)
	fmt.Printf("Downloading %s ...\n", filename)

	handle, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	return handle, nil
}

func main() {

	config := mustParseConfig()

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case argLs.FullCommand():
		s3svc := objstore.NewS3Service(config.S3Region)

		projects, err := objstore.S3ListProjects(s3svc, config.S3Bucket)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		printProjects(projects)
		os.Exit(0)

	case argGet.FullCommand():
		s3svc := objstore.NewS3Service(config.S3Region)

		newestBuildset, err := objstore.S3ListLatestBuildsetForProject(s3svc,
			*argGetProject, *argGetBranch, config.S3Bucket)

		if err != nil {
			fmt.Fprintf(os.Stderr, "get buildset error: %v\n", err)
			os.Exit(1)
		}

		//fmt.Printf("Newest buildset: %+v\n", *newestBuildset)

		fmt.Printf("Found compatible archive built at %v\n", time.Unix(newestBuildset.UnixTimestamp, 0))

		archiveKey, err := objstore.S3ListArchiveForBuildset(s3svc, newestBuildset.Key,
			getHostOSName(), getHostArch(), config.S3Bucket)
		if err != nil {
			fmt.Fprintf(os.Stderr, "get archive error: %v\n", err)
			os.Exit(1)
		}

		fileHandle, err := getDstFileHandleFromSrcPath(*archiveKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "get file handle error: %v\n", err)
			os.Exit(1)
		}

		err = objstore.S3DownloadArchive(config.S3Region, *archiveKey, config.S3Bucket, fileHandle)
		if err != nil {
			fmt.Fprintf(os.Stderr, "get download error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("success.")
	}

}
