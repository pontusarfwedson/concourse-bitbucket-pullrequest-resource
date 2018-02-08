package main

import (
	"bufio"
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

	reader := bufio.NewReader(os.Stdin)
	inStr, _ := reader.ReadString(byte('\n'))
	err := logging.PrintText(fmt.Sprintf(">>>>>>>>>>    os.Stdin is: %s", inStr), whoami)
	check(err)

	err = json.Unmarshal([]byte(inStr), &request)
	check(err)
	if err != nil {
		err = logging.PrintText(fmt.Sprintf("Could not unmarshal string: %s", err.Error()), whoami)
		check(err)
	}
	err = logging.PrintText("Unmarshalled struct into", whoami)
	check(err)
	err = logging.PrintStruct(request, whoami)
	check(err)

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
				response = append(response, responseOut)
			case "INPROGRESS":
				response = append(response, responseOut)
			case "FAILING", "FAILED":
				response = append(response, responseOut)
				counter++
			case "STOPPED":
				response = append(response, responseOut)
				counter++
			case "none":
				response = append(response, responseOut)
				counter++
			default:
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
