package workers

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func YoutubeDownloader(link string, msgs chan Message) {
	msgs <- info_msg(fmt.Sprintf("Downloading using yt-dlp %s...\n\n", link))

	cmd := exec.Command("yt-dlp", "--newline", "--progress-template", "{'progress_percentage':'%(progress._percent_str)s','progress total':'%(progress._total_bytes_str)s','speed':'%(progress._speed_str)s','ETA':'%(progress._eta_str)s'}", link)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		msgs <- error_msg(fmt.Sprintf("Error getting StdoutPipe: %v\n", err))
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		msgs <- error_msg(fmt.Sprintf("Error getting StderrPipe: %v\n", err))
		return
	}

	if err := cmd.Start(); err != nil {
		msgs <- error_msg(fmt.Sprintf("Error starting command: %v\n", err))
		return
	}

	// processing outputs
	go processOutput(stdout, msgs)
	go processOutput(stderr, msgs)

	if err := cmd.Wait(); err != nil {
		msgs <- error_msg(fmt.Sprintf("Command finished with error: %v\n", err))
	}

	close(msgs)
}

func processOutput(reader io.Reader, msgs chan Message) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text() + "\n"
		if strings.Contains(line, "progress_percentage") {
			msgs <- Message{Type: MessageTypeProgress, Content: line}
			continue
		}

		msgs <- info_msg(line)
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		msgs <- error_msg(fmt.Sprintf("Error reading: %v\n", err))
	}
}
