package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/pontusarfwedson/concourse-bitbucket-pullrequest-resource/cmd/bitbucket"
	"github.com/pontusarfwedson/concourse-bitbucket-pullrequest-resource/cmd/logging"
	"github.com/pontusarfwedson/concourse-bitbucket-pullrequest-resource/cmd/models"
)

const (
	whoami logging.ResourceModule = logging.Check
)

func main() {
	var response models.CheckResponse
	var request models.CheckRequest

	err := json.NewDecoder(os.Stdin).Decode(&request)
	check(err)
	err = logging.PrintText("Unmarshalled struct into", whoami)

	check(err)
	err = logging.PrintStruct(request, whoami)
	check(err)

	token, err := bitbucket.RequestToken(request.Source.Key, request.Source.Secret)
	check(err)

	out, err := bitbucket.GetPullRequests(request.Source.URL, token, request.Source.APIVersion, request.Source.Team, request.Source.Repo)
	check(err)

	counter := 0
	for counter < 1 {
		for _, pr := range *out {

			state, err := bitbucket.GetCommitStatus(pr.Source.Commit.Links.Self.Href, token)
			check(err)

			link := pr.Links.HTML.Href

			if pr.CommentCount > 0 {
				comments, err := bitbucket.GetPrComments(pr.Links.Comments.Href, token)
				check(err)

				for _, comment := range comments {

					possibleCommand := strings.Split(comment.Content.Raw, "\n")[0]

					// If the first line of the comment is "/retest", then include this link
					// in the output, instead of the default PR link. This should trigger
					// a new build.
					if possibleCommand == "/retest" {
						link = comment.Link
					}
				}
			}

			responseOut := models.Version{
				Commit:      pr.Source.Commit.Hash,
				PullRequest: strconv.Itoa(pr.ID),
				Link:        link,
			}

			switch state {
			case "SUCCESSFUL":
				fmt.Println("SUCCESSFUL" + responseOut.PullRequest)
				//response = append(response, responseOut)
			case "INPROGRESS":
				fmt.Println("INPROGRESS" + responseOut.PullRequest)
				//response = append(response, responseOut)
			case "FAILING", "FAILED":
				fmt.Println("FAILED" + responseOut.PullRequest)
				//response = append(response, responseOut)
				counter++
			case "STOPPED":
				fmt.Println("STOPPED" + responseOut.PullRequest)
				//response = append(response, responseOut)
				counter++
			case "none":
				fmt.Println("none" + responseOut.PullRequest)
				response = append(response, responseOut)
				counter++
			default:
				fmt.Println("default" + responseOut.PullRequest)
				counter++
			}
		}
		break
	}

	b, _ := json.Marshal(response)
	jsonStr := string(b)
	err = logging.PrintText(fmt.Sprintf(">>>>>>>>>>     Output to os.Stdout is %s", jsonStr), whoami)
	check(err)
	fmt.Fprintf(os.Stdout, jsonStr)

}

func check(err error) {
	if err != nil {
		log.Fatalf("%+v", err)
	}
}
