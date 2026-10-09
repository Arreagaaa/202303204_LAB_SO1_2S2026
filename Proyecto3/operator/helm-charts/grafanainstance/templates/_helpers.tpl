{{- define "grafanainstance.name" -}}
grafana-{{ .Release.Name }}
{{- end -}}
{{- define "grafanainstance.fullname" -}}
{{ include "grafanainstance.name" . }}
{{- end -}}
