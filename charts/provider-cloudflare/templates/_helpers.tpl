{{/*
Expand the name of the chart.
*/}}
{{- define "provider-cloudflare.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "provider-cloudflare.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "provider-cloudflare.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "provider-cloudflare.labels" -}}
helm.sh/chart: {{ include "provider-cloudflare.chart" . }}
{{ include "provider-cloudflare.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- with .Values.additionalLabels }}
{{ toYaml . }}
{{- end }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "provider-cloudflare.selectorLabels" -}}
app.kubernetes.io/name: {{ include "provider-cloudflare.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Provider package image
*/}}
{{- define "provider-cloudflare.image" -}}
{{- $tag := .Values.provider.package.tag | default .Chart.AppVersion }}
{{- printf "%s:%s" .Values.provider.package.repository $tag }}
{{- end }}

{{/*
Runtime config name
*/}}
{{- define "provider-cloudflare.runtimeConfigName" -}}
{{- .Values.provider.runtimeConfig.name | default (printf "%s-config" (include "provider-cloudflare.fullname" .)) }}
{{- end }}

{{/*
ProviderConfig name
*/}}
{{- define "provider-cloudflare.providerConfigName" -}}
{{- .Values.providerConfig.name | default "default" }}
{{- end }}

{{/*
Secret name
*/}}
{{- define "provider-cloudflare.secretName" -}}
{{- .Values.secret.name | default "cloudflare-creds" }}
{{- end }}

{{/*
Common annotations (including ArgoCD sync wave if specified)
*/}}
{{- define "provider-cloudflare.annotations" -}}
{{- $annotations := .annotations | default dict }}
{{- if .syncWave }}
argocd.argoproj.io/sync-wave: {{ .syncWave | quote }}
{{- end }}
{{- with $annotations }}
{{ toYaml . }}
{{- end }}
{{- with $.root.Values.additionalAnnotations }}
{{ toYaml . }}
{{- end }}
{{- end }}
