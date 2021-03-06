package probe

import (
	"fmt"

	"github.com/docker-slim/docker-slim/internal/app/master/commands"
	"github.com/docker-slim/docker-slim/internal/app/master/docker/dockerclient"
	"github.com/docker-slim/docker-slim/internal/app/master/version"
	"github.com/docker-slim/docker-slim/pkg/command"
	"github.com/docker-slim/docker-slim/pkg/report"
	"github.com/docker-slim/docker-slim/pkg/util/errutil"
	"github.com/docker-slim/docker-slim/pkg/util/fsutil"
	v "github.com/docker-slim/docker-slim/pkg/version"

	log "github.com/sirupsen/logrus"
)

const appName = commands.AppName

// OnCommand implements the 'probe' docker-slim command
func OnCommand(
	gparams *commands.GenericParams,
	targetRef string,
	ec *commands.ExecutionContext) {
	logger := log.WithFields(log.Fields{"app": appName, "command": Name})
	prefix := fmt.Sprintf("cmd=%s", Name)

	viChan := version.CheckAsync(gparams.CheckVersion, gparams.InContainer, gparams.IsDSImage)

	cmdReport := report.NewEditCommand(gparams.ReportLocation, gparams.InContainer)
	cmdReport.State = command.StateStarted

	fmt.Printf("cmd=%s state=started\n", Name)
	fmt.Printf("cmd=%s info=params target=%v\n", Name, targetRef)

	client, err := dockerclient.New(gparams.ClientConfig)
	if err == dockerclient.ErrNoDockerInfo {
		exitMsg := "missing Docker connection info"
		if gparams.InContainer && gparams.IsDSImage {
			exitMsg = "make sure to pass the Docker connect parameters to the docker-slim container"
		}
		fmt.Printf("cmd=%s info=docker.connect.error message='%s'\n", Name, exitMsg)
		fmt.Printf("cmd=%s state=exited version=%s location='%s'\n", Name, v.Current(), fsutil.ExeDir())
		commands.Exit(commands.ECTCommon | commands.ECNoDockerConnectInfo)
	}
	errutil.FailOn(err)

	if gparams.Debug {
		version.Print(prefix, logger, client, false, gparams.InContainer, gparams.IsDSImage)
	}

	fmt.Printf("cmd=%s state=completed\n", Name)
	cmdReport.State = command.StateCompleted

	fmt.Printf("cmd=%s state=done\n", Name)

	vinfo := <-viChan
	version.PrintCheckVersion(prefix, vinfo)

	cmdReport.State = command.StateDone
	if cmdReport.Save() {
		fmt.Printf("cmd=%s info=report file='%s'\n", Name, cmdReport.ReportLocation())
	}
}
