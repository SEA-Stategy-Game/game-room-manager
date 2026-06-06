package rooms

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func SpawnGameRoom(port int, id string, max int) (int, error) {

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

	cmd.Start()

	if cmd.Stderr != nil {
		return 0, &exec.Error{}
	}

	return cmd.Process.Pid, nil
}
