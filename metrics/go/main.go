package main

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

func main() {
	iowait, cpuuse, _ := GetCPUInfo()
	conns, _ := GetNumberHTTPConnections()
	SendMetrics(iowait, cpuuse, conns)
}

func GetCPUInfo() (float64, float64, error) {
	output, err := exec.Command("mpstat").CombinedOutput()
	if err != nil {
		log.Fatal("Error executing mpstat command to get cpu utilization", err)
		return -1, -1, err
	}

	lines := strings.Split(string(output), "\n")
	fields := strings.Fields(lines[3])
	iowait := fields[5]
	idle := fields[11]

	iowaitF, err := strconv.ParseFloat(strings.Replace(iowait, ",", ".", 1), 64)
	if err != nil {
		log.Fatal("Error parsing numeric value", err)
		return -1, -1, err
	}
	idleF, err := strconv.ParseFloat(strings.Replace(idle, ",", ".", 1), 64)
	if err != nil {
		log.Fatal("Error parsing numeric value", err)
		return -1, -1, err
	}

	return iowaitF, 100.00 - idleF, nil
}

func GetNumberHTTPConnections() (float64, error) {
	// Only TCP IP4 connections
	output, err := exec.Command("netstat", "-ant").CombinedOutput()
	if err != nil {
		log.Fatal("Error executing netstat command to get connections", err)
		return -1, err
	}

	lines := strings.Split(string(output), "\n")
	lines = lines[2 : len(lines)-1]
	var counter float64
	for _, line := range lines {
		address := strings.Fields(line)[3]
		port := strings.Split(address, ":")[1]
		if port == "80" {
			counter++
		}
	}

	return counter, nil
}

func SendMetrics(iowait float64, cpuuse float64, connections float64) {
	region, instanceId, _ := GetInstanceMetadata()
	sess := session.Must(session.NewSession())
	client := cloudwatch.New(sess, &aws.Config{
		Region: aws.String(region),
	})

	client.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace: aws.String("EC2/CPUConnections"),
		MetricData: []*cloudwatch.MetricDatum{
			&cloudwatch.MetricDatum{
				MetricName: aws.String("IOWait"),
				Unit:       aws.String("Percent"),
				Value:      aws.Float64(iowait),
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  aws.String("InstanceId"),
						Value: aws.String(instanceId),
					},
				},
			},
			&cloudwatch.MetricDatum{
				MetricName: aws.String("CPUUtilization"),
				Unit:       aws.String("Percent"),
				Value:      aws.Float64(cpuuse),
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  aws.String("InstanceId"),
						Value: aws.String(instanceId),
					},
				},
			},
			&cloudwatch.MetricDatum{
				MetricName: aws.String("TCP80Connections"),
				Unit:       aws.String("Count"),
				Value:      aws.Float64(connections),
				Dimensions: []*cloudwatch.Dimension{
					&cloudwatch.Dimension{
						Name:  aws.String("InstanceId"),
						Value: aws.String(instanceId),
					},
				},
			},
		},
	})
}

func GetInstanceMetadata() (string, string, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", "http://169.254.169.254/latest/dynamic/instance-identity/document", nil)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("No metadata could be retrieved", err)
		return "", "", err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Metadata could not be parsed", err)
		return "", "", err
	}

	var result map[string]string
	json.Unmarshal(data, &result)

	region := result["region"]
	instanceId := result["instanceId"]

	return region, instanceId, nil
}
