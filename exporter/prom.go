package exporter

import (
	"regexp"
	"strings"
	"time"

	"github.com/goodbins/go-jenkins-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

var prometheusMetrics map[string]*prometheus.GaugeVec

func init() {
	prometheusMetrics = make(map[string]*prometheus.GaugeVec)
	// Loop through statuses to create per status metrics
	for _, s := range jobStatuses {
		// Number
		prometheusMetrics[s+"Number"] = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "jenkins_job_" + toSnakeCase(s) + "_number",
				Help: "Jenkins build number for " + s,
			},
			[]string{
				"jobname",
			},
		)
		// Color
		prometheusMetrics[s+"Color"] = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "jenkins_job_" + toSnakeCase(s) + "_color",
				Help: "Jenkins build color for " + s,
			},
			[]string{
				"jobname",
			},
		)
		// Result
		prometheusMetrics[s+"Result"] = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "jenkins_job_" + toSnakeCase(s) + "_result",
				Help: "Jenkins build result for " + s,
			},
			[]string{
				"jobname",
			},
		)
		// Causes
		prometheusMetrics[s+"Cause"] = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "jenkins_job_" + toSnakeCase(s) + "_cause",
				Help: "Jenkins build cause for " + s,
			},
			[]string{
				"jobname",
			},
		)
		// Duration
		prometheusMetrics[s+"Duration"] = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "jenkins_job_" + toSnakeCase(s) + "_duration_seconds",
				Help: "Jenkins build duration in seconds for " + s,
			},
			[]string{
				"jobname",
			},
		)
		// Timestamp
		prometheusMetrics[s+"Timestamp"] = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "jenkins_job_" + toSnakeCase(s) + "_timestamp_seconds",
				Help: "Jenkins build timestamp in unixtime for " + s,
			},
			[]string{
				"jobname",
			},
		)
		// Queuing duration
		prometheusMetrics[s+"QueuingDuration"] = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "jenkins_job_" + toSnakeCase(s) + "_queuing_duration_seconds",
				Help: "Jenkins build queuing duration in seconds for " + s,
			},
			[]string{
				"jobname",
			},
		)
		// Total duration
		prometheusMetrics[s+"TotalDuration"] = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "jenkins_job_" + toSnakeCase(s) + "_total_duration_seconds",
				Help: "Jenkins build total duration in seconds for " + s,
			},
			[]string{
				"jobname",
			},
		)
		// Skip counts
		prometheusMetrics[s+"SkipCounts"] = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "jenkins_job_" + toSnakeCase(s) + "_skip_count",
				Help: "Jenkins build skip counts for " + s,
			},
			[]string{
				"jobname",
			},
		)
		// Fail counts
		prometheusMetrics[s+"FailCounts"] = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "jenkins_job_" + toSnakeCase(s) + "_fail_count",
				Help: "Jenkins build fail counts for " + s,
			},
			[]string{
				"jobname",
			},
		)
		// Pass counts
		prometheusMetrics[s+"PassCounts"] = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "jenkins_job_" + toSnakeCase(s) + "_pass_count",
				Help: "Jenkins build pass counts for " + s,
			},
			[]string{
				"jobname",
			},
		)
		// Total counts
		prometheusMetrics[s+"TotalCounts"] = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "jenkins_job_" + toSnakeCase(s) + "_total_count",
				Help: "Jenkins build total counts for " + s,
			},
			[]string{
				"jobname",
			},
		)
	}
}

// Get data from Jenkins and update prometheus metrics
func SetGauges() {
	logrus.Debug("Launching metrics update loop: updating rate is set to ", config.Global.MetricsUpdateRate)
	for {
		var jResp *[]job = GetData()
		for _, job := range *jResp {
			jobMetrics := prepareMetrics(&job)
			for key, value := range jobMetrics {
				// Check for older version of the API that doesn't have this JSON attribute
				if job.FullName != "" {
					prometheusMetrics[key].With(prometheus.Labels{"jobname": job.FullName}).Set(value)
				} else {
					prometheusMetrics[key].With(prometheus.Labels{"jobname": job.Name}).Set(value)
				}
			}
		}
		time.Sleep(config.Global.MetricsUpdateRate)
	}
}

