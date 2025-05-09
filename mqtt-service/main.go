package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	broker               = "tcp://emqx:1883"
	clientID             = "go-backend"
	username             = "your-username"
	password             = "your-password"
	commandTopicTemplate = "device/%s/command"
	statusTopic          = "device/+/status"
)

var (
	logsMu sync.Mutex
	logs   []string
)

func appendLog(line string) {
	logsMu.Lock()
	defer logsMu.Unlock()
	logs = append(logs, line)
}

func getLogs() string {
	logsMu.Lock()
	defer logsMu.Unlock()
	return fmt.Sprintln("---- LOG START ----") +
		fmt.Sprintln(strings.Join(logs, "\n"))
}

func main() {
	// 1) MQTT setup
	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID).
		SetUsername(username).
		SetPassword(password).
		SetDefaultPublishHandler(func(_ mqtt.Client, msg mqtt.Message) {
			line := fmt.Sprintf("[STATUS] %s → %s",
				msg.Topic(), string(msg.Payload()))
			appendLog(line)
		})

	client := mqtt.NewClient(opts)
	if tok := client.Connect(); tok.Wait() && tok.Error() != nil {
		log.Fatalf("MQTT connect error: %v", tok.Error())
	}
	defer client.Disconnect(250)
	appendLog("✅ Connected to MQTT broker")

	// Subscribe to status
	if tok := client.Subscribe(statusTopic, 1, nil); tok.Wait() && tok.Error() != nil {
		log.Fatalf("MQTT subscribe error: %v", tok.Error())
	}
	appendLog("✅ Subscribed to status topic")

	// 2) HTTP handlers
	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "only POST", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			DeviceID string `json:"deviceID"`
			Command  string `json:"command"`
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read error", 400)
			return
		}
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "bad json", 400)
			return
		}
		topic := fmt.Sprintf(commandTopicTemplate, req.DeviceID)
		appendLog(fmt.Sprintf("[PUBLISH] %s ← %s", topic, req.Command))
		tok := client.Publish(topic, 1, false, req.Command)
		tok.Wait()
		if err := tok.Error(); err != nil {
			appendLog(fmt.Sprintf("❌ publish error: %v", err))
		}
		w.WriteHeader(204)
	})

	http.HandleFunc("/logs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, getLogs())
	})

	// 3) Serve index.html as well
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	port := "8080"
	appendLog("🔌 HTTP server listening on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
