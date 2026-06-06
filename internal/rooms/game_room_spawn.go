package rooms

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"runtime"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

func SpawnGameRoom(port int, id string, max int, image string) (int, error) {
	if os.Getenv("RUNNING_IN_DOCKER") == "true" {
		return runDocker(port, id, max, image)
	}
	return runLocally(port, id, max)
}

func runDocker(port int, id string, max int, image string) (int, error) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return 0, fmt.Errorf("failed to create docker client: %w", err)
	}

	containerPort := nat.Port("12345/tcp")
	hostPort := fmt.Sprintf("%d", port)

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: image,
		Env: []string{
			"USE_REDIS=true",
			"REDIS_HOST=host.docker.internal",
			"REDIS_PORT=6379",
			fmt.Sprintf("MAX_PLAYERS=%d", max),
			"GAME_ROOM_MANAGER_URL=http://host.docker.internal:8080",
			"PLANNING_URL=http://host.docker.internal:5000",
			fmt.Sprintf("GAME_ROOM_ID=%s", id),
		},
		ExposedPorts: nat.PortSet{
			containerPort: struct{}{},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			containerPort: []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: hostPort,
				},
			},
		},
	}, &network.NetworkingConfig{}, nil, "")

	if err != nil {
		return 0, fmt.Errorf("failed to create container: %w", err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return 0, fmt.Errorf("failed to start container: %w", err)
	}

	inspect, err := cli.ContainerInspect(ctx, resp.ID)
	if err == nil {
		return inspect.State.Pid, nil
	}
	return 0, nil
}

// runLocally only runs in macbook for now
func runLocally(port int, id string, max int) (int, error) {
	var binary string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		binary = "./gameroom.app/Contents/MacOS/Core"
		args = []string{
			"--headless",
			"--",
			"--port",
			fmt.Sprintf("%d", port),
		}
	case "linux":
		binary = "./my_server.x86_64"
		args = []string{
			"--headless",
			fmt.Sprintf("--port=%d", port),
		}
	default:
		return 0, fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
	cmd := exec.Command(binary, args...)

	cmd.Env = append(os.Environ(),
		"USE_REDIS=true",
		fmt.Sprintf("GAME_ROOM_ID=%s", id),
		fmt.Sprintf("MAX_PLAYERS=%d", max),
	)

	err := cmd.Start()
	if err != nil {
		return 0, fmt.Errorf("failed to start game room: %w", err)
	}

	if cmd.Stderr != nil {
		return 0, &exec.Error{}
	}

	if cmd.Process == nil {
		return 0, fmt.Errorf("failed to get process pid")
	}

	return cmd.Process.Pid, nil
}
