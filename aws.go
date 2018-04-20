package main

import (
	"bufio"
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

const msgTemplate = `{ "text": "Instances",
  "attachments": [
  {{ range . }}
  {{ range .Instances }}
  { "title" : "{{- range .Tags }}
      {{- if eq ( Deref .Key ) "Name" }}
        {{- .Value}}
      {{- end}}
    {{- end}}",
    
    
    "footer": "{{ .LaunchTime }}",
    {{ if eq "running" (Deref .State.Name) }} "color":"#36a64f" ,{{end}}
    "fields": [
       {
          "title": "PublicIP",
          "value": "{{ .PublicIpAddress}}",
          "short": true
        },
       {
          "title": "Zone",
          "value": "{{ .Placement.AvailabilityZone}}",
          "short": true
        },
       {
          "title": "State",
          "value": "{{ .State.Name }}",
          "short": true
        },
       {
          "title": "InstanceId",
          "value": "{{ .InstanceId}}",
          "short": true
        }
    ]
    },
    {{end}}
    {{end}}
  ]
}`
const asciiTemplate = `{ "text": " instances: ` + "```" + `
-------------------------------------------------------------------------------------------------------------+
{{ printf "%-20s" "instanceId" }} | 
{{- printf " %-32s" "name" }} | 
{{- printf " %-16s" "publicIp" }} | 
{{- printf " %-12s" "region" }} | 
{{- printf " %-16s" "started" }} | 
-------------------------------------------------------------------------------------------------------------+
{{ range . }}
  {{- range .Instances }}
    {{- .InstanceId | Deref | printf "%-20s" }} |
    {{- range .Tags }}
      {{- if eq ( Deref .Key ) "Name" }}
        {{- .Value | Deref | printf " %-32s"}}
      {{- end}}
    {{- end}} |
   {{- .PublicIpAddress | Deref | printf " %-16s" }} |
   {{- .Placement.AvailabilityZone | Deref | printf " %-12s" }} | 
   {{- .LaunchTime | Hours  | printf "%-16s"}} | 
{{ end}}
{{- end}}-------------------------------------------------------------------------------------------------------------+` +
	"```" + `"}`

const asciiTemplateHours = `{ "text": " instances: ` + "```" + `
----------------------------------------------------------------------------------------------------------------+
{{ printf "%-20s" "instanceId" }} | 
{{- printf " %-30s" "name" }} | 
{{- printf " %-14s" "publicIp" }} | 
{{- printf " %-12s" "region" }} | 
{{- printf " %-5s" "hours" }} | 
{{- printf " %-9s" "insType" }} | 
{{- printf " %-3s" "spt" }} | 
----------------------------------------------------------------------------------------------------------------+
{{ range . }}
  {{- range .Instances }}
    {{- .InstanceId | Deref | printf "%-20s" }} |
    {{- if .Tags}}{{- range .Tags }}
      {{- if eq ( Deref .Key ) "Name" }}
        {{- .Value | Deref | printf " %-30s"}}
      {{- end}}
    {{- end}}{{- else}}{{"NONAME"| printf " %-30s"}}{{end}} |
   {{- .PublicIpAddress | Deref | printf " %-14s" }} |
   {{- .Placement.AvailabilityZone | Deref | printf " %-12s" }} |
   {{- .LaunchTime | Hours | printf "%5sh" }} | 
   {{- .InstanceType | Deref| printf " %-9s"}} |
   {{- if .InstanceLifecycle }}  X {{else}}    {{end}} |
{{ end}}
{{- end}}----------------------------------------------------------------------------------------------------------------+` +
	"```" + `"}`

const tabwriterTemplate = `
{{- printf "%s\t%s\t%s\t%s\t%s\t%s\t" "instanceId" "name" "publicIpAddress" "zone" "type" "started"}}
{{ range . }}
  {{- range .Instances }}
    {{- .InstanceId | Deref }}{{"\t"}}
    {{- . | Tag "Name" }}{{"\t"}}
    {{- .PublicIpAddress | Deref }}{{"\t"}}
    {{- .Placement.AvailabilityZone |  Deref }}{{"\t"}}
    {{- .InstanceType | Deref }}{{"\t"}}
    {{- .LaunchTime | Hours }}h{{"\t"}}
  {{- end}}
{{ end}}
`

func formatInstances(reservations []*ec2.Reservation, toAscii bool) string {
	actualTempl := msgTemplate
	if toAscii {
		//actualTempl = asciiTemplateHours
		actualTempl = tabwriterTemplate
	}
	tmpl, err := template.New("test").Funcs(template.FuncMap{
		"Deref": aws.StringValue,
		"Tag": func(keyname string, ins *ec2.Instance) string {
			for _, t := range ins.Tags {
				if aws.StringValue(t.Key) == keyname {
					return aws.StringValue(t.Value)
				}
			}
			return "n/a"
		},
		"Hours": func(t *time.Time) string {
			return strings.Split(fmt.Sprint(time.Since(*t).Truncate(time.Hour)), "h")[0]
		},
	}).Parse(actualTempl)

	if err != nil {
		panic(err)
	}

	var b bytes.Buffer
	var output bytes.Buffer
	w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', tabwriter.Debug)
	err = tmpl.Execute(w, reservations)
	if err != nil {
		panic(err)
	}
	w.Flush()

	scan := bufio.NewScanner(&b)
	scan.Scan()
	firstLine := scan.Text()
	fmt.Fprintln(&output, firstLine)
	fmt.Fprintln(&output, strings.Repeat("-", len(firstLine)))
	for scan.Scan() {
		fmt.Fprintln(&output, scan.Text())
	}

	return output.String()

}

func awsInstancesInRegion(reg string) []*ec2.Reservation {

	fmt.Println("query reg:", reg, "started ...")
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(reg),
	}))
	svc := ec2.New(sess)
	din, err := svc.DescribeInstances(&ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{
			&ec2.Filter{
				Name:   aws.String("instance-state-name"),
				Values: aws.StringSlice([]string{"running"}),
			},
		},
	})
	if err != nil {
		panic(err)
	}
	return din.Reservations
}

func awsInsatncesMsg(respUrl string, toAscii bool) string {

	_, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic(err)
	}

	if os.Getenv("regions") == "" {
		os.Setenv("regions", "eu-west-1,eu-central-1,us-east-1")
	}
	regions := strings.Split(os.Getenv("regions"), ",")
	chIns := make(chan []*ec2.Reservation, len(regions))

	for _, r := range regions {
		go func(reg string) {
			chIns <- awsInstancesInRegion(reg)
		}(r)
	}

	allInstances := make([]*ec2.Reservation, 0)
	for range regions {
		allInstances = append(allInstances, <-chIns...)
	}

	msg := formatInstances(allInstances, toAscii)

	//saveAwsResponse(allInstances)

	/* simple plain text respnse, insread of delayed answer on response_url
	slackMsg := `{ "text": " instances: ` + "```" + msg + "```" + `"}`
	resp, err := http.Post(respUrl, "application/json", bytes.NewBufferString(slackMsg))
	if err != nil {
		panic(err)
	}
	log.Println("slack resp:", resp.Status)
	*/
	return "```" + msg + "```"

}
