{{/*
Shared helpers for PROMPT subcharts.
All helpers take the subchart root context (`.`) unless noted; `.Values.global`
is available in every subchart because Helm merges the global tree.
*/}}

{{/* Release-scoped resource name: <release>-<suffix> */}}
{{- define "prompt.name" -}}
{{- printf "%s-%s" .root.Release.Name .suffix | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/* Fully-qualified image ref for a component. dict: root, image, tag */}}
{{- define "prompt.image" -}}
{{- $g := .root.Values.global.image -}}
{{- $tag := default $g.tag .tag -}}
{{- printf "%s/%s:%s" $g.registry .image $tag -}}
{{- end -}}

{{/* Common labels. dict: root, component(name), part */}}
{{- define "prompt.labels" -}}
helm.sh/chart: {{ printf "%s-%s" .root.Chart.Name .root.Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
app.kubernetes.io/managed-by: {{ .root.Release.Service }}
app.kubernetes.io/instance: {{ .root.Release.Name }}
app.kubernetes.io/part-of: prompt
app.kubernetes.io/name: {{ .component }}
{{- end -}}

{{/* Selector labels. dict: root, component(name) */}}
{{- define "prompt.selectorLabels" -}}
app.kubernetes.io/instance: {{ .root.Release.Name }}
app.kubernetes.io/name: {{ .component }}
{{- end -}}

{{/* Names of shared infra resources, derived from the release name. */}}
{{- define "prompt.appConfigName" -}}{{ printf "%s-app-config" .Release.Name }}{{- end -}}
{{- define "prompt.appSecretName" -}}
{{- $existing := .Values.global.appSecrets.existingSecret -}}
{{- if $existing }}{{ $existing }}{{ else }}{{ printf "%s-app-secrets" .Release.Name }}{{ end -}}
{{- end -}}
{{- define "prompt.dbSecretName" -}}{{ printf "%s-db-%s" .root.Release.Name .dbKey }}{{- end -}}

{{/* Host used by the DB-wait init container. dict: root */}}
{{- define "prompt.dbWaitHost" -}}
{{- $pg := .root.Values.global.postgresql -}}
{{- if eq $pg.mode "external" -}}
{{ $pg.external.host }}
{{- else if $pg.pooler.enabled -}}
{{ printf "%s-pooler-rw" $pg.clusterName }}
{{- else -}}
{{ printf "%s-rw" $pg.clusterName }}
{{- end -}}
{{- end -}}

{{- define "prompt.dbWaitPort" -}}
{{- $pg := .root.Values.global.postgresql -}}
{{- if eq $pg.mode "external" -}}{{ $pg.external.port }}{{- else -}}5432{{- end -}}
{{- end -}}
