###  The MoMo is grabbing logs from docker containers
1. Relatime views logs on the all machines
2. Notication error or info messages to the telegram bot

# demo screen
![demo momo](https://github.com/krol44/momo/raw/master/readme/demo-screen.png?raw=true)

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

CA_CERT=/ssl/ca-cert.pem
CLIENT_CERT=/ssl/client-cert.pem
CLIENT_KEY=/ssl/client-key.pem
```
4. change and save - `DOMAIN`, `DOMAIN_MOMO_SERVICE`, `DASHBOARD_PASS`, `RABBIT_PASS`, `TG_BOT_TOKEN`
5. cd momo/rabbitMQ/gen-ssl/
6. ./docker-start.sh
7. cd ../../
8. ./build
9. open url - https://localhost:8844, login, you may copy link in the setting and install own machine
10. enjoy :3