#!/bin/bash

systemctl stop docker
systemctl disable docker
rm -rf "${INSTALL_ROOT}/etc/systemd/system/docker.service"
systemctl daemon-reload
