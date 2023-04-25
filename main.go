package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

type Game struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Price string `json:"price"`
	Image struct {
		URL string `json:"url"`
	} `json:"image"`
}

func main() {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/111.0.0.0 Safari/537.36"),
		chromedp.DisableGPU,
		chromedp.Flag("headless", true),
	)

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var gameNames, prices, urls, imageUrls []*cdp.Node
	if err := chromedp.Run(ctx,
		chromedp.Navigate("https://store.epicgames.com/en-US/free-games"),
		chromedp.Nodes(`[data-testid="direction-auto"]`, &gameNames),
		chromedp.Nodes(`[data-testid="offer-title-info-subtitle"]`, &prices),
		chromedp.Nodes(`[data-testid="offer-card-image-portrait"] img`, &imageUrls),
		chromedp.Nodes(`a[role="link"]`, &urls),
	); err != nil {
		log.Fatal(err)
	}

	var freeNowGames []Game
	for i := range gameNames {
		var game Game

		if err := chromedp.Run(ctx, chromedp.Text(gameNames[i].FullXPath(), &game.Name)); err != nil {
			log.Fatalln(err)
		}

		if err := chromedp.Run(ctx, chromedp.Text(prices[i].FullXPath(), &game.Price)); err != nil {
			log.Fatalln(err)
		}

		for _, url := range urls {
			if strings.Contains(url.AttributeValue("href"), strings.ToLower(strings.Split(game.Name, " ")[0])) {
				game.URL = fmt.Sprintf("https://store.epicgames.com%s", url.AttributeValue("href"))
				break
			}
		}

		for _, imageUrl := range imageUrls {
			if strings.Contains(imageUrl.AttributeValue("alt"), game.Name) {
				game.Image.URL = imageUrl.AttributeValue("src")
				break
			}
		}

		if strings.Contains(game.Price, "Free Now") {
			freeNowGames = append(freeNowGames, game)
		}
	}

	freeNowGamesJson, err := json.Marshal(freeNowGames)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(freeNowGamesJson))
}
