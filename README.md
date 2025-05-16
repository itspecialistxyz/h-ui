<div align="center">

<a href="https://github.com/itspecialistxyz/h-ui"><img src="./docs/images/head-cover.png" alt="H UI" width="150" /></a>

<h1 align="center">Hysteria2 UI</h1>

<p align="center">
<a href="https://www.gnu.org/licenses/gpl-3.0.html"><img src="https://img.shields.io/github/license/itspecialistxyz/h-ui" alt="License: GPL-3.0"></a>
<a href="https://github.com/itspecialistxyz/h-ui/stargazers"><img src="https://img.shields.io/github/stars/itspecialistxyz/h-ui" alt="GitHub stars"></a>
<a href="https://github.com/itspecialistxyz/h-ui/forks"><img src="https://img.shields.io/github/forks/itspecialistxyz/h-ui" alt="GitHub forks"></a>
<a href="https://github.com/itspecialistxyz/h-ui/releases"><img src="https://img.shields.io/github/v/release/itspecialistxyz/h-ui" alt="GitHub release"></a>
<a href="https://hub.docker.com/r/itspecialistxyz/h-ui"><img src="https://img.shields.io/docker/pulls/itspecialistxyz/h-ui" alt="Docker pulls"></a>
<a href="https://github.com/itspecialistxyz/h-ui/actions/workflows/release.yml"><img src="https://github.com/itspecialistxyz/h-ui/actions/workflows/release.yml/badge.svg" alt="Build Status"></a>
<a href="https://github.com/itspecialistxyz/h-ui/actions/workflows/docker-build.yml"><img src="https://github.com/itspecialistxyz/h-ui/actions/workflows/docker-build.yml/badge.svg" alt="Docker Build Status"></a>
</p>

![cover](./docs/images/cover.png)

</div>

## Features

- Lightweight, low resource usage, easy to deploy
- Monitor system status and Hysteria2 status
- Limit user traffic, user online status, force users to log off, number of online users, reset user traffic
- Limit the number of users' online devices at the same time, the number of online devices
- User subscription link, node URL, import and export users
- Managing Hysteria2 configurations and Hysteria2 versions, port hopping
- Change the Web port, modify the Hysteria2 traffic multiplier
- Telegram notification
- View, import, and export system logs and Hysteria2 logs
- I18n: English, 简体中文
- Page adaptation, support night mode, custom page themes
- More features waiting for you to discover

## Recommended OS

OS: CentOS 8+/Ubuntu 20+/Debian 11+

CPU: x86_64/amd64 arm64/aarch64

Memory: ≥ 256MB

## Deployment

### Quick Install (Recommended)

Install Latest Version

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/itspecialistxyz/h-ui/main/install.sh)
```

Install [Custom Version](https://github.com/itspecialistxyz/h-ui/releases)

```bash
bash <(curl -fsSL https://raw.githubusercontent.com/itspecialistxyz/h-ui/main/install.sh) v0.0.1
```

### Manual Install
#### Download
Executable files: https://github.com/itspecialistxyz/h-ui/releases

#### Install
curl -fsSL https://github.com/itspecialistxyz/h-ui/releases/latest/download/h-ui-linux-amd64 -o /usr/local/h-ui/h-ui && chmod +x /usr/local/h-ui/h-ui
curl -fsSL https://raw.githubusercontent.com/itspecialistxyz/h-ui/main/h-ui.service -o /etc/systemd/system/h-ui.service
# Custom web port, default is 8081
# sed -i "s|^ExecStart=.*|ExecStart=/usr/local/h-ui/h-ui -p 8081|" "/etc/systemd/system/h-ui.service"
systemctl daemon-reload
systemctl enable h-ui
systemctl restart h-ui
```

Uninstall

```bash
systemctl stop h-ui
rm -rf /etc/systemd/system/h-ui.service /usr/local/h-ui/
```

### Docker

1. Install Docker

   https://docs.docker.com/engine/install/

   ```bash
   bash <(curl -fsSL https://get.docker.com)
   ```

2. Start a container

   ```bash
   docker pull itspecialistxyz/h-ui

   docker run -d --cap-add=NET_ADMIN \
     --name h-ui --restart always \
     --network=host \
     -v /h-ui/bin:/h-ui/bin \
     -v /h-ui/data:/h-ui/data \
     -v /h-ui/export:/h-ui/export \
     -v /h-ui/logs:/h-ui/logs \
     itspecialistxyz/h-ui
   ```

   Custom web port, default is 8081

   ```bash
   docker run -d --cap-add=NET_ADMIN \
     --name h-ui --restart always \
     --network=host \
     -v /h-ui/bin:/h-ui/bin \
     -v /h-ui/data:/h-ui/data \
     -v /h-ui/export:/h-ui/export \
     -v /h-ui/logs:/h-ui/logs \
     itspecialistxyz/h-ui \
     ./h-ui -p 8081
   ```

   Set the time zone, default is Asia/Shanghai

   ```bash
   docker run -d --cap-add=NET_ADMIN \
     --name h-ui --restart always \
     --network=host \
     -e TZ=Asia/Shanghai \
     -v /h-ui/bin:/h-ui/bin \
     -v /h-ui/data:/h-ui/data \
     -v /h-ui/export:/h-ui/export \
     -v /h-ui/logs:/h-ui/logs \
     itspecialistxyz/h-ui
   ```

Uninstall

```bash
docker rm -f h-ui
docker rmi itspecialistxyz/h-ui
rm -rf /h-ui
```

## Default Installation Information

- Panel Port: 8081
- SSH local forwarded port: 8082
- Login Username/Password: Random 6 characters
- Connection Password: {Login Username}.{Login Password}

## System Upgrade

Export the user, system configuration, and Hysteria2 configuration in the management background, redeploy the latest
version of h-ui, and import the data into the management background after the deployment is complete.

## FAQ

[English > FAQ](./docs/FAQ.md)

## Performance Optimization

- Scheduled server restart

    ```bash
    0 4 * * * /sbin/reboot
    ```

- Install Network Accelerator
    - [TCP Brutal](https://github.com/apernet/tcp-brutal) (Recommended)
    - [teddysun/across#bbrsh](https://github.com/teddysun/across#bbrsh)
    - [Chikage0o0/Linux-NetSpeed](https://github.com/ylx2016/Linux-NetSpeed)
    - [ylx2016/Linux-NetSpeed](https://github.com/ylx2016/Linux-NetSpeed)

## Client

https://v2.hysteria.network/docs/getting-started/3rd-party-apps/

## Development

Go >= 1.20, Node.js >= 18.12.0

- frontend

   ```bash
   cd frontend
   pnpm install
   npm run dev
   ```

- backend

   ```bash
   go run main.go
   ```

## Build

- frontend

   ```bash
   npm run build:prod
   ```

- backend

  Windows: [build.bat](build.bat)

  Mac/Linux: [build.sh](build.sh)

## Other

Telegram Channel: https://t.me/jonssonyan_channel

You can subscribe to my channel on YouTube: https://www.youtube.com/@jonssonyan

## Contributors

Thanks to everyone who contributed to this project.

<a href="https://github.com/jonssonyan/h-ui/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=jonssonyan/h-ui" />
</a>

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=jonssonyan/h-ui&type=Date)](https://star-history.com/#jonssonyan/h-ui&Date)

## License

[GPL-3.0](LICENSE)