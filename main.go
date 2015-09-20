package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
	"github.com/ChimeraCoder/anaconda"
	"github.com/mqu/go-notify"
)

func main() {
	consumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	consumerSecret := os.Getenv("TWITTER_CONSUMER_SECRET")
	accessToken := os.Getenv("TWITTER_ACCESS_TOKEN")
	accessTokenSecret := os.Getenv("TWITTER_ACCESS_TOKEN_SECRET")

	anaconda.SetConsumerKey(consumerKey)
	anaconda.SetConsumerSecret(consumerSecret)
	api := anaconda.NewTwitterApi(accessToken, accessTokenSecret)

	user, err := api.GetSelf(nil)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Logged in as @%s\n", user.ScreenName)
	go doNotify("Twitter", "Logged in as @" + user.ScreenName, user.ProfileImageURL)

	stream := api.UserStream(nil)

	for {
		select {
		case data := <-stream.C:

			if tweet, ok := data.(anaconda.Tweet); ok {
				fmt.Printf("@%s: %s\n", tweet.User.ScreenName, tweet.Text)
				go doNotify("@" + tweet.User.ScreenName, tweet.Text, tweet.User.ProfileImageURL)
			}
		}
	}

	fmt.Println("exiting")
}

func doNotify(title, text, image string) {
	if "" != image {
		filename := "/tmp/twitter-" + getMd5(image)

		if _, err := os.Stat(filename); err != nil {
			output, _ := os.Create(filename)
			response, _ := http.Get(image)
			io.Copy(output, response.Body)

			defer output.Close()
			defer response.Body.Close()
			defer os.Remove(filename)
		}

		image = filename
	}

	notify.Init("twitter-stream-notify")
	defer notify.UnInit()

	notification := notify.NotificationNew(title, text, image)
	notify.NotificationShow(notification)

	time.Sleep(10 * time.Second)
	notify.NotificationClose(notification)
}

func getMd5(input string) string {
	hasher := md5.New()
	hasher.Write([]byte(input))

	return hex.EncodeToString(hasher.Sum(nil))
}
