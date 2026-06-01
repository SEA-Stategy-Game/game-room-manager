package rooms

import (
	"fmt"
	"os"
	"os/exec"
)

func SpawnGameRoom(port int) (int, error) {
	format := os.Getenv("GAME_ROOM_CMD")
	cmd := exec.Command(format, fmt.Sprintf("%d", port))

	cmd.Start()

	if cmd.Stderr != nil {
		return 0, &exec.Error{}
	}

	return cmd.Process.Pid, nil
}
