package docker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/go-logr/logr"
)

type ErrorLine struct {
	Error       string      `json:"error"`
	ErrorDetail ErrorDetail `json:"errorDetail"`
}

type ErrorDetail struct {
	Message string `json:"message"`
}

type Dock struct {
	client *client.Client
	log    logr.Logger
}

func New(log logr.Logger) (*Dock, error) {
	// TODO: change this to take  options
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("docker.New: %w", err)
	}

	log = log.WithName("docker")

	dock := Dock{
		client: cli,
		log:    log,
	}

	return &dock, nil
}

// BuildContainer: builds container with name tag of `name`, username of `user`, and password of `password`
func (d Dock) BuildContainer(name, user, password, contextDir string) error {
	options := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		NoCache:    false,
		Remove:     false,
		Tags:       []string{name},
		BuildArgs:  map[string]*string{"USER": &user, "PASS": &password},
		Labels:     map[string]string{},
	}

	buildContext, err := makeBuildContext(contextDir)
	if err != nil {
		return fmt.Errorf("docker.BuildContainer: %w", err)
	}

	resp, err := d.client.ImageBuild(context.TODO(), *buildContext, options)

	scanner := bufio.NewScanner(resp.Body)
	var lastLine string
	for scanner.Scan() {
		lastLine = scanner.Text()
	}

	errLine := ErrorLine{}
	err = json.Unmarshal([]byte(lastLine), &errLine)
	if err != nil {
		return fmt.Errorf("docker.BuildContainer: %w", err)
	}

	if errLine.Error != "" {
		return fmt.Errorf("docker.BuildContainer: %s", errLine.Error)
	}

	return nil
}

// makeBuildContext: makes build context using tar from contextDir
func makeBuildContext(contextDir string) (*io.ReadCloser, error) {
	buildContext, err := archive.TarWithOptions(contextDir, &archive.TarOptions{})
	if err != nil {
		return nil, fmt.Errorf("docker.MakeBuildContext: tar with options: %w", err)
	}

	return &buildContext, nil
}

// gets container ids by image name `name`
func (d Dock) getIdsByName(name string) ([]string, error) {
	conts, err := d.client.ContainerList(context.TODO(), types.ContainerListOptions{
		All: true,
	})
	if err != nil {
		return []string{}, fmt.Errorf("dock.getIdsByName: container list: %w", err)
	}

	var ids []string
	for _, cont := range conts {
		if cont.Image == name {
			ids = append(ids, cont.ID)
		}
	}

	return ids, nil
}

// dock.RunContainer: removes containers with image name `name`,creates container with name `name` and config `config`
// and runs that container
//
// supresses container removing errors
func (d Dock) RunContainer(name string, config container.Config) (string, error) {
	config.Image = name

	ids, err := d.getIdsByName(name)
	if err != nil {
		d.log.Info("DEBUG", "error", err.Error())
	}

	// remove all containers with image name `name`
	if err == nil && len(ids) > 0 {

		for _, id := range ids {
			err = d.client.ContainerRemove(context.TODO(), id, types.ContainerRemoveOptions{
				Force: true,
			})
			if err != nil {
				d.log.Info("DEBUG", "error", fmt.Errorf("dock.RunContainer: remove container: %w", err))
			}
		}
	}

	resp, err := d.client.ContainerCreate(context.TODO(), &config, nil, nil, nil, name)
	if err != nil {
		return "", fmt.Errorf("dock.RunContainer: create container: %w", err)
	}

	err = d.client.ContainerStart(context.TODO(), resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return "", fmt.Errorf("dock.RunContainer: start container: %w", err)
	}

	return resp.ID, nil
}

func (d Dock) AttachContainer(id string, attachIn io.Reader, attachOut io.Writer) (*types.HijackedResponse, error) {
	waiter, err := d.client.ContainerAttach(context.TODO(), id, types.ContainerAttachOptions{
		Stream: true,
		Stderr: true,
		Stdout: true,
		Stdin:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("dock.AttachContainer: attach container: %w", err)
	}

	// attach readers/writers
	go io.Copy(attachOut, waiter.Reader)
	go io.Copy(waiter.Conn, attachIn)

	return &waiter, nil
}

func (d Dock) WaitContainer(id string) error {
	respC, errC := d.client.ContainerWait(context.TODO(), id, container.WaitConditionNextExit)
	select {
	case err := <-errC:
		if err != nil {
			return fmt.Errorf("dock.WaitContainer: %w", err)
		}
	case resp := <-respC:
		if resp.Error != nil {
			return fmt.Errorf("dock.WaitContainer: %s", resp.Error.Message)
		}
	}

	return nil
}
