package main

import (
	"encoding/gob"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/service/ec2"
)

func saveAwsResponse(allInstances []*ec2.Reservation) {
	f, err := os.Create("insatnces.gob")
	if err != nil {
		panic(err)
	}

	enc := gob.NewEncoder(f)
	err = enc.Encode(allInstances)
	if err != nil {
		panic(err)
	}
}

func read() {
	f, err := os.Open("insatnces.gob")
	if err != nil {
		panic(err)
	}
	dec := gob.NewDecoder(f)
	var reservations []*ec2.Reservation
	dec.Decode(&reservations)

	/*
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
		}).Parse(tabwriterTemplate)
		if err != nil {
			panic(err)
		}

		var b bytes.Buffer
		w := tabwriter.NewWriter(&b, 0, 0, 1, ' ', tabwriter.Debug)
		tmpl.Execute(w, reservations)
		w.Flush()

		scan := bufio.NewScanner(&b)
		scan.Scan()
		firstLine := scan.Text()
		fmt.Println(firstLine)
		fmt.Println(strings.Repeat("-", len(firstLine)))
		for scan.Scan() {
			fmt.Println(scan.Text())
		}
	*/

	fmt.Printf("FORMATTED:\n%s\n", formatInstances(reservations, true))
}
