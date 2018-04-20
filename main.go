package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	// ErrNameNotProvided is thrown when a name is not provided
	ErrNameNotProvided = errors.New("no name was provided in the HTTP body")
)

const DEFAULT_RESPONSE string = "done"

var Version = ""

// Handler is your Lambda function handler
// It uses Amazon API Gateway request/responses provided by the aws-lambda-go/events package,
// However you could use other event sources (S3, Kinesis etc), or JSON-decoded primitive types such as 'string'.
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Println("ENV", os.Environ())

	// stdout and stderr are sent to AWS CloudWatch Logs
	log.Printf("[v0.4] Processing Lambda request %s\n", request.RequestContext.RequestID)

	log.Println("method:", request.HTTPMethod, "path:", request.Path, "res:", request.Resource)

	if request.HTTPMethod == "GET" {

		return events.APIGatewayProxyResponse{
			Body:       fmt.Sprintf("version: %v", Version),
			StatusCode: 200,
		}, nil
	}
	// If no name is provided in the HTTP request body, throw an error
	if len(request.Body) < 1 {
		return events.APIGatewayProxyResponse{}, ErrNameNotProvided
	}

	log.Printf("[v0.4] BODY: %s\n", request.Body)

	/*
		m, err := url.ParseQuery(request.Body)
		if err != nil {
			panic(err)
		}
		resp := m["response_url"][0]
		command := m["command"][0]
		text := m["text"][0]
		log.Println("response_url", resp, "command:", command, "text:", text)
	*/

	msg := awsInsatncesMsg("", true)
	return events.APIGatewayProxyResponse{
		Body:       msg,
		StatusCode: 200,
	}, nil

}

func main() {
	if os.Getenv("LAMBDA_RUNTIME_DIR") != "" {
		lambda.Start(Handler)
	} else {
		if len(os.Args) > 1 && strings.Contains(os.Args[1], "version") {
			fmt.Println("version:", Version)
		} else {
			fmt.Println(awsInsatncesMsg("", true))
		}
	}
}
