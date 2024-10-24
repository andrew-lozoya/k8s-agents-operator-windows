{{/*
Returns if the template should render, it checks if the required values are set.
*/}}
{{- define "k8s-agents-operator.areValuesValid" -}}
{{- $licenseKey := include "newrelic.common.license._licenseKey" . -}}
{{- and (or $licenseKey)}}
{{- end -}}
{{/*
Create a node selector for Linux nodes
*/}}
{{- define "linux.nodeSelector" -}}
nodeSelector:
  os: linux  # Change this label according to your needs
{{- end -}}