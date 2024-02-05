// news aggregator server
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"go-sf-newsaggr-36a4-1/pkg/api"
	"go-sf-newsaggr-36a4-1/pkg/rss"
	"go-sf-newsaggr-36a4-1/pkg/storage"
)

// sleep routine for configured timeout
func waitAMoment(period int) {
	time.Sleep(time.Minute * time.Duration(period))
}

// async reading RSS feed function for routine. Errors and news are beeing written in separate channels
func parseURL(url string, db *storage.DB, posts chan<- []storage.Post, errs chan<- error, period int) {
	r := 0 //unexpected element type counter
	for {
		feed, err := rss.Parse(url)
		if err != nil {
			errs <- err
			if strings.HasPrefix(err.Error(), "expected element type <rss>") {
				r++
				if r > 5 { // if more then 5, then finish tries
					errs <- errors.New("finish with " + url)
					break
				}
			}
			//if we have an error we should wait for a moment
			waitAMoment(period)
			continue
		}
		posts <- feed
		//all is ok - error counter set to zero
		r = 0
		waitAMoment(period)
	}
}

// application entry point =)
func main() {

	// initialization from environment to keep safe your secrets =)
	connstr := os.Getenv("newsdb")

	_, exeFilename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("unable to get the current filename")
	}

	homePath := filepath.Dir(exeFilename)

	// --db
	db, err := storage.New(connstr)
	if err != nil {
		log.Fatal(err)
	}
	// --api
	api := api.New(db, homePath)

	// reading configuration
	cfile, err := os.ReadFile(homePath + "/config.json")
	if err != nil {
		log.Fatal(err)
	}

	//and deserializing it (according to TA)
	var config config
	err = json.Unmarshal(cfile, &config)
	if err != nil {
		log.Fatal(err)
	}

	// parsing news in separate routine for each link
	chPosts := make(chan []storage.Post)
	chErrs := make(chan error)
	for _, url := range config.URLS {
		go parseURL(url, db, chPosts, chErrs, config.Period)
	}

	// store into database
	go func() {
		for posts := range chPosts {
			err := db.StoreNews(posts)
			if err != nil {
				chErrs <- err
			}
		}
	}()

	// reading errors from routines
	go func() {
		for err := range chErrs {
			log.Println("ERROR:", err)
		}
	}()

	//for fast testing =) Opens webapp in default browser
	go func() {
		time.Sleep(time.Second * 3)
		url := fmt.Sprintf("http://localhost:%d/", config.Port)
		var cmd string
		var args []string

		switch runtime.GOOS {
		case "windows":
			cmd = "cmd"
			args = []string{"/c", "start"}
		case "darwin":
			cmd = "open"
		default: // "linux", "freebsd", "openbsd", "netbsd"
			cmd = "xdg-open"
		}
		args = append(args, url)
		err := exec.Command(cmd, args...).Start()
		if err != nil {
			log.Println(err)
		}
	}()

	// start http server
	err = http.ListenAndServe(fmt.Sprintf(":%d", config.Port), api.Router())
	if err != nil {
		log.Fatal(err)
	}
}
