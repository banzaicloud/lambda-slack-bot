## lambda-go-samples

An example of using AWS Lambda with Go

## Compile

```
GOOS=linux go build -o main
```

## Deployment

```
export AWS_STACK_NAME=lalyos-lambda-go
export AWS_S3_BUCKET=lp-lambda-go

aws cloudformation package --template-file template.yml --s3-bucket $AWS_S3_BUCKET --output-template-file packaged.yml
aws cloudformation deploy --template-file ./packaged.yml --stack-name $AWS_STACK_NAME --capabilities CAPABILITY_IAM
```

## Invoke API

```
AWS_API_ID=$(aws apigateway get-rest-apis --query 'items[? name == `'$AWS_STACK_NAME'`].id' --out text)
AWS_STAGE=$(aws apigateway get-stages --rest-api-id $AWS_API_ID --query 'item[0].stageName' --out text)

curl https://$AWS_API_ID.execute-api.$AWS_DEFAULT_REGION.amazonaws.com/$AWS_STAGE/ -d geza
```

## List stacks

```
aws cloudformation list-stacks \
  --stack-status-filter CREATE_COMPLETE  \
  --query StackSummaries[].StackName \
  --out table
```

list lambda fn in stack
```
aws cloudformation list-stack-resources \
  --stack-name $AWS_STACK_NAME \
  --query 'StackResourceSummaries[?ResourceType == `AWS::Lambda::Function`].PhysicalResourceId' \
  --out text
```

list all resorces:
```
aws cloudformation list-stack-resources \
  --stack-name $AWS_STACK_NAME \
  --query 'StackResourceSummaries[].[ResourceType,PhysicalResourceId,LogicalResourceId]' \
  --out table
```

list regions:
```
aws ec2 describe-regions \
  --query 'join(`,`, Regions[].RegionName | sort(@))' \
  --out text
```

exclude one:
```
aws ec2 describe-regions \
  --query 'join(`,`, Regions[? RegionName != `ap-northeast-2`].RegionName | sort(@))' \
  --out text
```
