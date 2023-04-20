package main

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	b64 "encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/krol44/telegram-bot-api"
	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
	"io"
	"math/rand"
	_ "modernc.org/sqlite"
	"net/http"
	"os"
	"path"
	"runtime"
	"strconv"
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
var bot *tgbotapi.BotAPI
var alertKeys sync.Map
var alertChan chan Line
var manyRequestMu sync.Mutex
var manyRequest bool

var upgrader = websocket.Upgrader{}

func main() {
	initTables()
	startTime = time.Now()
	password = os.Getenv("DASHBOARD_PASS")

	logSetup()
	genToken()
	gettingStatistic()
	tgBot()
	gettingAlerts()
	alertChan = make(chan Line, 0)

	// alert find
	go func() {
		for true {
			f := findAlert()
			go sendAlert(f)
		}
	}()

	// dashboard
	go func() {
		mux := http.NewServeMux()
		log.Info("start dashboard")
		mux.HandleFunc("/ws", ws)
		mux.HandleFunc("/install-momo.sh", func(w http.ResponseWriter, r *http.Request) {
			if r.FormValue("token") == tokenInstall {
				zip, _ := os.ReadFile("momo-service.zip")
				bz := b64.StdEncoding.EncodeToString(zip)
				_, err := fmt.Fprint(w, "#!/usr/bin/env bash\n"+
					"cd /tmp || exit && base64 -d <<< "+bz+" > momo-service.zip\n"+
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

		mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("dist"))))
		for _, v := range []string{"/stats", "/alert", "/setting"} {
			mux.HandleFunc(v, func(w http.ResponseWriter, r *http.Request) {
				index, _ := os.ReadFile("dist/index.html")
				w.Write(index)
			})
		}

		if os.Getenv("DOMAIN") == "localhost" {
			log.Fatal(http.ListenAndServe(":8844", mux))
		} else {
			log.Fatal(http.ListenAndServeTLS(":8844",
				os.Getenv("CLIENT_CERT"), os.Getenv("CLIENT_KEY"), mux))
		}
	}()

	go gettingContainers()
	go gettingStats()
	gettingLogs()
}

func gettingStats() {
	conn := connectRabbit("stats")
	go func() {
		chClose := make(chan *amqp.Error)
		conn.NotifyClose(chClose)
		log.Info(<-chClose)
		os.Exit(1)
	}()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"stats", // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
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
		var jsonMess Line
		json.Unmarshal(mess.Body, &jsonMess)

		jsonMess.Md5Name = fmt.Sprintf("%x",
			md5.Sum([]byte(jsonMess.Hostname+jsonMess.Name)))

		wsClient.Range(func(conn, _ any) bool {
			wsMu.Lock()
			err := conn.(*websocket.Conn).WriteJSON(struct {
				TypeMess string `json:"typeMess"`
				Data     Line   `json:"data"`
			}{TypeMess: "stats", Data: jsonMess})
			if err != nil {
				wsClient.Delete(conn)
				conn.(*websocket.Conn).Close()
			}
			wsMu.Unlock()

			return true
		})
	}
}

