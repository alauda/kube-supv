#!/bin/bash

systemctl stop docker
systemctl disable docker
rm -rf /etc/systemd/system/docker.service
systemctl daemon-reload
