package workers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
)

type YoutubeDownloader struct{}

func (y YoutubeDownloader) Process(link string, msgs chan Message) {
	cmd := exec.Command("yt-dlp",
		"-f", "bestvideo[height<=1440]+bestaudio/best[height<=1440]", // good format
		"--newline",           // prevent flushing progress
		"--no-warnings",       // without any warnings, except errors
		"--progress-template", // indicate that we should use the provided template for display progress
		"{'progress_percentage':'%(progress._percent_str)s','progress total':'%(progress._total_bytes_str)s','speed':'%(progress._speed_str)s','ETA':'%(progress._eta_str)s'}",
		link, // actually a link on the target video
	)

	go func() {
		title_cmd := exec.Command("yt-dlp", "--ignore-errors", "--no-warnings", "--dump-json", link)
		out, err := title_cmd.Output()
		if err != nil {
			msgs <- error_msg(fmt.Sprintf("Error starting command: %v\n", err))
		}

		var dump map[string]any
		json.Unmarshal(out, &dump)
		if err != nil {
			msgs <- error_msg(fmt.Sprintf("Error parsing json: %v\n", err))
		}

		if _, ok := <-msgs; !ok {
			return
		}

		msgs <- Message{Type: MessageTypeTitle, Content: fmt.Sprintf(
			"**[%s](%s) - [%s](%s)**",
			dump["title"].(string),
			link,
			dump["channel"].(string),
			dump["channel_url"].(string),
		)}
	}()

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
		} else if strings.Contains(line, "ERROR") {
			msgs <- Message{Type: MessageTypeError, Content: line}
		} else if strings.Contains(line, "has already been downloaded") {
			msgs <- Message{Type: MessageTypeAlreadyExists, Content: line}
		}
		log.Println(line)
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		msgs <- error_msg(fmt.Sprintf("Error reading: %v\n", err))
	}
}
