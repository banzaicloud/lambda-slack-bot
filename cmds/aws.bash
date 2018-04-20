init() {
    cmd-export-ns aws "aws namespace"
    cmd-export aws-logstreams
    cmd-export aws-logs
    cmd-export aws-listStacks
    cmd-export aws-price
    
    #AWS=$(which aws)
    deps-require aws
    AWS=.gun/bin/aws
}
aws-loggroups() {
  declare desc="lists aws loggroups"
  declare groupPattern=${1}

  $AWS logs describe-log-groups \
     --query 'logGroups[?contains(logGroupName,`'${groupPattern}'`)].logGroupName' \
     --out text
}

aws-logstreams() {
  declare desc="lists log streams belonging to a log group pattern"
  declare groupPattern=${1:? groupPattern required}

  groupName=$(aws-loggroups $groupPattern)
  $AWS logs describe-log-streams \
      --log-group-name $groupName \
      --query 'logStreams[-1].logStreamName' \
      --out text
}

aws-logs() {
  declare desc="list log messages belonging to a log group pattern"
  declare groupPattern=${1:? groupPattern required}
  declare messagePattern=${2}

  groupName=$(aws-loggroups $groupPattern)
  streamName=$(aws-logstreams $groupName)
  $AWS logs filter-log-events \
    --log-group-name $groupName \
    --log-stream-names $streamName \
    --query 'events[?contains(message,`'$messagePattern'`)].message' \
    --out text
}

aws-listStacks() {
  declare desc="List cloudformation stacks with CREATE_XXX state"

  $AWS cloudformation list-stacks \
      --query 'StackSummaries[?contains([`CREATE_COMPLETE`,`CREATE_IN_PROGRESS`,`UPDATE_COMPLETE`,`UPDATE_IN_PROGRESS`],
  StackStatus)].StackName' \
      --out table
}

aws-price() {
  declare desc="Get instance price"
  declare instanceType=$1 region=$2

  : ${instanceType:? required} ${region:? required}

  $AWS pricing get-products \
    --service-code AmazonEC2 \
    --region us-east-1 \
    --filters  \
	Type=TERM_MATCH,Field=instanceType,Value=${instanceType} \
	Type=TERM_MATCH,Field=location,Value="EU (Ireland)" \
	Type=TERM_MATCH,Field=operatingSystem,Value=Linux  \
	Type=TERM_MATCH,Field=tenancy,Value=Shared \
	Type=TERM_MATCH,Field=preInstalledSw,Value=NA \
	Type=TERM_MATCH,Field=operation,Value=RunInstances \
    --query 'PriceList[]' --out text \
     | jq  -r '.terms.OnDemand|..|.pricePerUnit?| select(.)|.USD'
}
