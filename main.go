package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
)

type position struct {
	Timestamp int64   `json:"timestamp"`
	Hexid     string  `json:"hexId"`
	Ident     string  `json:"ident"`
	Squawk    int64   `json:"squawk"`
	Alt       int64   `json:"alt"`
	Speed     int64   `json:"speed"`
	AirGround string  `json:"airground"`
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Heading   int64   `json:"heading"`
}

// Format
// clock	1526120887	hexid	4CC270	ident	ICE470  	squawk	1427	alt	13950	speed	319	airGround	A	lat	51.28232	lon	-0.71182	heading	135
func convertLine(line string) string {
	parts := strings.Split(line, "\t")
	if len(parts) == 20 {
		timestamp, _ := strconv.ParseInt(parts[1], 10, 64)
		hexid := parts[3]
		ident := strings.TrimSpace(parts[5])
		squawk, _ := strconv.ParseInt(parts[7], 10, 64)
		alt, _ := strconv.ParseInt(parts[9], 10, 64)
		speed, _ := strconv.ParseInt(parts[11], 10, 64)
		airGround := parts[13]
		lat, _ := strconv.ParseFloat(parts[15], 64)
		lon, _ := strconv.ParseFloat(parts[17], 64)
		heading, _ := strconv.ParseInt(parts[19], 10, 64)

		p := position{timestamp, hexid, ident, squawk, alt, speed, airGround, lat, lon, heading}

		json, err := json.Marshal(p)
		if err != nil {
			fmt.Printf("Error: %s", err)
			return ""
		}
		return string(json)
	}
	return ""
}

var HostPtr string
var ProjectPtr string
var TopicPtr string
var KeyfilePtr string

func main() {

	flag.StringVar(&HostPtr, "host", "192.168.3.5:10001", "hostname and port running dump1090")
	flag.StringVar(&ProjectPtr, "project", "alex-olivier", "GCP Project")
	flag.StringVar(&TopicPtr, "topic", "flight-data-test", "Pub/Sub Topic Name")
	flag.StringVar(&KeyfilePtr, "keyfile", "default", "Path to keyfile")
	flag.Parse()
	println(fmt.Sprintf("Connecting to %s", HostPtr))
	println(fmt.Sprintf("Project: %s", ProjectPtr))
	println(fmt.Sprintf("Topic: %s", TopicPtr))
	println(fmt.Sprintf("Keyfile: %s", KeyfilePtr))

	// Setup Pub/Sub Connection
	ctx := context.Background()
	var pubsubClient *pubsub.Client
	var err error
	if KeyfilePtr == "default" {
		c, e := pubsub.NewClient(ctx, ProjectPtr)
		pubsubClient = c
		err = e

	} else {
		c, e := pubsub.NewClient(ctx, ProjectPtr, option.WithCredentialsFile(KeyfilePtr))
		pubsubClient = c
		err = e
	}
	if err != nil {
		log.Fatalln(err)
	}
	topic := pubsubClient.Topic(TopicPtr)

	// Setup Socket
	conn, err := net.Dial("tcp", HostPtr)
	defer conn.Close()
	if err != nil {
		log.Fatalln(err)
	}
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)

	// Do something with each line
	for {
		line, err := tp.ReadLine()
		if err != nil {
			log.Fatalln(err)
		}
		json := convertLine(line)
		if json != "" {
			_, err := topic.Publish(ctx, &pubsub.Message{Data: []byte(json)}).Get(ctx)
			if err != nil {
				log.Fatalln(err)
			}
			t := time.Now()
			fmt.Println(t.Format("2006/01/02 15:04:05"), "Published message to", TopicPtr)
		}
	}
}
