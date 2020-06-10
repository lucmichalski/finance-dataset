package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	// https://godoc.org/github.com/docker/docker/client
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

const (
	initCount    = 60
	initInterval = 1 * time.Second
	restInterval = 10 * time.Second
	networkName  = "weave"
	pluginToken  = "weavemesh"
)

func main() {
	if os.Getenv("DOCKER_HOST") == "" {
		err := os.Setenv("DOCKER_HOST", "unix:///var/run/docker.sock")
		if err != nil {
			panic(err)
		}
	}
	fmt.Printf("%s Starting monitoring DOCKER_HOST = %s\n", now(), os.Getenv("DOCKER_HOST"))

	cli, err := client.NewClientWithOpts(client.FromEnv)
	for err != nil {
		fmt.Println("Failed to open Docker connection:", err)
		time.Sleep(initInterval)
		cli, err = client.NewClientWithOpts(client.FromEnv)
	}

	ctx := context.Background()
	cli.NegotiateAPIVersion(ctx)

	for i := 0; i < initCount; i++ {
		time.Sleep(initInterval)
		tick(ctx, cli)
	}

	for {
		time.Sleep(restInterval)
		tick(ctx, cli)
	}
}

func tick(ctx context.Context, cli *client.Client) {
	// fmt.Printf("%s: Checking up containers\n", now())

	options := types.ContainerListOptions{All: true, Filters: filters.NewArgs()}
	options.Filters.Add("network", networkName)
	options.Filters.Add("status", "exited")

	containers, err := cli.ContainerList(ctx, options)
	if err != nil {
		panic(err)
	}

	if len(containers) == 0 {
		return
	}

	for _, c := range containers {
		cont, err := cli.ContainerInspect(ctx, c.ID)
		if err != nil {
			continue
		}
		/* Skip healthy containers */
		if cont.State.ExitCode == 0 {
			continue
		}
		/* Monitor only {restart: always} */
		if !cont.HostConfig.RestartPolicy.IsAlways() {
			continue
		}
		/* Skip errors, not related to the WeaveMesh */
		if !strings.Contains(cont.State.Error, pluginToken) {
			continue
		}

		identifier := cont.Name
		if identifier == "" {
			identifier = cont.ID[:10]
		}

		fmt.Printf("%s: Restarting container %s\n", now(), identifier)
		err = cli.ContainerRestart(ctx, c.ID, nil)
	}
}

func now() string {
	return time.Now().Format(time.RFC3339)
}
