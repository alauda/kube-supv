#!/bin/bash

systemctl daemon-reload
systemctl start docker.service
systemctl enable docker.service
