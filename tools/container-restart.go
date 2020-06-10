package main

import (
	"github.com/k0kubun/pp"

	docker "github.com/fsouza/go-dockerclient"
)

func main() {
	client, err := docker.NewClientFromEnv()
	if err != nil {
		panic(err)
	}

	containerName := "cars-protonvpn"
	containers, err := client.ListContainers(docker.ListContainersOptions{
		All: true,
		Filters: map[string][]string{
			"name": []string{containerName},
		},
	})

	pp.Println("containers:", containers)
	for _, container := range containers {
		err := client.RestartContainer(container.ID, 10)
		if err != nil {
			panic(err)
		}
	}

}
