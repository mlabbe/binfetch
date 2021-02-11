package objstore

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type Project struct {
	Name     string
	Branches []string
}

type Buildset struct {
	Key           string
	Project       string
	Branch        string
	UnixTimestamp int64
	Tag           string
}

var reBuildset *regexp.Regexp

func NewS3Service(s3region string) *s3.S3 {
	s3Session := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(s3region),
	}))
	s3svc := s3.New(s3Session)

	return s3svc
}

func parseBuildset(str string) *Buildset {
	if reBuildset == nil {
		reBuildset = regexp.MustCompile(`(\w+)\/(\w+)\/(\w+)`)
	}

	match := reBuildset.FindStringSubmatch(str)
	if len(match) != 4 {
		return nil
	}

	var buildset Buildset
	buildset.Key = str
	buildset.Project = match[1]
	buildset.Branch = match[2]

	submatch := strings.Split(match[3], "__")
	if len(submatch) != 2 {
		return nil
	}

	unixTimestamp, err := strconv.ParseInt(submatch[0], 10, 64)
	if err != nil {
		return nil
	}
	buildset.UnixTimestamp = unixTimestamp

	buildset.Tag = submatch[1]

	return &buildset
}

// S3ListProjects returns a list of project names and all corresponding branch names
func S3ListProjects(s3svc *s3.S3, s3bucket string) ([]Project, error) {

	resp, err := s3svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:    &s3bucket,
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return []Project{}, err
	}
	if *resp.IsTruncated {
		// if you have 1000 root projects this could happen.
		return []Project{}, errors.New("Truncated response in S3ListProjects()")
	}

	projects := make([]Project, 0, len(resp.CommonPrefixes))
	for _, prefix := range resp.CommonPrefixes {
		projectName := (*prefix.Prefix)[:len(*prefix.Prefix)-1]
		projectRecord := Project{Name: projectName}

		// get the branches for the project
		respBranch, err := s3svc.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket:    &s3bucket,
			Prefix:    prefix.Prefix,
			Delimiter: aws.String("/"),
		})
		if err != nil {
			return []Project{}, err
		}

		projectRecord.Branches = make([]string, 0, len(respBranch.CommonPrefixes))

		re := regexp.MustCompile(`.+\/(\w+)`)
		for _, branch := range respBranch.CommonPrefixes {

			matches := re.FindStringSubmatch(*branch.Prefix)
			if len(matches) != 2 {
				continue
			}
			projectRecord.Branches = append(projectRecord.Branches, matches[1])
		}

		projects = append(projects, projectRecord)
	}

	return projects, nil
}

// S3ListLatesBuildsetForProject returns the latest buildset object
// path for a project and the unix epoch timestamp that corresponds to
// it
func S3ListLatestBuildsetForProject(s3svc *s3.S3, project, branch, s3bucket string) (*Buildset, error) {

	prefix := fmt.Sprintf("%s/%s/", project, branch)
	respBuildSets, err := s3svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:    &s3bucket,
		Prefix:    &prefix,
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return nil, err
	}

	if *respBuildSets.IsTruncated {
		return nil, errors.New("Truncated response in S3GetLatestBuildsetForProject()")
	}

	// sample common prefix:
	// 'some_project/master/1612926009__35da77044ea797e94c2e6fc8d69b1e1c51e49378/'
	newestBuildsetTimestamp := int64(math.MinInt64)
	var newestBuildsetPtr *string

	re := regexp.MustCompile(`^.+?\/(.+?)\/(\d+)`)
	for _, prefix := range respBuildSets.CommonPrefixes {

		matches := re.FindStringSubmatch(*prefix.Prefix)
		if len(matches) != 3 {
			continue
		}
		matchBranch := matches[1]
		matchTimestamp := matches[2]

		if matchBranch != branch {
			continue
		}

		unixTimestamp, err := strconv.ParseInt(matchTimestamp, 10, 64)
		if err != nil {
			return nil, err
		}

		if unixTimestamp > newestBuildsetTimestamp {
			newestBuildsetTimestamp = unixTimestamp
			newestBuildsetPtr = prefix.Prefix
		}
	}

	if newestBuildsetPtr == nil {
		return nil, errors.New("No matching buildset")
	}

	buildset := parseBuildset(*newestBuildsetPtr)
	if buildset == nil {
		return nil, errors.New(fmt.Sprintf("Could not parse '%s' as Buildset", *newestBuildsetPtr))
	}

	return buildset, nil
}

// S3ListArchiveForBuildset returns the object key for a file that matches os and arch
//
// in the event that os == "macos", an arch "universal" match will be returned
// in the event that multiple os and arch matches are made, the first one is returned
func S3ListArchiveForBuildset(s3svc *s3.S3, buildset, os, arch, s3bucket string) (*string, error) {
	respArchive, err := s3svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:    &s3bucket,
		Prefix:    &buildset,
		Delimiter: aws.String("/"),
	})
	if err != nil {
		return nil, err
	}

	for _, item := range respArchive.Contents {
		if strings.Contains(*item.Key, os) &&
			strings.Contains(*item.Key, arch) {
			return item.Key, nil
		}

	}

	return nil, errors.New(fmt.Sprintf("No matching archive found for '%s', '%s'", os, arch))
}
