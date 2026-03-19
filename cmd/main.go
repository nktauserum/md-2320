package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"golang.org/x/net/proxy"

	"github.com/nktauserum/md-2320/internal/config"
	"github.com/nktauserum/md-2320/internal/handler"
	"github.com/nktauserum/md-2320/pkg/workers"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	log.Println("Starting the bot...")

	log.Println("Initiating the proxy connection...")

	socksDialer, err := proxy.SOCKS5(
		"tcp",
		cfg.SOCKS_PROXY_ADDR,
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
		cfg.TELEGRAM_TOKEN,
		telego.WithDefaultDebugLogger(),
		telego.WithHTTPClient(httpClient),
	)
	if err != nil {
		log.Fatalln(err)
	}

	handler := handler.NewHandler(cfg, map[string]workers.Worker{
		"/y": workers.YoutubeDownloader{
			DOWNLOAD_FOLDER: cfg.DOWNLOAD_FOLDER,
		},
		"/m": workers.MagnetDownloader{
			API_URL:         cfg.QBITTORRENT_API_URL,
			USERNAME:        cfg.QBITTORRENT_API_USERNAME,
			PASSWORD:        cfg.QBITTORRENT_API_PASSWORD,
			DOWNLOAD_FOLDER: cfg.DOWNLOAD_FOLDER,
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
	if err := bh.Start(); err != nil {
		log.Fatalln(err)
	}
}
