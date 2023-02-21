package main

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	b64 "encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

var wsClient sync.Map
var wsMu sync.Mutex
var containersSub sync.Map
var containers sync.Map
var startTime time.Time
var password string
var tokenInstall string

var upgrader = websocket.Upgrader{} // use default options

func main() {
	startTime = time.Now()
	password = os.Getenv("PASSWORD_DASHBOARD")

	logSetup()
	genToken()
	loadingContainers()

	go func() {
		log.Info("start dashboard")
		http.HandleFunc("/ws", ws)
		http.HandleFunc("/install-momo.sh", func(w http.ResponseWriter, r *http.Request) {
			if r.FormValue("token") == tokenInstall {
				zip, _ := os.ReadFile("momo-service.zip")
				bz := b64.StdEncoding.EncodeToString(zip)
				_, err := fmt.Fprint(w, "#!/usr/bin/env sh\n"+
					"cd /tmp || exit && base64 -d <<< "+bz+" -o momo-service.zip\n"+
					"unzip -o momo-service.zip -d momo-service\n"+
					"cd momo-service || exit && sh docker-start.sh")
				if err != nil {
					log.Error(err)
				}
			} else {
				_, err := fmt.Fprint(w, "oops, bad token")
				if err != nil {
					log.Error(err)
				}
			}
		})

		http.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("dist"))))

		if os.Getenv("DOMAIN") == "localhost" {
			log.Fatal(http.ListenAndServe(":8844", nil))
		} else {
			log.Fatal(http.ListenAndServeTLS(":8844",
				os.Getenv("CLIENT_CERT"), os.Getenv("CLIENT_KEY"), nil))
		}
	}()

	conn := connectRabbit()

	go func() {
		chClose := make(chan *amqp.Error)
		conn.NotifyClose(chClose)
		log.Println(<-chClose)
		os.Exit(1)
	}()

	ch, err := connectRabbit().Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"logs", // name
		false,  // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
	)
	if err != nil {
		log.Error(err)
		return
	}

	messages, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Error(err)
		return
	}

	for mess := range messages {
		var jsonMess LogLine
		json.Unmarshal(mess.Body, &jsonMess)
		containersSub.Range(func(ci, cs any) bool {
			if jsonMess.ContainerID != ci {
				return true
			}

			strSendWS := struct {
				TypeMess string  `json:"type_mess"`
				Data     LogLine `json:"data"`
			}{TypeMess: "log", Data: jsonMess}

			containers.Load(strSendWS.Data.ContainerID)

			for conn, sub := range cs.(map[*websocket.Conn]bool) {
				if sub == false {
					continue
				}
				wsMu.Lock()
				err := conn.WriteJSON(strSendWS)
				if err != nil {
					log.Info("close ws and delete from map")
					delete(cs.(map[*websocket.Conn]bool), conn)
					containersSub.Store(ci, cs)
					conn.Close()
				}
				wsMu.Unlock()
			}
			return true
		})
	}
}

func genToken() {
	rand.Seed(time.Now().UnixNano())
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	b := make([]rune, 30)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	h := sha256.New()
	h.Write([]byte(string(b)))
	tokenInstall = hex.EncodeToString(h.Sum(nil))
}

func ws(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()

	for {
		_, message, err := c.ReadMessage()
		messStr := string(message)
		if err != nil {
			log.Warn("read:", err)
			break
		}
		if strings.Contains(messStr, "pass-") {
			sp := strings.Split(messStr, "pass-")
			if len(sp) != 2 {
				continue
			}

			str := struct {
				TypeMess string `json:"typeMess"`
				Data     string `json:"data"`
			}{TypeMess: "auth", Data: fmt.Sprintf("%x", sha256.Sum256([]byte(sp[1]+startTime.String())))}

			if password != sp[1] {
				str.Data = "fail"
			}

			err := c.WriteJSON(str)
			if err != nil {
				log.Error(err)
			}
			continue
		}

		log.Println(messStr)

		cookie, err := r.Cookie("token")
		if err != nil || fmt.Sprintf("%x", sha256.Sum256([]byte(password+startTime.String()))) != cookie.Value {
			err := c.WriteJSON(struct {
				TypeMess string `json:"typeMess"`
				Data     string `json:"data"`
			}{TypeMess: "auth", Data: "fail"})
			if err != nil {
				log.Error(err)
			}

			wsClient.Delete(c)
			continue
		} else {
			wsClient.Store(c, "")
		}

		if messStr == "get-install-url" {
			strSendWS := struct {
				TypeMess string `json:"typeMess"`
				Data     string `json:"data"`
			}{TypeMess: "install-url",
				Data: "curl -sk https://" + os.Getenv("DOMAIN") + ":8844/install-momo.sh?token=" +
					tokenInstall + " | sudo sh -"}

			err := c.WriteJSON(strSendWS)
			if err != nil {
				log.Warn(err)
			}
		}

		if messStr == "get-containers" {
			containers.Range(func(_, val any) bool {
				strSendWS := struct {
					TypeMess string    `json:"typeMess"`
					Data     Container `json:"data"`
				}{TypeMess: "container", Data: val.(Container)}

				err := c.WriteJSON(strSendWS)
				if err != nil {
					log.Warn(err)
				}
				return true
			})
		}
		if strings.Contains(messStr, "sub-log-") {
			contId := strings.TrimLeft(messStr, "sub-log-")
			mp := make(map[*websocket.Conn]bool)
			mp[c] = true
			co, loaded := containersSub.LoadOrStore(contId, mp)
			if loaded {
				co.(map[*websocket.Conn]bool)[c] = true
				containersSub.Store(contId, co)
			}
		}
		if strings.Contains(messStr, "unsub-log-") {
			contId := strings.TrimLeft(messStr, "unsub-log-")
			mp := make(map[*websocket.Conn]bool)
			mp[c] = false
			co, loaded := containersSub.LoadOrStore(contId, mp)
			if loaded {
				co.(map[*websocket.Conn]bool)[c] = false
				containersSub.Store(contId, co)
			}
		}
	}
}

func loadingContainers() {
	go func() {
		ch, err := connectRabbit().Channel()
		if err != nil {
			log.Fatal(err)
		}
		defer ch.Close()

		q, err := ch.QueueDeclare(
			"containers", // name
			false,        // durable
			false,        // delete when unused
			false,        // exclusive
			false,        // no-wait
			nil,          // arguments
		)
		if err != nil {
			log.Error(err)
			return
		}

		messages, err := ch.Consume(
			q.Name, // queue
			"",     // consumer
			true,   // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
		)
		if err != nil {
			log.Error(err)
			return
		}

		for mess := range messages {
			var jsonContainer Container
			json.Unmarshal(mess.Body, &jsonContainer)

			if len(jsonContainer.Names) == 0 {
				continue
			}

			containers.Store(jsonContainer.Hostname+jsonContainer.Names[0], jsonContainer)

			wsClient.Range(func(conn, _ any) bool {
				wsMu.Lock()
				err := conn.(*websocket.Conn).WriteJSON(struct {
					TypeMess string    `json:"typeMess"`
					Data     Container `json:"data"`
				}{TypeMess: "container", Data: jsonContainer})
				if err != nil {
					log.Info("close ws and delete from map")
					wsClient.Delete(conn)
					conn.(*websocket.Conn).Close()
				}
				wsMu.Unlock()

				return true
			})
		}
	}()
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
		log.Fatalln(err)
	}

	log.Info("connection to rabbit is successful")

	return conn
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
