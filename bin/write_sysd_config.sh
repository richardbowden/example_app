dbport=`aws ssm get-parameter --name "/database/url" --region eu-west-1 --with-decryption | jq '.Parameter.Value' -r | awk '{split($0,aa,":"); print aa[2]}'`
dbhost=`aws ssm get-parameter --name "/database/url" --region eu-west-1 --with-decryption | jq '.Parameter.Value' -r | awk '{split($0,aa,":"); print aa[1]}'`
dbuser=`aws ssm get-parameter --name "/database/user" --region eu-west-1 --with-decryption | jq '.Parameter.Value' -r`
dbpass=`aws ssm get-parameter --name "/database/passwd" --region eu-west-1 --with-decryption | jq '.Parameter.Value' -r`
dbname=`aws ssm get-parameter --name "/database/dbname" --region eu-west-1 --with-decryption | jq '.Parameter.Value' -r`

cat << EOF > /lib/systemd/system/exampleapp.conf
EXAMPLE_APP_DB_HOST=$dbhost
EXAMPLE_APP_DB_USER=$dbuser
EXAMPLE_APP_DB_PASSWD=$dbpass
EXAMPLE_APP_DB_PORT=$dbport
EXAMPLE_APP_DB_NAME=$dbname
EOF

cat << EOF > /lib/systemd/system/exampleapp.service
[Unit]
Description=example_app
After=network-online.target
Wants=network-online.target systemd-networkd-wait-online.service

[Service]
Restart=on-abnormal
WorkingDirectory=/app_server/

EnvironmentFile=/lib/systemd/system/exampleapp.conf

ExecStart=/app_server/example_crud_linux_amd64
ExecReload=/bin/kill -USR1 $MAINPID

; Use graceful shutdown with a reasonable timeout
KillMode=mixed
KillSignal=SIGQUIT
TimeoutStopSec=5s

LimitNOFILE=1048576
LimitNPROC=512

[Install]
WantedBy=multi-user.target
EOF

systemctl enable exampleapp.service
systemctl start exampleapp.service
