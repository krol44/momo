###  The Momo is grabbing logs from docker containers

### demo screen
![demo momo](https://github.com/krol44/momo/raw/master/readme/demo-screen.png?raw=true)

### setup

1. sudo -s && cd /home
2. git clone https://github.com/krol44/momo
3. nano momo/.env.secrets
```
COMPOSE_PROJECT_NAME=momo
DOMAIN=localhost
DOMAIN_MOMO_SERVICE=localhost
DASHBOARD_PASS=password_for_web_admin
RABBIT_PASS=password_for_rabbitMQ
CA_CERT=/ssl/ca-cert.pem
CLIENT_CERT=/ssl/client-cert.pem
CLIENT_KEY=/ssl/client-key.pem
```
4. change and save - DOMAIN, DOMAIN_MOMO_SERVICE, DASHBOARD_PASS and RABBIT_PASS
5. cd momo/rabbitMQ/gen-ssl/
6. ./docker-start.sh
7. cd ../../
8. ./build
9. open url - https://localhost:8844, login, copy link and install any machine
10. enjoy :3