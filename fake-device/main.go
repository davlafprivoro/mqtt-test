package main

import (
	"fmt"
	"os"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	broker               = "tcp://emqx:1883" // or tls://... for TLS
	clientID             = "android123"
	username             = "device-username"
	password             = "device-password"
	commandTopicTemplate = "device/%s/command"
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

	device_topic := fmt.Sprintf(commandTopicTemplate, clientID)

	// Subscribe to this device's status updates
	if token := client.Subscribe(device_topic, 1, nil); token.Wait() && token.Error() != nil {
		fmt.Println("Subscription error:", token.Error())
	}

	// Keep app running to receive messages
	select {}
}
