package controllers

import "time"

const (
	reconcileAfterDuration       = time.Second * 2
	reconcileHealthCheckDuration = time.Second * 300
	finalizer                    = "kubesupv"
)
