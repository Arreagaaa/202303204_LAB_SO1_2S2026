{{- define "valkeyinstance.name" -}}
valkey-{{ .Release.Name }}
{{- end -}}
{{- define "valkeyinstance.fullname" -}}
{{ include "valkeyinstance.name" . }}
{{- end -}}
