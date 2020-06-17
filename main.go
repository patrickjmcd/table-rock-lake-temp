package main

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/gocolly/colly"
	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/influxdata/influxdb-client-go/api"
)

//define a function for the default message handler
var f MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

func publishToMQTT(client MQTT.Client, name string, value string) error {
	prefix := os.Getenv("MQTT_PREFIX")
	topic := fmt.Sprintf("%s/%s", prefix, name)
	token := client.Publish(topic, 0, false, value)
	token.Wait()
	return nil
}

func randomString() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZÅÄÖ" +
		"abcdefghijklmnopqrstuvwxyzåäö" +
		"0123456789")
	length := 8
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	str := b.String() // E.g. "ExcbsVQs"
	return str
}

func getAnglerspyData(levelCh chan string, tempCh chan string) error {
	url := "https://anglerspy.com/table-rock-lake-water-temperature-ipm/"
	c := colly.NewCollector()

	c.OnHTML("#wrsn-temp-1", func(e *colly.HTMLElement) {
		tempString := strings.TrimSuffix(e.Text, "°F")
		tempCh <- tempString
	})

	c.OnHTML("#wrsn-temp-weather-1", func(e *colly.HTMLElement) {
		levelString := strings.TrimSuffix(e.Text, "′")
		levelCh <- levelString
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	c.Visit(url)

	return nil
}

func updateMQTT(temperature string) error {

	server := os.Getenv("MQTT_SERVER")
	if server == "" {
		return errors.New("no MQTT_SERVER specified")
	}

	mqttPort := os.Getenv("MQTT_PORT")
	port := "1883"
	if mqttPort != "" {
		port = mqttPort
	}

	prefix := os.Getenv("MQTT_PREFIX")
	if prefix == "" {
		return errors.New("no MQTT_PREFIX specified")
	}

	mqttURI := fmt.Sprintf("tcp://%s:%s", server, port)
	username := os.Getenv("MQTT_USERNAME")
	password := os.Getenv("MQTT_PASSWORD")

	clientID := fmt.Sprintf("lake-svc-%s", randomString())

	fmt.Printf("Trying to connect to %s\n -- clientID=%s\n", mqttURI, clientID)

	opts := MQTT.NewClientOptions().AddBroker(mqttURI)
	if username != "" {
		opts.SetUsername(username)
	}
	if password != "" {
		opts.SetPassword(password)
	}
	opts.SetClientID(clientID)
	opts.SetDefaultPublishHandler(f)

	mq := MQTT.NewClient(opts)
	if token := mq.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	publishToMQTT(mq, "temperature", temperature)
	mq.Disconnect(1000)
	return nil
}

func publishToInfluxdb(writeAPI api.WriteApiBlocking, prefix string, name string, value string) error {
	fullName := fmt.Sprintf("%s%s", prefix, name)
	units := "ºF"
	if name == "level" {
		units = "ft"
	}

	floatVal, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return err
	}

	p := influxdb2.NewPoint(fullName,
		map[string]string{"unit": units},
		map[string]interface{}{"value": value, "valueNum": floatVal},
		time.Now())
	err = writeAPI.WritePoint(context.Background(), p)
	return err
}

func updateInfluxdb(temperature string) error {
	server := os.Getenv("INFLUXDB_SERVER")
	if server == "" {
		return errors.New("no INFLUXDB_SERVER specified")
	}

	influxdbPort := os.Getenv("INFLUXDB_PORT")
	port := "8086"
	if influxdbPort != "" {
		port = influxdbPort
	}

	username := os.Getenv("INFLUXDB_USERNAME")
	password := os.Getenv("INFLUXDB_PASSWORD")

	token := ""
	if username != "" && password != "" {
		token = fmt.Sprintf("%s:%s", username, password)
	}

	prefix := os.Getenv("INFLUXDB_PREFIX")
	if prefix == "" {
		return errors.New("no INFLUXDB_PREFIX specified")
	}

	influxDatabase := os.Getenv("INFLUXDB_DATABASE")
	database := "lakeinfo"
	if prefix == "" {
		database = influxDatabase
	}

	influxdbURI := fmt.Sprintf("http://%s:%s", server, port)
	client := influxdb2.NewClient(influxdbURI, token)
	writeAPI := client.WriteApiBlocking("", database)

	err := publishToInfluxdb(writeAPI, prefix, "temperature", temperature)

	return err
}

func getLatestValues() string {

	anglerspyLevelCh := make(chan string, 1)
	anglerspyTempCh := make(chan string, 1)
	getAnglerspyData(anglerspyLevelCh, anglerspyTempCh)
	var l, t string

	for i := 0; i < 2; i++ {
		select {
		case anglerSpyLevel := <-anglerspyLevelCh:
			fmt.Printf("received anglerspy level %s ft\n", anglerSpyLevel)
			l = anglerSpyLevel
		case anglerSpyTemp := <-anglerspyTempCh:
			fmt.Printf("received anglerspy temp %s ºF\n", anglerSpyTemp)
			t = anglerSpyTemp
		}
	}
	err := updateMQTT(t)
	if err != nil {
		fmt.Printf("Couldn't send to MQTT: %s\n", err)
	} else {
		fmt.Println("Successfully wrote to MQTT")
	}
	err = updateInfluxdb(t)
	if err != nil {
		fmt.Printf("Couldn't send to InfluxDB: %s\n", err)
	} else {
		fmt.Println("Successfully wrote to InfluxDB")
	}
	return fmt.Sprintf("Level: %s ft\nTemp: %s ºF\n", l, t)
}

func main() {
	getLatestValues()
}