func gettingLogs() {
	conn := connectRabbit("logs")
	go func() {
		chClose := make(chan *amqp.Error)
		conn.NotifyClose(chClose)
		log.Info(<-chClose)
		os.Exit(1)
	}()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"logs", // name
		true,   // durable
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
		var jsonMess Line
		json.Unmarshal(mess.Body, &jsonMess)

		jsonMess.Md5Name = fmt.Sprintf("%x",
			md5.Sum([]byte(jsonMess.Hostname+jsonMess.Name)))

		alertChan <- jsonMess

		containersSub.Range(func(md5Name, cs any) bool {
			if jsonMess.Md5Name != md5Name {
				return true
			}

			strSendWS := struct {
				TypeMess string `json:"typeMess"`
				Data     Line   `json:"data"`
			}{TypeMess: "log", Data: jsonMess}

			strSendWS.Data.Md5Name = fmt.Sprintf("%x",
				md5.Sum([]byte(strSendWS.Data.Hostname+strSendWS.Data.Name)))

			for conn, sub := range cs.(map[*websocket.Conn]bool) {
				if sub == false {
					continue
				}
				wsMu.Lock()
				err := conn.WriteJSON(strSendWS)
				if err != nil {
					delete(cs.(map[*websocket.Conn]bool), conn)
					containersSub.Store(md5Name, cs)
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
					tokenInstall + " | sudo bash -"}

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
			md5NameCont := strings.TrimPrefix(messStr, "sub-log-")
			mp := make(map[*websocket.Conn]bool)
			mp[c] = true
			co, loaded := containersSub.LoadOrStore(md5NameCont, mp)

			if loaded {
				co.(map[*websocket.Conn]bool)[c] = true
				containersSub.Store(md5NameCont, co)
			}
		}
		if strings.Contains(messStr, "unsub-log-") {
			md5NameCont := strings.TrimPrefix(messStr, "unsub-log-")
			mp := make(map[*websocket.Conn]bool)
			mp[c] = false
			co, loaded := containersSub.LoadOrStore(md5NameCont, mp)
			if loaded {
				co.(map[*websocket.Conn]bool)[c] = false
				containersSub.Store(md5NameCont, co)
			}
		}

		if strings.Contains(messStr, "get-alerts") {
			var tgc []TelegramChat
			err := sqlite().Select(&tgc, `SELECT telegram_id, telegram_name FROM users`)
			if err != nil {
				log.Error(err)
			}

			var al []Alert
			err = sqlite().
				Select(&al, `SELECT a.id, a.container_md5, a.telegram_id, a.key_alert, a.date_create,
       										u.telegram_name
									FROM alerts a
									LEFT JOIN users u on u.telegram_id = a.telegram_id`)
			if err != nil {
				log.Error(err)
			}
			sqlite().Close()

			strSendWS := struct {
				TypeMess        string         `json:"typeMess"`
				Telegrams       []TelegramChat `json:"telegrams"`
				TelegramBotName string         `json:"telegram_bot_name"`
				Alerts          []Alert        `json:"alerts"`
			}{TypeMess: "alerts", Telegrams: tgc, TelegramBotName: bot.Self.UserName, Alerts: al}

			err = c.WriteJSON(strSendWS)
			if err != nil {
				log.Warn(err)
			}
		}
		if strings.Contains(messStr, "add-alert-") {
			var j struct {
				TelegramID string `json:"telegram_id"`
				KeyAlert   string `json:"key_alert"`
				Md5        string `json:"md5"`
			}
			err := json.Unmarshal([]byte(strings.TrimPrefix(messStr, "add-alert-")), &j)
			if err != nil {
				log.Error(j)
				continue
			}
			if j.TelegramID == "" {
				continue
			}

			var al Alert
			err = sqlite().Get(&al, `SELECT id, container_md5, telegram_id, key_alert FROM alerts 
			                                             WHERE container_md5 = ? AND telegram_id = ? AND key_alert = ?`,
				j.Md5, j.TelegramID, j.KeyAlert)
			if err != nil && sql.ErrNoRows != err {
				log.Error(err)
				continue
			}
			if al.ID != 0 {
				continue
			}

			sqlite().Query(`INSERT INTO alerts (container_md5, telegram_id, key_alert, date_create)
									VALUES (?, ?, ?, datetime('now'))`, j.Md5, j.TelegramID, j.KeyAlert)
			sqlite().Close()
		}
		if strings.Contains(messStr, "rm-alert-") {
			sqlite().Query(`DELETE FROM alerts WHERE id = ?`, strings.TrimPrefix(messStr, "rm-alert-"))
			sqlite().Close()
		}
	}
}

func gettingContainers() {
	go func() {
		for {
			time.Sleep(time.Second * 10)
			containers.Range(func(key, _ any) bool {
				containers.Delete(key)
				return true
			})
		}
	}()

	ch, err := connectRabbit("containers").Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"containers", // name
		true,         // durable
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

		jsonContainer.Md5Name = fmt.Sprintf("%x",
			md5.Sum([]byte(jsonContainer.Hostname+jsonContainer.Names[0])))

		containers.Store(jsonContainer.Md5Name, jsonContainer)

		wsClient.Range(func(conn, _ any) bool {
			wsMu.Lock()
			err := conn.(*websocket.Conn).WriteJSON(struct {
				TypeMess string    `json:"typeMess"`
				Data     Container `json:"data"`
			}{TypeMess: "container", Data: jsonContainer})
			if err != nil {
				wsClient.Delete(conn)
				conn.(*websocket.Conn).Close()
			}
			wsMu.Unlock()

			return true
		})
	}
}

func connectRabbit(typeCh string) *amqp.Connection {
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

	log.Info("connection to rabbit is successful - " + typeCh)

	return conn
}

func gettingStatistic() {
	go func() {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		for {
			url := "https://" + strings.Replace(os.Getenv("AMQP_URL"), ":5671", ":15671/api/overview", 1)

			get, err := http.Get(url)
			if err != nil {
				log.Error(err)
			}
			body, _ := io.ReadAll(get.Body)
			//Statistic
			var js Statistic
			errUn := json.Unmarshal(body, &js)
			if errUn != nil {
				log.Error(err)
			}

			wsClient.Range(func(conn, _ any) bool {
				wsMu.Lock()
				err := conn.(*websocket.Conn).WriteJSON(struct {
					TypeMess string    `json:"typeMess"`
					Data     Statistic `json:"data"`
				}{TypeMess: "statistic", Data: js})
				if err != nil {
					log.Info("close ws and delete from map")
					wsClient.Delete(conn)
					conn.(*websocket.Conn).Close()
				}
				wsMu.Unlock()

				return true
			})

			time.Sleep(time.Second)
		}
	}()
}

