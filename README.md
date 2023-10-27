# auto-southwest-check-in-gui
Golang Gui for [auto-southwest-check-in](https://github.com/jdholtz/auto-southwest-check-in)

## Install

```
apt install -y tmux
git clone ...
cd ...
go run website.go
```

By default it only runs on localhost:8080 and expects southwest.py to be in `/root/auto-southwest-check-in` if you don't update the paths

If you use this it's a good idea to pair it with a reverse proxy with authentication.

```
wget https://github.com/caddyserver/caddy/releases/download/v2.7.5/caddy_2.7.5_linux_amd64.tar.gz
tar xf caddy_2.7.5_linux_amd64.tar.gz
CADDYUSER="admin"
CADDYPASS=`./caddy hash-password -p REPLACEPASSWORD`
SUBDOMAIN="yoursubdomain"
cat > Caddyfile << EOF
$SUBDOMAIN.duckdns.org {
        reverse_proxy * 127.0.0.1:8080
        basicauth {
                $USER $CADDYPASS
        }
}
EOF
./caddy run
```
## Screenshot

![image](https://github.com/mkunz7/auto-southwest-check-in-gui/assets/5001991/4389cd32-cda8-411f-b254-0f0696a56f95)
