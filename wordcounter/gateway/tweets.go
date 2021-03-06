package gateway

import (
	"fmt"
	"log"
	"net/http"

	"encoding/json"

	"github.com/ReneGa/tweetcount-microservices/wordcounter/domain"
)

// Tweets is a gateway to a tweet producing service
type Tweets interface {
	Tweets(query string, offset string) domain.Tweets
}

// HTTPTweets is the gateway to get tweets over http
type HTTPTweets struct {
	Client *http.Client
	URL    string
}

type decodeResult int

const (
	decodeError decodeResult = iota
	decodeStopped
)

func decodeResponse(res *http.Response, data chan domain.Tweet, stop chan bool) decodeResult {
	defer res.Body.Close()
	var tweet domain.Tweet
	jd := json.NewDecoder(res.Body)
	for {
		select {
		case <-stop:
			return decodeStopped
		default:
			err := jd.Decode(&tweet)
			if err != nil {
				return decodeError
			}
			select {
			case data <- tweet:
			case <-stop:
				return decodeStopped
			}
		}
	}
}

// Tweets return a stream of tweets for a given search query
func (t *HTTPTweets) Tweets(query string, offset string) domain.Tweets {
	url := fmt.Sprintf("%s?q=%s&t=%s", t.URL, query, offset)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	data := make(chan domain.Tweet)
	stop := make(chan bool)

	tweets := domain.Tweets{
		Data: data,
		Stop: stop,
	}

	go func() {
		defer close(data)
		reconnect := true
		for reconnect {
			select {
			case <-stop:
				return
			default:
			}
			res, err := t.Client.Do(req)
			if err == nil {
				decodeResult := decodeResponse(res, data, stop)
				reconnect = decodeResult == decodeError
			} else {
				log.Println("Error: ", err)
			}
		}
	}()

	return tweets
}
