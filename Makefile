.PHONY: help
#AWS_STACK_NAME=lalyos-lambda-go
#AWS_S3_BUCKET=lp-lambda-go
VERSION = 0.0.7

help: ## Generates this help message
	@grep -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

zip: # builds the linux binary, and creates the zip for lambda upload
	GOOS=linux go build -o main  -ldflags '-X main.Version="$(VERSION)"' main.go aws.go
	aws cloudformation package --template-file template.yml --s3-bucket $(AWS_S3_BUCKET) --output-template-file packaged.yml

build-osx:
	go build -o main  -ldflags '-X main.Version="$(VERSION)"' main.go aws.go

update-fn-code: zip ## Updates the lambda code in the existing CF stack
	aws lambda update-function-code --function-name $(shell aws cloudformation list-stack-resources --stack-name $(AWS_STACK_NAME) --query 'StackResourceSummaries[?ResourceType == `AWS::Lambda::Function`].PhysicalResourceId' --out text) \
	  --s3-bucket $(AWS_S3_BUCKET) \
	  --s3-key $(shell sed -n '/CodeUri/ s:.*/::p'  packaged.yml)

deploy-stack: zip ## Deploys/Updates the cloudformation stack 
	aws cloudformation deploy \
	    --stack-name $(AWS_STACK_NAME) \
	    --template-file ./packaged.yml  \
	    --capabilities CAPABILITY_IAM

print-api-url: ## Prints ApiGateway url base
	@aws cloudformation describe-stacks \
	    --stack-name $(AWS_STACK_NAME) \
	    --query 'Stacks[0].Outputs[0].OutputValue' \
	    --out text

