package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var queue []string
var _pollAndDrainSeconds time.Duration
var _maxItemsInQueueBeforeDrain = 10
var _logseneElasticsearchURL = ""
var _logseneToken = ""
var _logType = ""
var _debugMode = false

func sendBulk() (string, time.Duration) {
	baseURL := _logseneElasticsearchURL
	qurl := baseURL + "_bulk"

	var desc = ELKActionDescription{Index: _logseneToken, Type: _logType}
	var action = ELKAction{Index: desc}

	bAction, errA := json.Marshal(action)
	if errA != nil {
		log.Println(errA)
	}
	//Performance gain: just use the byte[] instead of converting to string multiple times
	payload := string(bAction)
	for i := 0; i < len(queue); i++ {
		var msg LogMessage
		msg.Message.Text = queue[i] + string(i)
		b, err := json.Marshal(msg)
		if err != nil {
			log.Println(err)
			continue
		} else {
			payload += "\n" + string(b)
		}
	}

	req, _ := http.NewRequest("POST", qurl, bytes.NewBufferString(payload))
	body, elapsed := makeRequest(req)
	if _debugMode {
		log.Println(string(body))
	}
	return string(body), elapsed
}

func makeRequest(request *http.Request) ([]byte, time.Duration) {
	start := time.Now()
	res, _ := http.DefaultClient.Do(request)
	elapsed := time.Since(start)
	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	return body, elapsed
}

func drainQueue() {
	log.Println("Draining queue with length", len(queue))
	if len(queue) > 0 {
		sendBulk()
		queue = queue[:0]
		log.Println("Drain done")
	}
}

func startPolling(d time.Duration) {
	for {
		time.Sleep(d)
		go drainQueue()
	}
}

func main() {
	fi, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}
	if fi.Mode()&os.ModeNamedPipe == 0 {
		fmt.Println("must be used with pipe :(")
		return
	}

	logseneReceiverURL := flag.String("logseneurl", "https://logsene-receiver.eu.sematext.com/", "The elasticsearch host to send the logs to. Default: https://logsene-receiver.eu.sematext.com/")
	logseneToken := flag.String("logsenetoken", "", "<your logsene token>")
	timebeforedrain := flag.Int("timebeforedrain", 10000, "Milliseconds between drains. If 0, no drain will be set. Default: 10000ms")
	maxitems := flag.Int("maxitems", 100, "The amount of items that can be in queue before the queue is forced to drain. Default: 100")
	logType := flag.String("logtype", "logstreamer", "Tag logs with a certain logtype. Default: logstreamer")
	debugMode := flag.Bool("debug", false, "Debug mode enables more logging. Default: false")
	flag.Parse()

	_logseneElasticsearchURL = *logseneReceiverURL
	_pollAndDrainSeconds = time.Duration(*timebeforedrain) * time.Millisecond
	_logseneToken = *logseneToken
	_maxItemsInQueueBeforeDrain = *maxitems
	_logType = *logType
	_debugMode = *debugMode

	log.Println("Watch started")
	if _pollAndDrainSeconds > 0 {
		go startPolling(_pollAndDrainSeconds)
	}

	r := bufio.NewReader(os.Stdin)
	scanner := bufio.NewScanner(r)
	entriesSinceDrained := 0
	forceDrain := _maxItemsInQueueBeforeDrain > 0
	var idx uint64 //just to keep track

	for scanner.Scan() {
		line := scanner.Text()
		queue = append(queue, line)
		if _debugMode {
			log.Println(idx, ":", line)
			idx++
		}
		entriesSinceDrained++
		if forceDrain && entriesSinceDrained > _maxItemsInQueueBeforeDrain {
			log.Println("Drain limit reached - draining..")
			entriesSinceDrained = 0
		}
	}
}