func prepareMetrics(job *job) map[string]float64 {
	var jobMetrics = make(map[string]float64, 100)
	// LastBuild
	jobMetrics["lastBuildNumber"] = i2F64(job.LastBuild.Number)
	jobMetrics["lastBuildColor"] = whichColor(job.ColorPtr)
	jobMetrics["lastBuildResult"] = whichResult(job.LastBuild)
	jobMetrics["lastBuildCause"] = whichCause(job.LastBuild)
	jobMetrics["lastBuildDuration"] = i2F64(job.LastBuild.Duration) / 1000.0
	jobMetrics["lastBuildTimestamp"] = i2F64(job.LastBuild.Timestamp) / 1000.0
	jobMetrics["lastBuildQueuingDuration"] = getTimeInQueueValue(job.LastBuild.Actions, "QueuingDuration")
	jobMetrics["lastBuildTotalDuration"] = getTimeInQueueValue(job.LastBuild.Actions, "TotalDuration")
	// if len(job.LastBuild.Actions) == 1 {
	// jobMetrics["lastBuildSkipCount"] = i2F64(job.LastBuild.Actions[0].SkipCount)
	// jobMetrics["lastBuildFailCount"] = i2F64(job.LastBuild.Actions[0].FailCount)
	// jobMetrics["lastBuildTotalCount"] = i2F64(job.LastBuild.Actions[0].TotalCount)
	// jobMetrics["lastBuildPassCount"] = i2F64(job.LastBuild.Actions[0].PassCount)
	// }
	// LastCompletedBuild
	jobMetrics["lastCompletedBuildNumber"] = i2F64(job.LastCompletedBuild.Number)
	jobMetrics["lastCompletedBuildResult"] = whichResult(job.LastCompletedBuild)
	jobMetrics["lastCompletedBuildDuration"] = i2F64(job.LastCompletedBuild.Duration) / 1000
	jobMetrics["lastCompletedBuildTimestamp"] = i2F64(job.LastCompletedBuild.Timestamp) / 1000
	jobMetrics["lastCompletedBuildCause"] = whichCause(job.LastCompletedBuild)
	jobMetrics["lastCompletedBuildQueuingDuration"] = getTimeInQueueValue(job.LastCompletedBuild.Actions, "QueuingDuration")
	jobMetrics["lastCompletedBuildTotalDuration"] = getTimeInQueueValue(job.LastCompletedBuild.Actions, "TotalDuration")
	// if len(job.LastCompletedBuild.Actions) == 1 {
	// jobMetrics["lastCompletedBuildSkipCount"] = i2F64(job.LastCompletedBuild.Actions[0].SkipCount)
	// jobMetrics["lastCompletedBuildFailCount"] = i2F64(job.LastCompletedBuild.Actions[0].FailCount)
	// jobMetrics["lastCompletedBuildTotalCount"] = i2F64(job.LastCompletedBuild.Actions[0].TotalCount)
	// jobMetrics["lastCompletedBuildPassCount"] = i2F64(job.LastCompletedBuild.Actions[0].PassCount)
	// }
	// LastFailedBuild
	jobMetrics["lastFailedBuildNumber"] = i2F64(job.LastFailedBuild.Number)
	jobMetrics["lastFailedBuildResult"] = whichResult(job.LastFailedBuild)
	jobMetrics["lastFailedBuildDuration"] = i2F64(job.LastFailedBuild.Duration) / 1000
	jobMetrics["lastFailedBuildTimestamp"] = i2F64(job.LastFailedBuild.Timestamp) / 1000
	jobMetrics["lastFailedBuildCause"] = whichCause(job.LastFailedBuild)
	jobMetrics["lastFailedBuildQueuingDuration"] = getTimeInQueueValue(job.LastFailedBuild.Actions, "QueuingDuration")
	jobMetrics["lastFailedBuildTotalDuration"] = getTimeInQueueValue(job.LastFailedBuild.Actions, "TotalDuration")
	// if len(job.LastFailedBuild.Actions) == 1 {
	// jobMetrics["lastFailedBuildSkipCount"] = i2F64(job.LastFailedBuild.Actions[0].SkipCount)
	// jobMetrics["lastFailedBuildFailCount"] = i2F64(job.LastFailedBuild.Actions[0].FailCount)
	// jobMetrics["lastFailedBuildTotalCount"] = i2F64(job.LastFailedBuild.Actions[0].TotalCount)
	// jobMetrics["lastFailedBuildPassCount"] = i2F64(job.LastFailedBuild.Actions[0].PassCount)
	// }
	// LastStableBuild
	jobMetrics["lastStableBuildNumber"] = i2F64(job.LastStableBuild.Number)
	jobMetrics["lastStableBuildResult"] = whichResult(job.LastStableBuild)
	jobMetrics["lastStableBuildDuration"] = i2F64(job.LastStableBuild.Duration) / 1000
	jobMetrics["lastStableBuildTimestamp"] = i2F64(job.LastStableBuild.Timestamp) / 1000
	jobMetrics["lastStableBuildCause"] = whichCause(job.LastStableBuild)
	jobMetrics["lastStableBuildQueuingDuration"] = getTimeInQueueValue(job.LastStableBuild.Actions, "QueuingDuration")
	jobMetrics["lastStableBuildTotalDuration"] = getTimeInQueueValue(job.LastStableBuild.Actions, "TotalDuration")
	// if len(job.LastStableBuild.Actions) == 1 {
	// jobMetrics["lastStableBuildSkipCount"] = i2F64(job.LastStableBuild.Actions[0].SkipCount)
	// jobMetrics["lastStableBuildFailCount"] = i2F64(job.LastStableBuild.Actions[0].FailCount)
	// jobMetrics["lastStableBuildTotalCount"] = i2F64(job.LastStableBuild.Actions[0].TotalCount)
	// jobMetrics["lastStableBuildPassCount"] = i2F64(job.LastStableBuild.Actions[0].PassCount)
	// }
	// LastSuccessfulBuild
	jobMetrics["lastSuccessfulBuildNumber"] = i2F64(job.LastSuccessfulBuild.Number)
	jobMetrics["lastSuccessfulBuildResult"] = whichResult(job.LastSuccessfulBuild)
	jobMetrics["lastSuccessfulBuildDuration"] = i2F64(job.LastSuccessfulBuild.Duration) / 1000
	jobMetrics["lastSuccessfulBuildTimestamp"] = i2F64(job.LastSuccessfulBuild.Timestamp) / 1000
	jobMetrics["lastSuccessfulBuildCause"] = whichCause(job.LastSuccessfulBuild)
	jobMetrics["lastSuccessfulBuildQueuingDuration"] = getTimeInQueueValue(job.LastSuccessfulBuild.Actions, "QueuingDuration")
	jobMetrics["lastSuccessfulBuildTotalDuration"] = getTimeInQueueValue(job.LastSuccessfulBuild.Actions, "TotalDuration")
	// if len(job.LastSuccessfulBuild.Actions) == 1 {
	// jobMetrics["lastSuccessfulBuildSkipCount"] = i2F64(job.LastSuccessfulBuild.Actions[0].SkipCount)
	// jobMetrics["lastSuccessfulBuildFailCount"] = i2F64(job.LastSuccessfulBuild.Actions[0].FailCount)
	// jobMetrics["lastSuccessfulBuildTotalCount"] = i2F64(job.LastSuccessfulBuild.Actions[0].TotalCount)
	// jobMetrics["lastSuccessfulBuildPassCount"] = i2F64(job.LastSuccessfulBuild.Actions[0].PassCount)
	// }
	// LastUnstableBuild
	jobMetrics["lastUnstableBuildNumber"] = i2F64(job.LastUnstableBuild.Number)
	jobMetrics["lastUnstableBuildResult"] = whichResult(job.LastUnstableBuild)
	jobMetrics["lastUnstableBuildDuration"] = i2F64(job.LastUnstableBuild.Duration) / 1000
	jobMetrics["lastUnstableBuildTimestamp"] = i2F64(job.LastUnstableBuild.Timestamp) / 1000
	jobMetrics["lastUnstableBuildCause"] = whichCause(job.LastUnstableBuild)
	jobMetrics["lastUnstableBuildQueuingDuration"] = getTimeInQueueValue(job.LastUnstableBuild.Actions, "QueuingDuration")
	jobMetrics["lastUnstableBuildTotalDuration"] = getTimeInQueueValue(job.LastUnstableBuild.Actions, "TotalDuration")
	// if len(job.LastUnstableBuild.Actions) == 1 {
	// jobMetrics["lastUnstableBuildSkipCount"] = i2F64(job.LastUnstableBuild.Actions[0].SkipCount)
	// jobMetrics["lastUnstableBuildFailCount"] = i2F64(job.LastUnstableBuild.Actions[0].FailCount)
	// jobMetrics["lastUnstableBuildTotalCount"] = i2F64(job.LastUnstableBuild.Actions[0].TotalCount)
	// jobMetrics["lastUnstableBuildPassCount"] = i2F64(job.LastUnstableBuild.Actions[0].PassCount)
	// }
	// LastUnsuccessfulBuild
	jobMetrics["lastUnsuccessfulBuildNumber"] = i2F64(job.LastUnsuccessfulBuild.Number)
	jobMetrics["lastUnsuccessfulBuildResult"] = whichResult(job.LastUnsuccessfulBuild)
	jobMetrics["lastUnsuccessfulBuildDuration"] = i2F64(job.LastUnsuccessfulBuild.Duration) / 1000
	jobMetrics["lastUnsuccessfulBuildTimestamp"] = i2F64(job.LastUnsuccessfulBuild.Timestamp) / 1000
	jobMetrics["lastUnsuccessfulBuildCause"] = whichCause(job.LastUnsuccessfulBuild)
	jobMetrics["lastUnsuccessfulBuildQueuingDuration"] = getTimeInQueueValue(job.LastUnsuccessfulBuild.Actions, "QueuingDuration")
	jobMetrics["lastUnsuccessfulBuildTotalDuration"] = getTimeInQueueValue(job.LastUnsuccessfulBuild.Actions, "TotalDuration")
	// if len(job.LastUnsuccessfulBuild.Actions) == 1 {
	// jobMetrics["lastUnsuccessfulBuildSkipCount"] = i2F64(job.LastUnsuccessfulBuild.Actions[0].SkipCount)
	// jobMetrics["lastUnsuccessfulBuildFailCount"] = i2F64(job.LastUnsuccessfulBuild.Actions[0].FailCount)
	// jobMetrics["lastUnsuccessfulBuildTotalCount"] = i2F64(job.LastUnsuccessfulBuild.Actions[0].TotalCount)
	// jobMetrics["lastUnsuccessfulBuildPassCount"] = i2F64(job.LastUnsuccessfulBuild.Actions[0].PassCount)
	// }

	return jobMetrics
}

