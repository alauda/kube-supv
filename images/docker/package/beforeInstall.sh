#!/bin/bash

systemctl daemon-reload
systemctl stop docker.service || true
