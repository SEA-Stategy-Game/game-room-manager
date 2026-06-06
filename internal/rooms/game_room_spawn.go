package rooms

import (
	"fmt"
	"os"
	"os/exec"
)

func SpawnGameRoom(port int, id string, max int) (int, error) {
	// format := os.Getenv("GAME_ROOM_CMD")
	args := []string{

		"--headless",
		"--",
		fmt.Sprintf("--port=%d", port),
	}
	cmd := exec.Command("./gameroom.app/Contents/MacOS/Core", args...)

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