func whichColor(color *string) float64 {
	switch {
	case color == nil:
		// No value
		return -1
	case *color == "blue" || *color == "blue_anime":
		return 0
	case *color == "red" || *color == "red_anime":
		return 1
	case *color == "yellow" || *color == "yellow_anime":
		return 2
	case *color == "notbuilt" || *color == "notbuilt_anime":
		return 3
	case *color == "disabled" || *color == "disabled_anime":
		return 4
	case *color == "aborted" || *color == "aborted_anime":
		return 5
	case *color == "grey" || *color == "grey_anime":
		return 6
	default:
		// Return for unknown values
		return 100
	}
}

func whichResult(build jStatus) float64 {
	if build.Result == "FAILURE" {
		return 0
	}
	if build.Result == "UNSTABLE" {
		return 0.5
	}
	if build.Result == "SUCCESS" {
		return 1
	}
	if build.Result == "ABORTED" {
		return 2
	}
	// Return a value when the job has no build
	if build.Timestamp == 0 || build.Result == "NOT_BUILT" {
		return 3
	}
	// Return a value when the last job build is running
	if build.Duration == 0 {
		return 4
	}
	// Return for unknown values
	return 100
}

// Return action, by the class given in param
func findActionByClass(actions []jActions, className string) *jActions {
	for _, action := range actions {
		if action.Class == className {
			return &action
		}
	}
	return nil
}

