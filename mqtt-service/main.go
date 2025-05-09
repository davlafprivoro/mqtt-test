package main

import (
	"fmt"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	broker               = "tcp://emqx:1883" // or tls://... for TLS
	clientID             = "go-backend"
	username             = "your-username"
	password             = "your-password"
	commandTopicTemplate = "device/%s/command"
	statusTopic          = "device/+/status" // subscribe to all devices
)

func main() {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetDefaultPublishHandler(func(client mqtt.Client, msg mqtt.Message) {
		fmt.Printf("Received message: [%s] %s\n", msg.Topic(), string(msg.Payload()))
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("Error connecting:", token.Error())
		os.Exit(1)
	}
	defer client.Disconnect(250)

	// Subscribe to device status updates
	if token := client.Subscribe(statusTopic, 1, nil); token.Wait() && token.Error() != nil {
		fmt.Println("Subscription error:", token.Error())
	}

	// Send command to device with ID "android123"
	deviceID := "android123"
	commandTopic := fmt.Sprintf(commandTopicTemplate, deviceID)
	command := "TURN_ON"

	fmt.Printf("Sending command to %s: %s\n", commandTopic, command)
	token := client.Publish(commandTopic, 1, false, command)
	token.Wait()

	// Keep app running to receive messages
	select {}
}
