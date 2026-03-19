package workers

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"
)

type YoutubeDownloader struct {
	DOWNLOAD_FOLDER string
}

func (y YoutubeDownloader) Process(link string, msgs chan Message) {
	defer close(msgs)

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := exec.CommandContext(ctx, "yt-dlp",
		"-P", y.DOWNLOAD_FOLDER, // specify download folder
		"-f", "bestvideo[height<=1440]+bestaudio/best[height<=1440]", // good format
		"--newline",           // prevent flushing progress
		"--no-warnings",       // without any warnings, except errors
		"--progress-template", // indicate that we should use the provided template for display progress
		"{'progress_percentage':'%(progress._percent_str)s','progress total':'%(progress._total_bytes_str)s','speed':'%(progress._speed_str)s','ETA':'%(progress._eta_str)s'}",
		link, // actually a link on the target video
	)

	// Fetch video title in a separate goroutine
	wg.Go(func() {
		title_cmd := exec.CommandContext(ctx, "yt-dlp", "--ignore-errors", "--no-warnings", "--dump-json", link)
		out, err := title_cmd.Output()
		if err != nil {
			msgs <- error_msg(fmt.Sprintf("Error getting video info: %v\n", err))
			return
		}

		var dump map[string]any
		err = json.Unmarshal(out, &dump)
		if err != nil {
			msgs <- error_msg(fmt.Sprintf("Error parsing json: %v\n", err))
			return
		}

		// Safe type assertion to avoid panic
		if title, ok := dump["title"].(string); ok {
			if channel, ok := dump["channel"].(string); ok {
				if channelURL, ok := dump["channel_url"].(string); ok {
					msgs <- Message{Type: MessageTypeTitle, Content: fmt.Sprintf(
						"**[%s](%s) - [%s](%s)**",
						title,
						link,
						channel,
						channelURL,
					)}
				}
			}
		}
	})

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
	wg.Add(2)
	go processOutput(stdout, msgs, &wg)
	go processOutput(stderr, msgs, &wg)

	if err := cmd.Wait(); err != nil {
		msgs <- error_msg(fmt.Sprintf("Command finished with error: %v\n", err))
	}

	wg.Wait()
}

func processOutput(reader io.Reader, msgs chan Message, wg *sync.WaitGroup) {
	defer wg.Done()
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