// Return cause action (for older version of the API that doesn't have class attribute)
func findOldCauseAction(actions []jActions) *jActions {
	for _, action := range actions {
		if len(action.Causes) >= 1 {
			return &action
		}
	}
	return nil
}

func whichCause(lastBuild jStatus) float64 {
	causeAction := findActionByClass(lastBuild.Actions, "hudson.model.CauseAction")
	// Case for newer API version
	if causeAction != nil {
		desc := causeAction.Causes[0].Class
		switch {
		// Started by timer or Started by timer with parameters
		case desc == "hudson.triggers.TimerTrigger$TimerTriggerCause" || desc == "org.jenkinsci.plugins.parameterizedscheduler.ParameterizedTimerTriggerCause":
			return 0
		// Started by user
		case desc == "hudson.model.Cause$UserIdCause" || desc == "au.com.centrumsystems.hudson.plugin.buildpipeline.BuildPipelineView$MyUserIdCause":
			return 1
		// Started by upstream project
		case desc == "hudson.model.Cause$UpstreamCause":
			return 2
		case desc == "hudson.triggers.SCMTrigger$SCMTriggerCause":
			return 3
		case desc == "jenkins.branch.BranchIndexingCause":
			return 4
		case desc == "com.dabsquared.gitlabjenkins.cause.GitLabWebHookCause":
			return 5
		// Started from command line
		case desc == "hudson.cli.BuildCommand$CLICause":
			return 6
		// Started by remote host
		case desc == "hudson.model.Cause$RemoteCause":
			return 7
		// Replayed
		case desc == "org.jenkinsci.plugins.workflow.cps.replay.ReplayCause":
			return 8
		// Restarted from build
		case desc == "org.jenkinsci.plugins.pipeline.modeldefinition.causes.RestartDeclarativePipelineCause":
			return 9
		// Push event to branch or Merge request
		case desc == "jenkins.branch.BranchEventCause":
			return 10
		default:
			// Return another value for unknow value
			return 100
		}
	}
	oldCauseAction := findOldCauseAction(lastBuild.Actions)
	// Case for older API version
	if oldCauseAction != nil {
		desc := oldCauseAction.Causes[0].ShortDescription
		switch {
		case strings.HasPrefix(desc, "Started by timer"):
			return 0
		case strings.HasPrefix(desc, "Started by user"):
			return 1
		case strings.HasPrefix(desc, "Started by upstream project"):
			return 2
		case strings.HasPrefix(desc, "Started by an SCM change"):
			return 3
		case strings.HasPrefix(desc, "Started by remote host"):
			return 7
		default:
			// Return another value for unknow value
			return 100
		}
	}
	// Return a value if nil (ex: job with no build or no data)
	return -1
}

func getTimeInQueueValue(actions []jActions, value string) float64 {
	timeInQueueAction := findActionByClass(actions, "jenkins.metrics.impl.TimeInQueueAction")
	if timeInQueueAction != nil {
		switch value {
		case "QueuingDuration":
			return i2F64(timeInQueueAction.QueuingDurationMillis) / 1000
		case "TotalDuration":
			return i2F64(timeInQueueAction.TotalDurationMillis) / 1000
		default:
			// Return another value for unknow value
			return 100
		}
	}
	return -1
}

func i2F64(i int) float64 {
	return float64(i)
}

// Thanks to https://gist.github.com/stoewer/fbe273b711e6a06315d19552dd4d33e6
var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
