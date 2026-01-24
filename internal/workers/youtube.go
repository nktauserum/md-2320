package workers

import "fmt"

func YoutubeDownloader(link string, msg chan string) {
	msg <- fmt.Sprintf("downloading %s...\n", link)
}
