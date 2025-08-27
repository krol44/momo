###  The MoMo is grabbing logs from docker containers
1. Relatime views logs on the all machines
2. Stats of the containers
3. Notification error or info messages to the telegram bot
4. Export statistics to prometheus
5. Ready-made template for grafana

# demo screen
![demo momo](https://github.com/krol44/momo/raw/master/readme/demo-screen.png?raw=true)

# demo grafana
![demo grafana](https://github.com/krol44/momo/raw/master/readme/demo-grafana.jpg?raw=true)

# setup and run local

1. `sudo -s && cd /home`
2. `git clone https://github.com/krol44/momo`
3. `nano momo/.env.secrets`
```
COMPOSE_PROJECT_NAME=momo

DOMAIN=localhost
DOMAIN_MOMO_SERVICE=localhost

DASHBOARD_PASS=password_for_web_admin
RABBIT_PASS=password_for_rabbitMQ

TG_BOT_TOKEN=0000000:AAAAAAAAAAAAAAAA
```
4. change `DASHBOARD_PASS`, `RABBIT_PASS`, `TG_BOT_TOKEN` and save
5. `cd momo/rabbitMQ/gen-ssl/`
6. `./docker-start.sh` – creating ssl cert
7. `cd ../../`
8. `./build`
9. open url `https://localhost:8844` to login, you may copy link in the setting page and install own machine
10. enjoy :3
11. https://localhost:8777/metrics – add to prometheus
12. Import your grafana json model – readme/grafana-model.json
