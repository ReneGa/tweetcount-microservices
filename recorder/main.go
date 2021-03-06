package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/ReneGa/tweetcount-microservices/generic"
	"github.com/ReneGa/tweetcount-microservices/recorder/datamapper"
	"github.com/ReneGa/tweetcount-microservices/recorder/gateway"
	"github.com/ReneGa/tweetcount-microservices/recorder/resource"
	"github.com/ReneGa/tweetcount-microservices/recorder/service"
	"github.com/julienschmidt/httprouter"
)

var address = flag.String("address", "localhost:8085", "Address to listen on")
var bucketsDirectory = flag.String("bucketsDirectory", "./buckets", "Directory to write tweet buckets to.")
var bucketDuration = flag.Duration("bucketDuration", time.Hour, "Duration of a tweet bucket.")

var tweetsURL = flag.String("tweetsURL", "http://localhost:8080/tweets", "URL of the tweet producer to connect to")

func main() {
	flag.Parse()

	queriesDataMapper := &datamapper.JSONFileTweetBucketsPerQuery{
		IOUtil:    &generic.RealIOUtil{},
		OS:        &generic.RealOS{},
		FileNamer: datamapper.RFC3339BucketFileNamer{},
		FileMode:  0777,
		Directory: *bucketsDirectory,
		Duration:  *bucketDuration,
	}

	tweetsGateway := &gateway.HTTPTweets{
		Client: http.DefaultClient,
		URL:    *tweetsURL,
	}
	tweetsService := &service.Tweets{
		DataMapper: queriesDataMapper,
		Gateway:    tweetsGateway,
	}
	tweetsResource := resource.Tweets{
		Service: tweetsService,
	}

	router := httprouter.New()
	router.GET("/tweets", tweetsResource.GET)

	done := make(chan bool)
	go func() {
		err := http.ListenAndServe(*address, router)
		if err != nil {
			log.Fatal(err)
		}
		done <- true
	}()
	log.Printf("listening on %s", *address)
	<-done
}
