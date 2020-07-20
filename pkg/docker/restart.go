package docker

import (
	"github.com/k0kubun/pp"

	docker "github.com/fsouza/go-dockerclient"
)

var InProgress = false

func Restart(containerName string) (bool, error) {
	if InProgress == true {
		return true, nil
	}
	InProgress = true

	client, err := docker.NewClientFromEnv()
	if err != nil {
		InProgress = false
		return false, err
	}

	// containerName := "cars-protonvpn"
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
			InProgress = false
			return false, err
		}
	}

	InProgress = false
	return true, nil
}
