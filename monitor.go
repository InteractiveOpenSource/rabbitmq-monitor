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
	"math"
	"strings"
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
	fmt.Println(`=======================================`)
	fmt.Println(`- RabbitMQ Server Monitor (0.0.1) ===`)
	fmt.Printf("- Vhost %s\n", m.Vhost)
	fmt.Println(`=======================================`)

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
	maxName := 0
	for k := range m.Data {
		keys = append(keys, k)
		if maxName < len(k) {
			maxName = len(k)
		}
	}
	sort.Strings(keys)

	for _, name := range keys {
		q := m.Data[name]
		sign := "+"
		if q.Rates.Messages < 0 {
			sign = "-"
		} else {
			sign = "+"
		}
		line1 := fmt.Sprintf("-- %s (%s%0.2f/s) %s\n", name, sign, math.Abs(q.Rates.Messages), strings.Repeat("-", 20 + maxName - len(name)))
		fmt.Printf(line1)
		q.onScreen()
		// fmt.Printf("%s\n\n", strings.Repeat("-", len(line1)-1))
		fmt.Println("")
	}
}

func (d *queueData) onScreen() {
	// fmt.Printf("Name: %s\n", d.Name)
	fmt.Printf("Memory: %s\n", bytefmt.ByteSize(d.Memory))
	fmt.Printf("Consumers: %d\n", d.Consumers)
	fmt.Printf("Messages: %d\n", d.Messages)
	fmt.Printf("Rates: R=%0.2f/s, Pub=%0.2f/s, Ack=%0.2f/s, W=%0.2f/s\n", math.Abs(d.Rates.Ready), math.Abs(d.Rates.Publish), math.Abs(d.Rates.Ack), math.Abs(d.Rates.DiskWrites))
	// fmt.Printf("Url: %s\n", d.Url)
	// fmt.Printf("Since: %s\n", d.Since)
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

type MessagesStats struct {
	Ack               int     `json:"ack"`
	Deliver           int     `json:"deliver"`
	DeliverGet        int     `json:"deliver_get"`
	DiskWrite         int     `json:"disk_writes"`
	Publish           int     `json:"publish"`
	Redeliver         int     `json:"redeliver"`
	AckDetails        Details `json:"ack_details"`
	DeliverDetails    Details `json:"deliver_details"`
	DeliverGetDetails Details `json:"deliver_get_details"`
	DiskWriteDetails  Details `json:"disk_writes_details"`
	PublishDetails    Details `json:"publish_details"`
	RedeliverDetails  Details `json:"redeliver_details"`
}

type queueJSON struct {
	Name                          string        `json:"name"`
	Node                          string        `json:"node"`
	Messages                      int           `json:"messages"`
	MessagesReady                 int           `json:"messages_ready"`
	Since                         string        `json:"idle_since"`
	Consumers                     int           `json:"consumers"`
	Memory                        uint64        `json:"memory"`
	MessagesDetails               Details       `json:"messages_details"`
	MessagesReadyDetails          Details       `json:"messages_ready_details"`
	MessagesUnacknowledgedDetails Details       `json:"messages_unacknowledged_details"`
	MessagesStats                 MessagesStats `json:"message_stats"`
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
		// from messages_stats
		Deliver    float64 `json:"deliver"`
		DeliverGet float64 `json:"deliver_get"`
		Redeliver  float64 `json:"redeliver"`
		Publish    float64 `json:"publish"`
		Ack        float64 `json:"ack"`
		DiskWrites float64 `json:"disk_writes"`
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

	d.Rates.Messages = jsonData.MessagesDetails.Rate
	d.Rates.Ready = jsonData.MessagesReadyDetails.Rate
	d.Rates.Unacknowledged = jsonData.MessagesUnacknowledgedDetails.Rate
	d.Rates.Deliver = jsonData.MessagesStats.DeliverDetails.Rate
	d.Rates.DeliverGet = jsonData.MessagesStats.DeliverGetDetails.Rate
	d.Rates.Redeliver = jsonData.MessagesStats.RedeliverDetails.Rate
	d.Rates.Publish = jsonData.MessagesStats.PublishDetails.Rate
	d.Rates.Ack = jsonData.MessagesStats.AckDetails.Rate
	d.Rates.DiskWrites = jsonData.MessagesStats.DiskWriteDetails.Rate

	return nil
}
