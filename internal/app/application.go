package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"golang.org/x/net/proxy"

	"github.com/nktauserum/md-2320/internal/config"
	"github.com/nktauserum/md-2320/pkg/workers"
)

type Application struct {
	c *config.Config
}

func NewApplication(config *config.Config) *Application {
	return &Application{c: config}
}

func (a *Application) Run() error {
	ctx := context.Background()
	log.Println("Starting the bot...")

	log.Println("Initiating the proxy connection...")

	socksDialer, err := proxy.SOCKS5(
		"tcp",
		a.c.SOCKS_PROXY_ADDR,
		nil,
		&net.Dialer{Timeout: 10 * time.Second},
	)

	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return socksDialer.Dial(network, addr)
	}

	transport := &http.Transport{
		DialContext:           dialContext,
		DisableKeepAlives:     false,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
	}

	httpClient := &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
	}

	bot, err := telego.NewBot(
		a.c.TELEGRAM_TOKEN,
		telego.WithDefaultDebugLogger(),
		telego.WithHTTPClient(httpClient),
	)
	if err != nil {
		fmt.Println(err)
		return err
	}

	handler := NewHandler(a.c, map[string]workers.Worker{
		"/y": workers.YoutubeDownloader{
			DOWNLOAD_FOLDER: a.c.DOWNLOAD_FOLDER,
		},
		"/m": workers.MagnetDownloader{
			API_URL:         a.c.QBITTORRENT_API_URL,
			USERNAME:        a.c.QBITTORRENT_API_USERNAME,
			PASSWORD:        a.c.QBITTORRENT_API_PASSWORD,
			DOWNLOAD_FOLDER: a.c.DOWNLOAD_FOLDER,
		},
	})

	updates, _ := bot.UpdatesViaLongPolling(ctx, nil)
	bh, _ := th.NewBotHandler(bot, updates)
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ChatID: update.Message.Chat.ChatID(), Text: "*..0111010110..*\n\nTHE SPIRIT OF THE MACHINE IS WORKING. HAIL OMNISSIAH!", ParseMode: telego.ModeMarkdown})
		return nil
	}, th.CommandEqual("start"))
	bh.HandleMessage(handler.HandleRequest)

	defer func() { _ = bh.Stop() }()
	return bh.Start()
}
