package Docker

import (
    "github.com/docker/docker/api/types"
    "github.com/docker/docker/api/types/container"
    "github.com/docker/docker/api/types/network"
    "github.com/docker/docker/client"
    "github.com/docker/docker/api/types/mount"
    "golang.org/x/net/context"
    "encoding/json"
    "io"
    "sync"
    "os"
    "github.com/pterm/pterm"
    "io/ioutil"
    log "github.com/sirupsen/logrus"
    "math/rand"
)

var (
	runNumber string
)

func PullImage(wg *sync.WaitGroup, progressbar *pterm.SpinnerPrinter, imageName string)  {

    defer wg.Done()

    cli, ctx := factoryClient()

    events, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
    if err != nil {
        log.Fatal(err)
        panic(err)
    }
    defer events.Close()

    if log.GetLevel() == log.DebugLevel {
		io.Copy(os.Stderr, events)
	} else {
		io.Copy(ioutil.Discard, events)
	}

    d := json.NewDecoder(events)
    type Event struct {
        Status         string `json:"status"`
        Error          string `json:"error"`
        Progress       string `json:"progress"`
        ProgressDetail struct {
            Current int `json:"current"`
            Total   int `json:"total"`
        } `json:"progressDetail"`
    }

    var event *Event
    for {
        if err := d.Decode(&event); err != nil {
            if err == io.EOF {
                break
            }

            log.Fatal(err)
            panic(err)
        }

        //progressbar.UpdateText("Pulling docker " + imageName + " (" + event.ProgressDetail.Current + "/" + event.ProgressDetail.Total + ")")

    }

    // Latest event for new image
    // EVENT: {Status:Status: Downloaded newer image for busybox:latest Error: Progress:[==================================================>]  699.2kB/699.2kB ProgressDetail:{Current:699243 Total:699243}}
    // Latest event for up-to-date image
    // EVENT: {Status:Status: Image is up to date for busybox:latest Error: Progress: ProgressDetail:{Current:0 Total:0}}
    if event != nil {
        //if strings.Contains(event.Status, fmt.Sprintf("Downloaded newer image for %s", imageName)) {
        //    // new
        //    return
        //}
//
        //if strings.Contains(event.Status, fmt.Sprintf("Image is up to date for %s", imageName)) {
        //    // up-to-date
        //    return
        //}
    }
}


func RunImage(imageName string, containerName string, mounts []mount.Mount, commands []string) ([]byte, error) {

    cli, ctx := factoryClient()

    // Check first if container is already running. List containers, including stopped ones
    containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: true})
    if err != nil {
        panic(err)
    }
    var alreadyRunning bool = false
    for _, container := range containers {
        if container.Names[0] == "/" + containerName {
            alreadyRunning = true
        }
    }

    if alreadyRunning {
        err = RemoveContainer(containerName)
        if err != nil {
            log.Fatal(err)
            panic(err)
        }
    }

    resp, err := cli.ContainerCreate(
        ctx,
        &container.Config{
            Image: imageName,
            Cmd:   commands,
            Tty:   false,
            AttachStdout: true,
            AttachStderr: true,
        },
        &container.HostConfig{
            AutoRemove: true,
            ReadonlyRootfs: false,
            Mounts: mounts,
            Privileged: true,
        },
        &network.NetworkingConfig{},
        nil,
        containerName)

    if err != nil {
        log.Fatal(err)
        panic(err)
    }

    if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
        log.Fatal(err)
        panic(err)
    }
    return nil, nil
}


func ExecuteInRunningContainer(containerName string, cmdStatement []string) (error) {

    cli, ctx := factoryClient()

    cmdStatementExecuteScript := cmdStatement
    optionsCreateExecuteScript := types.ExecConfig {
        AttachStdout: true,
        AttachStderr: true,
        Cmd: cmdStatementExecuteScript,
    }
    response, err := cli.ContainerExecCreate(ctx, containerName, optionsCreateExecuteScript)
    if err != nil {
        log.Fatal(err)
        panic(err)
    }
    hijackedResponse, err := cli.ContainerExecAttach(ctx, response.ID, types.ExecStartCheck{})
    if err != nil {
        log.Fatal(err)
        panic(err)
    }

    // read out
    if log.GetLevel() == log.DebugLevel {
        // Debug
        io.Copy(os.Stderr, hijackedResponse.Reader)
    } else {
        io.Copy(ioutil.Discard, hijackedResponse.Reader)
    }

    defer hijackedResponse.Close()

    return nil
}


func factoryClient() (*client.Client, context.Context){

    ctx := context.Background()
    cli, err := client.NewClientWithOpts(client.FromEnv)
    cli.NegotiateAPIVersion(ctx) // compatibility
    if err != nil {
        log.Fatal(err)
        panic(err)
    }

    return cli, ctx
}

func RemoveContainer(containerName string) (error) {

    cli, ctx := factoryClient()
    // Stop and remove
    noWaitTimeout := 0 // to not wait for the container to exit gracefully
    if err := cli.ContainerStop(ctx, containerName, container.StopOptions{Timeout: &noWaitTimeout}); err != nil {
        log.Fatal(err)
        return err
    }

    removeOptions := types.ContainerRemoveOptions{
        RemoveVolumes: true,
        Force: true,
    }

    if err := cli.ContainerRemove(ctx, containerName, removeOptions); err != nil {
        log.Fatal(err)
        return err
    }

    return nil
}


var charset = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
func generateRandomString(n int) string {
   b := make([]byte, n)
   	for i := range b {
   		// randomly select 1 character from given charset
   		b[i] = charset[rand.Intn(len(charset))]
   	}
   	return string(b)
}

func getRunNumber() string {
return "unique" // conflicts with mounts
    if runNumber == "" {
        runNumber = generateRandomString(10)
    }
    return runNumber
}