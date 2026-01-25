package workers

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

type MagnetDownloader struct {
	API_URL, USERNAME, PASSWORD string
}

func (m MagnetDownloader) authorize() (string, error) {
	var sid_cookie string

	req, err := http.NewRequest("POST", m.API_URL+"/api/v2/auth/login", strings.NewReader(fmt.Sprintf("username=%s&password=%s", m.USERNAME, m.PASSWORD)))
	if err != nil {
		return sid_cookie, err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")

	client := http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return sid_cookie, err
	}

	sid_cookie = response.Cookies()[0].Value

	return sid_cookie, nil
}

func (m MagnetDownloader) Process(magnet_link string, msgs chan Message) {
	sid, err := m.authorize()
	if err != nil {
		msgs <- error_msg("error auth: " + err.Error())
		return
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	writer.WriteField("urls", magnet_link)
	writer.WriteField("autoTMM", "false")
	writer.WriteField("savepath", "/downloads")
	writer.WriteField("rename", "")
	writer.WriteField("category", "")
	writer.WriteField("stopped", "false")
	writer.WriteField("stopCondition", "None")
	writer.WriteField("contentLayout", "Original")
	writer.WriteField("dlLimit", "0")
	writer.WriteField("upLimit", "0")

	if err := writer.Close(); err != nil {
		msgs <- error_msg("error closing multipart writer: " + err.Error())
		return
	}

	req, err := http.NewRequest("POST", m.API_URL+"/api/v2/torrents/add", body)
	if err != nil {
		msgs <- error_msg("error doing request: " + err.Error())
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: "SID", Value: sid})

	client := http.Client{}
	response, err := client.Do(req)
	if err != nil {
		msgs <- error_msg("error send request: " + err.Error())
		return
	}

	text, err := io.ReadAll(response.Body)
	if err != nil {
		msgs <- error_msg("error read response: " + err.Error())
		return
	}
	defer response.Body.Close()

	msgs <- info_msg("status: " + string(text))

	close(msgs)
}
