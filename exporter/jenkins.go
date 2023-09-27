package exporter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/goodbins/go-jenkins-exporter/config"
	"github.com/sirupsen/logrus"
)

// Jenkins job actions struct
type jActions struct {
	Class                 string `json:"_class"`
	QueuingDurationMillis int    `json:"queuingDurationMillis"`
	TotalDurationMillis   int    `json:"totalDurationMillis"`
	SkipCount             int    `json:"skipCount"`
	FailCount             int    `json:"failCount"`
	TotalCount            int    `json:"totalCount"`
	PassCount             int    `json:"passCount"`
}

// Jenkins job statuses struct
type jStatus struct {
	Class     string `json:"_class"`
	Actions   []jActions
	Duration  int `json:"duration"`
	Number    int `json:"number"`
	Timestamp int `json:"timestamp"`
}

// Jenkins job struct
type job struct {
	Class                 string  `json:"_class"`
	FullName              string  `json:"fullName"`
	URL                   string  `json:"url"`
	LastBuild             jStatus `json:"lastBuild"`
	LastCompletedBuild    jStatus `json:"lastCompletedBuild"`
	LastFailedBuild       jStatus `json:"lastFailedBuild"`
	LastStableBuild       jStatus `json:"lastStableBuild"`
	LastSuccessfulBuild   jStatus `json:"lastSuccessfulBuild"`
	LastUnstableBuild     jStatus `json:"lastUnstableBuild"`
	LastUnsuccessfulBuild jStatus `json:"lastUnsuccessfulBuild"`
}

// Jenkins API response struct
type JenkinsResponse struct {
	Class string `json:"_class"`
	Jobs  []job  `json:"jobs"`
}

var jenkinsFolderClasses = []string{
	"com.cloudbees.hudson.plugins.folder.Folder",
	"jenkins.branch.OrganizationFolder",
	"org.jenkinsci.plugins.workflow.multibranch.WorkflowMultiBranchProject"}

func GetData() *[]job {
	var jobsList []job                 // List of discovered jobs
	var jobFolderLinks []string        // List of job folders
	var jobFolderVisitedLinks []string // List of visited/explored folders
	logrus.Debug("Get data from jenkins..")
	walkAndGetJobs(getJenkinsApiUrl(), &jobsList, &jobFolderLinks, &jobFolderVisitedLinks)
	logrus.Debug("Data retrieved successfully")
	return &jobsList
}

// First url is the API's
func walkAndGetJobs(url string, jobsList *[]job, jobFolderLinks *[]string, jobFolderVisitedLinks *[]string) {
	logrus.Debug("Walking ", url)
	jobs := requestJson(url + "api/json" + createQuery())
	*jobFolderVisitedLinks = append(*jobFolderVisitedLinks, url)
	updateJobsAndFolders(jobs, jobsList, jobFolderLinks)
	for _, fL := range *jobFolderLinks {
		if !isVisited(&fL, *jobFolderVisitedLinks) {
			walkAndGetJobs(fL, jobsList, jobFolderLinks, jobFolderVisitedLinks)
		}
	}
}

func updateJobsAndFolders(reply, jL *[]job, jF *[]string) {
	for _, j := range *reply {
		if isJobsFolder(&j.Class) {
			*jF = append(*jF, j.URL)
			continue
		}
		*jL = append(*jL, j)
	}
}

func isJobsFolder(class *string) bool {
	for _, c := range jenkinsFolderClasses {
		if *class == c {
			return true
		}
	}
	return false
}

func isVisited(link *string, jobFolderVisitedLinks []string) bool {
	for _, j := range jobFolderVisitedLinks {
		if *link == j {
			return true
		}
	}
	return false
}

func requestJson(url string) *[]job {
	// Init a map whose keys are strings and whose values are themselves
	// stored as empty interface values
	var jResp JenkinsResponse
	resp := request(url)
	defer resp.Body.Close()
	// Decode to json the jenkins reply
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(body, &jResp)
	if err != nil {
		logrus.Error("An error has occured while decoding JSON")
		panic(err)
	}
	return &jResp.Jobs
}

func request(apiurl string) *http.Response {
	// Init an http client
	httpClient := &http.Client{Timeout: config.Global.JenkinsAPITimeout * time.Second}
	// Init a http request, set basic auth and Do the request
	req, err := http.NewRequest("GET", apiurl, nil)
	// Test if credentials are used
	if config.Global.JenkinsWithCreds {
		if config.Global.JenkinsPassword != "" {
			req.SetBasicAuth(config.Global.JenkinsUsername, config.Global.JenkinsPassword)
		}
		if config.Global.JenkinsToken != "" {
			req.SetBasicAuth(config.Global.JenkinsUsername, config.Global.JenkinsToken)
		}
	}
	// Make the request
	resp, err := httpClient.Do(req)
	// Control the response code
	logrus.Debug("Request HTTP response code ", resp.StatusCode)
	if resp.StatusCode >= 400 {
		logrus.Error("An error occured while requesting Jenkins. HTTP response code ", resp.StatusCode)
		os.Exit(1)
	}
	// Panic if an error occurs
	if err != nil {
		logrus.Error("An error has occured when getting: ", apiurl)
		panic(err)
	}
	// Return the Jenskins response
	return resp
}

func getJenkinsApiUrl() string {
	var apiurl string = "http://"
	if config.Global.SSLOn {
		apiurl = "https://"
	}
	apiurl += config.Global.JenkinsAPIHostPort + "/"
	return apiurl
}

var jobStatuses = []string{
	"lastBuild",
	"lastCompletedBuild",
	"lastFailedBuild",
	"lastStableBuild",
	"lastSuccessfulBuild",
	"lastUnstableBuild",
	"lastUnsuccessfulBuild",
}

func createQuery() string {

	var jobStatusProperties string = `[
		fullName,
		number,
		timestamp,
		duration,
		actions[
			queuingDurationMillis,
			totalDurationMillis,
			skipCount,
			failCount,
			totalCount,
			passCount]]`

	var query string
	for _, s := range jobStatuses {
		query += "," + s + jobStatusProperties
	}
	return strings.ReplaceAll(strings.ReplaceAll(
		fmt.Sprintf("?tree=jobs[fullName,url%s]", query),
		"\n", ""),
		"\t", "")
}
