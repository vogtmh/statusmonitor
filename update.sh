#!/bin/bash

appname="statusmonitor"
applabel="Status Monitor"

echo "Building executable .."
rm -rf build
GOOS=linux GOARCH=amd64 go build -o build/${appname}_linux main.go

if [ -f "/etc/systemd/system/${appname}.service" ]; then
  echo "Stopping service .."
  systemctl stop ${appname}
fi

echo "Updating executable .."
mkdir -p /opt/${appname}
mv build/${appname}_linux /opt/${appname}/${appname}_linux

echo "Updating service .."
mkdir -p /var/log/${appname}
cp template.service ${appname}.service
sed -i -e "s/APPNAME/${appname}/g" ${appname}.service
sed -i -e "s/APPLABEL/${applabel}/g" ${appname}.service
sudo mv ${appname}.service /etc/systemd/system/${appname}.service
sudo systemctl daemon-reload

echo "Enabling service .."
systemctl enable ${appname}
echo

echo "Starting service .."
systemctl start ${appname}
echo

systemctl status ${appname}
echo
