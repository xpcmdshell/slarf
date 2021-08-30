package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"

	"github.com/slack-go/slack"
)

func fetchAllWorkspaceUsers(client *slack.Client) ([]slack.User, error) {
	var users []slack.User
	var err error
	paginator := client.GetUsersPaginated(slack.GetUsersOptionLimit(1000))

	for {
		paginator, err = paginator.Next(context.Background())
		if paginator.Failure(err) != nil {
			// Check for rate limited error
			if rateLimitedError, ok := err.(*slack.RateLimitedError); ok {
				// Check for non-fatal case
				if rateLimitedError.Retryable() {
					// Wait the recommended time before trying again so the Slack API doesn't
					// kick us in the face
					fmt.Println("Hit rate limit")
					time.Sleep(rateLimitedError.RetryAfter)
					continue
				}
				return nil, err
			}
			return nil, err
		}
		if paginator.Done(err) {
			// If this was the last iteration, we should bail out
			break
		}

		//  Save user objects from current iteration
		users = append(users, paginator.Users...)
	}
	return users, nil
}

func populateCookieJar(dCookieStr string) (http.CookieJar, error) {
	var cookies []*http.Cookie

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	dCookie := &http.Cookie{
		Name:   "d",
		Value:  dCookieStr,
		Path:   "/",
		Domain: ".slack.com",
	}

	cookies = append(cookies, dCookie)
	slackURL, err := url.Parse("https://slack.com/")
	if err != nil {
		return nil, err
	}
	jar.SetCookies(slackURL, cookies)
	return jar, nil
}

func initializeSlackClient(token string, dCookie string) (*slack.Client, error) {
	jar, err := populateCookieJar(dCookie)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Jar: jar,
	}

	return slack.New(token, slack.OptionHTTPClient(client)), nil
}

func main() {
	var file *os.File
	var err error
	authToken := flag.String("token", "", "xoxc auth token")
	authCookie := flag.String("cookie", "", "'d' auth cookie")
	outFile := flag.String("outfile", "", "File path to save result in")
	flag.Parse()
	if *authCookie == "" || *authToken == "" {
		log.Fatalf("[-] Both auth token and auth cookie must be set")
	}
	if *outFile != "" {
		file, err = os.Create(*outFile)
		if err != nil {
			log.Fatalf("[-] Failed to create outfile: %v", err)
		}
	}

	client, err := initializeSlackClient(*authToken, *authCookie)
	if err != nil {
		log.Fatalf("[-] Failed to initialize slack client: %v", err)
	}

	users, err := fetchAllWorkspaceUsers(client)
	if err != nil {
		log.Fatalf("[-] Failed to fetch workspace users: %v", err)
	}
	if file != nil {
		encoder := json.NewEncoder(file)
		err = encoder.Encode(users)
		if err != nil {
			log.Fatalf("[-] Failed to export results as json: %v", err)
		}
	} else {
		fmt.Printf("%+v\n", users)
	}

	// All finished
	fmt.Printf("[+] Exported %d users\n", len(users))
}
