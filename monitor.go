package main

import (
	"os"
	"fmt"
	"time"
	"sort"
	"os/exec"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"code.cloudfoundry.org/bytefmt"
)

func Monitor(env ServerConfig) *MonitorScreen {
	b := map[string]queueData{}
	monitor := &MonitorScreen{env, b}

	return monitor
}

type MonitorScreen struct {
	ServerConfig
	Data map[string]queueData
}

func (m *MonitorScreen) Show() (succeeded bool) {
	fmt.Println(">> Date", time.Now())

	url := fmt.Sprintf("http://%s:%s@%s:%d/api/queues/%s", m.User, m.Password, m.Host, m.Port, m.Vhost)
	if response, err := http.Get(url); err != nil {
		fmt.Println("HTTP Error", err, url)
	} else {
		if response.StatusCode != 200 {
			fmt.Println("HTTP Error", response.Status, url)
			os.Exit(1)

			return
		} else {
			// got correct response
			var jsonData []queueJSON
			buf, _ := ioutil.ReadAll(response.Body)
			if err := json.Unmarshal(buf, &jsonData); err != nil {
				fmt.Println("Incorrect response", err)
				os.Exit(1)

				return
			}

			// build data for each queue found in jsonData
			channel := make(chan queueData)
			for _, q := range jsonData {
				go func(q queueJSON) {
					d := queueData{}
					d.Url = fmt.Sprintf("%s/%s", url, q.Name)

					d.build(q)

					channel <- d
				}(q)
			}

			for _, _ = range jsonData {
				data := <-channel
				m.Data[data.Name] = data
			}

			m.OnScreen()
		}
	}

	return
}

func (m *MonitorScreen) OnScreen() {
	var keys []string
	for k := range m.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, name := range keys {
		q := m.Data[name]
		line1 := fmt.Sprintf("-- %s ---------------\n", name)
		fmt.Printf(line1)
		q.onScreen()
		// fmt.Printf("%s\n\n", strings.Repeat("-", len(line1)-1))
		fmt.Println("")
	}
}

func (m *MonitorScreen) Clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func (m *MonitorScreen) Tick(tickLag int) {
	forever := make(chan struct{})
	for {
		func() {
			m.Clear()
			m.Show()
			time.Sleep(time.Duration(tickLag) * time.Millisecond)
		}()
	}
	<-forever
}

type Details struct {
	Avg     float64 `json:"avg"`
	AvgRate float64 `json:"avg_rate"`
	Rate    float64 `json:"rate"`
}

type queueJSON struct {
	Name                          string  `json:"name"`
	Node                          string  `json:"node"`
	Messages                      int     `json:"messages"`
	MessagesReady                 int     `json:"messages_ready"`
	Since                         string  `json:"idle_since"`
	Consumers                     int     `json:"consumers"`
	Memory                        uint64  `json:"memory"`
	MessagesDetails               Details `json:"messages_details"`
	MessagesReadyDetails          Details `json:"messages_ready_details"`
	MessagesUnacknowledgedDetails Details `json:"messages_unacknowledged_details"`
}

type queueData struct {
	Url       string `json:"url"`
	Name      string `json:"name"`
	Messages  int    `json:"messages"`
	Since     string `json:"since"`
	Consumers int    `json:"consumers"`
	Memory    uint64 `json:"memory"`
	Rates struct {
		Messages       float64 `json:"messages"`
		Ready          float64 `json:"ready"`
		Unacknowledged float64 `json:"unack"`
	} `json:"rates"`
}

func (d *queueData) build(q queueJSON) error {
	d.Name = q.Name
	d.Messages = q.Messages

	// get data from d.Url
	response, err := http.Get(d.Url)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("http Error: status=%v", response)
	}

	var jsonData queueJSON
	buf, _ := ioutil.ReadAll(response.Body)
	if err := json.Unmarshal(buf, &jsonData); err != nil {
		return err
	}

	d.Since = jsonData.Since
	d.Consumers = jsonData.Consumers
	d.Memory = jsonData.Memory

	d.Rates.Messages, d.Rates.Ready, d.Rates.Unacknowledged = jsonData.MessagesDetails.Rate, jsonData.MessagesReadyDetails.Rate, jsonData.MessagesUnacknowledgedDetails.Rate

	return nil
}

func (d *queueData) onScreen() {
	// fmt.Printf("Name: %s\n", d.Name)
	fmt.Printf("Memory: %s\n", bytefmt.ByteSize(d.Memory))
	fmt.Printf("Consumers: %d\n", d.Consumers)
	fmt.Printf("Messages: %d\n", d.Messages)
	fmt.Printf("Rates: M=%s/s, U=%s/s, R=%s/s\n", bytefmt.ByteSize(uint64(d.Rates.Messages)), bytefmt.ByteSize(uint64(d.Rates.Unacknowledged)), bytefmt.ByteSize(uint64(d.Rates.Ready)))
	// fmt.Printf("Url: %s\n", d.Url)
	// fmt.Printf("Since: %s\n", d.Since)
}
