package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
	"net"
	"net/textproto"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

var syncContainers sync.Map
var hostname = os.Getenv("HOSTNAME")
var chanLogLines chan LogLine
var chanInfoCont chan Container

func main() {
	logSetup()
	chanLogLines = make(chan LogLine)
	chanInfoCont = make(chan Container)

	go func() {
		for {
			for _, val := range getContainers() {
				if val.State != "running" {
					continue
				}
				_, loaded := syncContainers.LoadOrStore(val.ID, "following")
				if !loaded {
					log.Info("Follow ", val.Names)

					val.Hostname = hostname
					chanInfoCont <- val
					go runObserver(val)
				}
			}

			// time update info about containers
			time.Sleep(time.Second * 5)
		}
	}()

	// send all working containers
	go func() {
		for {
			time.Sleep(time.Second * 5)
			for _, val := range getContainers() {
				val.Hostname = hostname
				chanInfoCont <- val
			}
		}
	}()

	go func() {
		openChRab, err := connectRabbit().Channel()
		defer openChRab.Close()
		if err != nil {
			log.Fatal(err)
		}

		for val := range chanInfoCont {
			jsonLine, err := json.Marshal(val)
			if err != nil {
				log.Error(err)
			}

			if err := sendToRabbit(openChRab, "containers", string(jsonLine)); err != nil {
				log.Error("sleep 10 sec / sendToRabbit / " + err.Error())
				time.Sleep(10 * time.Second)
			}
		}
	}()

	openChRab, err := connectRabbit().Channel()
	defer openChRab.Close()
	if err != nil {
		log.Fatal(err)
	}
	for val := range chanLogLines {
		jsonLine, err := json.Marshal(val)
		if err != nil {
			log.Error(err)
		}

		if sendToRabbit(openChRab, "logs", string(jsonLine)) != nil {
			log.Error("sleep 10 sec / sendToRabbit / " + err.Error())
			time.Sleep(10 * time.Second)
		}
	}
}

func connectDocker() net.Conn {
	conn, err := net.Dial("unix", "/var/run/docker.sock")

	if err != nil {
		log.Fatal("Not connect to Docker / " + err.Error())
	}

	return conn
}

func getContainers() []Container {
	conn := connectDocker()
	defer conn.Close()
	tp := request(conn, "/containers/json?all=1")

	out, _ := tp.ReadDotBytes()

	var containers []Container
	err := json.Unmarshal([]byte(strings.Split(string(out), "\n\n")[1]), &containers)
	if err != nil {
		log.Fatal(err)
	}

	return containers
}

func request(conn net.Conn, url string) *textproto.Reader {
	_, err := conn.Write([]byte("GET /v1.41" + url + " HTTP/1.0\r\n\n"))
	if err != nil {
		log.Fatal(err)
	}

	return textproto.NewReader(bufio.NewReader(conn))
}

func runObserver(container Container) {
	conn := connectDocker()
	tp := request(conn,
		fmt.Sprintf("/containers/"+container.ID+"/logs"+
			"?stdout=true&stderr=true&follow=true&since=%d",
			time.Now().Unix()-10))

	toggle := false
	for {
		line, err := tp.ReadLine()

		if err != nil {
			log.Info(err)
			if _, loaded := syncContainers.LoadAndDelete(container.ID); loaded {
				log.Info("Unfollow ", container.Names)
			}
			conn.Close()
			return
		}

		var lineOut string
		if toggle && len(line) >= 1 {
			lineOut = line[8:]
		} else {
			lineOut = line
		}

		chanLogLines <- LogLine{container.ID, hostname, lineOut}

		if !toggle && strings.Contains(lineOut, "Server: ") {
			toggle = true
		}
	}
}

func connectRabbit() *amqp.Connection {
	cfg := new(tls.Config)

	cfg.RootCAs = x509.NewCertPool()

	ca, err := os.ReadFile(os.Getenv("CA_CERT"))
	if err != nil {
		log.Fatal(err)
	}
	cfg.RootCAs.AppendCertsFromPEM(ca)

	cert, err := tls.LoadX509KeyPair(os.Getenv("CLIENT_CERT"),
		os.Getenv("CLIENT_KEY"))
	cfg.Certificates = append(cfg.Certificates, cert)
	if err != nil {
		log.Fatalln(err)
	}

	conn, err := amqp.DialTLS("amqps://"+os.Getenv("AMQP_URL")+"/", cfg)
	if err != nil {
		log.Fatalln(err.Error())
	}

	return conn
}

func sendToRabbit(ch *amqp.Channel, channel string, body string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	q, err := ch.QueueDeclare(
		channel,
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		log.Error(err)
		return err
	}

	err = ch.PublishWithContext(ctx,
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	if err != nil {
		return err
	}

	return nil
}

func logSetup() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf(" %s:%d", filename, f.Line)
		},
	})
	if l, err := log.ParseLevel("debug"); err == nil {
		log.SetLevel(l)
		log.SetReportCaller(l == log.DebugLevel)
		log.SetOutput(os.Stdout)
	}
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}
