{{- define "project.name" -}}
{{- default "project" .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* Helm required labels */}}
{{- define "project.labels" -}}
heritage: {{ .Release.Service }}
release: {{ .Release.Name }}
chart: {{ .Chart.Name }}
app: "{{ template "project.name" . }}"
{{- end -}}

{{/* matchLabels */}}
{{- define "project.matchLabels" -}}
release: {{ .Release.Name }}
app: "{{ template "project.name" . }}"
{{- end -}}