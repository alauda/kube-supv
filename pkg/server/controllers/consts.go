package controllers

import "time"

const (
	reconcileAfterDuration       = time.Second * 2
	reconcileHealthCheckDuration = time.Second * 300
	finalizer                    = "kubesupv"
	checkerNameTemplate          = "kubesupv-checker-%s"
	checkerPodTemplate           = `
apiVersion: v1
kind: Pod
metadata:
  name: "kubesupv-checker- {{ .NodeName }}"
  namespace: kube-system
spec:
  containers:
  - command: ["nsenter", "-t", "1", "-m", "-u", "-i", "-n", "-p", "kubesupv", "package", "upload", "--node={{ .NodeName }}"]
    image: {{ .Image }}
	name: checker
	securityContext:
      privileged: true
	hostNetwork: true
	hostPID: true
	nodeName: "{{ .NodeName }}"
	restartPolicy: Never
	tolerations:
	- operator: Exists
`
)