func sqlite() *sqlx.DB {
	conn, err := sqlx.Connect("sqlite", "sqlite/store.db")
	if err != nil {
		log.Error(err)
	}
	return conn
}

func initTables() {
	db := sqlite()
	defer db.Close()

	if _, err := db.Exec(`
create table if not exists users
(
  telegram_id   BIGINT,
  telegram_name TEXT,
  date_create   TEXT
);

create unique index if not exists users_telegram_id_uindex
    on users (telegram_id);

create table if not exists alerts
(
    id integer
        constraint alerts_pk
            primary key autoincrement,
    container_md5       TEXT,
    telegram_id     	TEXT,
    key_alert 			TEXT,
    date_create			TEXT
);`); err != nil {
		log.Error(err)
	}
}

func tgBot() {
	botApi, err := tgbotapi.NewBotAPI(os.Getenv("TG_BOT_TOKEN"))

	if err != nil {
		log.Panic(err)
		os.Exit(1)
	}
	botApi.Debug = false

	log.Infof("Authorized on account %s", botApi.Self.UserName)

	bot = botApi

	go func(botIn *tgbotapi.BotAPI) {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60

		for u := range botIn.GetUpdatesChan(u) {
			if u.Message != nil {
				if u.Message.Text == "/start" {
					db := sqlite()
					user := struct {
						TelegramId int64 `db:"telegram_id"`
					}{}
					_ = db.Get(&user, "SELECT telegram_id FROM users WHERE telegram_id = ?",
						u.Message.From.ID)

					if user.TelegramId == 0 {
						_, err := db.Exec(`INSERT INTO users (telegram_id, telegram_name, date_create)
							VALUES (?, ?, datetime('now'))`, u.Message.From.ID, u.Message.From.UserName)
						if err != nil {
							log.Error(err)
						}
					}
					db.Close()

					botIn.Send(tgbotapi.NewSticker(u.Message.Chat.ID,
						tgbotapi.FileID("CAACAgIAAxkBAAMHZDqdPa-ipjZbt5tFJ6g0rMNqc6gAAjEAAygPahTT_70FDNZySC8E")))
				}
			}
		}
	}(botApi)
}

func gettingAlerts() {
	go func() {
		db := sqlite()
		defer db.Close()
		for true {
			var data []Alert
			err := db.Select(&data, `SELECT id, container_md5, telegram_id, key_alert FROM alerts`)
			if err != nil {
				log.Fatal(err)
			}

			for _, val := range data {
				alertKeys.Store(val.ID, val)
			}

			alertKeys.Range(func(key, _ any) bool {
				for _, val := range data {
					if val.ID == key {
						return true
					}
				}
				alertKeys.Delete(key)
				return true
			})

			time.Sleep(time.Second * 5)
		}
	}()
}

func findAlert() PreparedAlert {
	ticker := time.NewTicker(time.Second * 2)
	mapSend := make(PreparedAlert)

	for v := range alertChan {
		alertKeys.Range(func(_, k any) bool {
			if v.Md5Name != k.(Alert).ContainerMd5 {
				return true
			}
			if strings.Contains(strings.ToLower(v.Body), strings.ToLower(k.(Alert).KeyAlert)) {
				alertData := k.(Alert)
				mapSend[v.Md5Name+alertData.KeyAlert] = append(mapSend[v.Md5Name+alertData.KeyAlert],
					struct {
						Alert Alert
						Data  Line
					}{alertData, v})
			}
			return true
		})

		select {
		case <-ticker.C:
			return mapSend
		default:
			for _, v := range mapSend {
				if len(v) >= 20 {
					return mapSend
				}
			}
			continue
		}
	}

	return mapSend
}

func sendAlert(a PreparedAlert) {
	if manyRequest {
		log.Error("too many messages are sending in tg")
		return
	}
	for _, v := range a {
		var (
			info string
			lg   string
		)
		cl := map[string]bool{}
		for _, l := range v {
			lg += l.Data.Body + "\n"
			info = "<b>Key alert:</b> " + l.Alert.KeyAlert + " — " + l.Data.Hostname + " <b>" + l.Data.Name + "</b>\n\n"
			cl[l.Alert.TelegramID] = true
		}

		for c := range cl {
			ci, _ := strconv.Atoi(c)
			if ci == 0 {
				continue
			}
			mess := tgbotapi.NewMessage(int64(ci), info+lg)
			mess.DisableWebPagePreview = true
			mess.ParseMode = tgbotapi.ModeHTML
			_, err := bot.Send(mess)
			if err != nil {
				log.Warn(err)
				if strings.Contains(err.Error(), "Too Many Requests") {
					go func() {
						manyRequestMu.Lock()
						manyRequest = true
						manyRequestMu.Unlock()
						time.Sleep(time.Minute * 10)
						manyRequestMu.Lock()
						manyRequest = false
						manyRequestMu.Unlock()
					}()
				}
			}
		}
	}
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
