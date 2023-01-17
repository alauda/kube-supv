#!/bin/bash

systemctl daemon-reload
systemctl restart docker.service
systemctl enable docker.service
