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

{{/* ---------------------------------------------------------------------------
Endpoints derived from global values. Used by the infrastructure subchart and by
prompt.appConfigData, so they live here to keep this library chart standalone.
--------------------------------------------------------------------------- */}}

{{- define "prompt.baseUrl" -}}
{{- printf "%s://%s" .Values.global.scheme .Values.global.host -}}
{{- end -}}

{{- define "prompt.keycloakHost" -}}
{{- $kc := .Values.global.keycloak -}}
{{- if eq $kc.mode "in-cluster" -}}
{{ printf "%s://auth.%s" .Values.global.scheme .Values.global.host }}
{{- else -}}
{{ $kc.externalHost }}
{{- end -}}
{{- end -}}

{{- define "prompt.s3Endpoint" -}}
{{- $os := .Values.global.objectStorage -}}
{{- if eq $os.mode "in-cluster" -}}
{{ printf "http://%s-seaweedfs-s3:8333" .Release.Name }}
{{- else -}}
{{ required "global.objectStorage.external.endpoint is required when global.objectStorage.mode=external" $os.external.endpoint }}
{{- end -}}
{{- end -}}

{{- define "prompt.s3PublicEndpoint" -}}
{{- $os := .Values.global.objectStorage -}}
{{- if eq $os.mode "in-cluster" -}}
{{ printf "%s://s3.%s" .Values.global.scheme .Values.global.host }}
{{- else -}}
{{ required "global.objectStorage.external.publicEndpoint is required when global.objectStorage.mode=external" $os.external.publicEndpoint }}
{{- end -}}
{{- end -}}

{{/* ---------------------------------------------------------------------------
Database connection settings. Every consumer (per-phase Secret, DB-wait init
container, checksum) resolves through these so they cannot disagree.
--------------------------------------------------------------------------- */}}

{{/* dict: root */}}
{{- define "prompt.dbHost" -}}
{{- $pg := .root.Values.global.postgresql -}}
{{- if eq $pg.mode "external" -}}
{{ required "global.postgresql.external.host is required when global.postgresql.mode=external" $pg.external.host }}
{{- else if $pg.pooler.enabled -}}
{{ printf "%s-pooler-rw" $pg.clusterName }}
{{- else -}}
{{ printf "%s-rw" $pg.clusterName }}
{{- end -}}
{{- end -}}

{{/* dict: root */}}
{{- define "prompt.dbPort" -}}
{{- $pg := .root.Values.global.postgresql -}}
{{- if eq $pg.mode "external" -}}{{ $pg.external.port }}{{- else -}}5432{{- end -}}
{{- end -}}

{{/* dict: root. Explicit value wins; otherwise TLS is required off-cluster. */}}
{{- define "prompt.dbSslMode" -}}
{{- $pg := .root.Values.global.postgresql -}}
{{- if $pg.sslMode -}}
{{ $pg.sslMode }}
{{- else if eq $pg.mode "external" -}}
require
{{- else -}}
disable
{{- end -}}
{{- end -}}

{{/* dict: root, dbKey. External installs use one account for every database,
matching the Compose contract; per-phase owners are managed by CNPG only. */}}
{{- define "prompt.dbUser" -}}
{{- $pg := .root.Values.global.postgresql -}}
{{- if eq $pg.mode "external" -}}
{{ required "global.postgresql.external.user is required when global.postgresql.mode=external" $pg.external.user }}
{{- else -}}
{{ (index $pg.databases .dbKey).owner }}
{{- end -}}
{{- end -}}

{{/* dict: root */}}
{{- define "prompt.dbExternalPassword" -}}
{{ required "global.postgresql.external.password is required when global.postgresql.mode=external" .root.Values.global.postgresql.external.password }}
{{- end -}}

{{/* dict: root, dbKey. Hashed into a pod annotation so a changed connection
rolls the backends. Excludes the in-cluster role password on purpose: it comes
from `lookup`, and re-resolving it here would produce a different value than the
Secret got on a lookup miss, changing the hash on every render. */}}
{{- define "prompt.dbSecretInputs" -}}
mode={{ .root.Values.global.postgresql.mode }}
host={{ include "prompt.dbHost" . }}
port={{ include "prompt.dbPort" . }}
user={{ include "prompt.dbUser" . }}
name={{ (index .root.Values.global.postgresql.databases .dbKey).dbName }}
sslmode={{ include "prompt.dbSslMode" . }}
{{- if eq .root.Values.global.postgresql.mode "external" }}
password={{ include "prompt.dbExternalPassword" . }}
{{- end }}
{{- end -}}

{{/* ---------------------------------------------------------------------------
Shared application config. Defined here rather than inline in the infrastructure
subchart so every workload can hash the exact rendered body (checksum/appconfig)
and roll when it changes.
Takes the subchart root context (`.`).
--------------------------------------------------------------------------- */}}
{{- define "prompt.appConfigData" -}}
{{- $g := .Values.global -}}
{{- $base := include "prompt.baseUrl" . -}}
ENVIRONMENT: {{ $g.environment | quote }}
CORE_HOST: {{ $base | quote }}
SERVER_CORE_HOST: {{ printf "http://%s-core-server:8080" .Release.Name | quote }}
SERVER_ADDRESS: {{ $g.appConfig.serverAddress | quote }}
KEYCLOAK_HOST: {{ include "prompt.keycloakHost" . | quote }}
KEYCLOAK_REALM_NAME: {{ $g.keycloak.realm | quote }}
KEYCLOAK_CLIENT_ID: {{ $g.keycloak.clientId | quote }}
KEYCLOAK_ID_OF_CLIENT: {{ $g.keycloak.idOfClient | quote }}
KEYCLOAK_AUTHORIZED_PARTY: {{ $g.keycloak.authorizedParty | quote }}
SSL_MODE: {{ include "prompt.dbSslMode" (dict "root" .) | quote }}
SENTRY_ENABLED: {{ $g.appConfig.sentryEnabled | quote }}
MAX_FILE_UPLOAD_SIZE_MB: {{ $g.appConfig.maxFileUploadSizeMb | quote }}
ALLOWED_FILE_TYPES: {{ $g.appConfig.allowedFileTypes | quote }}
S3_BUCKET: {{ $g.objectStorage.bucket | quote }}
S3_REGION: {{ $g.objectStorage.region | quote }}
S3_ENDPOINT: {{ include "prompt.s3Endpoint" . | quote }}
S3_PUBLIC_ENDPOINT: {{ include "prompt.s3PublicEndpoint" . | quote }}
S3_FORCE_PATH_STYLE: {{ $g.objectStorage.forcePathStyle | quote }}
S3_PRESIGN_UPLOAD_TTL_SECONDS: {{ $g.objectStorage.presignUploadTtlSeconds | quote }}
S3_PRESIGN_DOWNLOAD_TTL_SECONDS: {{ $g.objectStorage.presignDownloadTtlSeconds | quote }}
CHAIR_NAME_SHORT: {{ $g.chairNameShort | quote }}
CHAIR_NAME_LONG: {{ $g.chairNameLong | quote }}
GITHUB_SHA: {{ $g.githubSha | quote }}
GITHUB_REF: {{ $g.githubRef | quote }}
SERVER_CORE_IMAGE_TAG: {{ $g.image.tag | quote }}
CORE_API_HOST: {{ $base | quote }}
ASSESSMENT_HOST: {{ $base | quote }}
INTERVIEW_HOST: {{ $base | quote }}
TEAM_ALLOCATION_HOST: {{ $base | quote }}
SELF_TEAM_ALLOCATION_HOST: {{ $base | quote }}
CERTIFICATE_HOST: {{ $base | quote }}
EXAMPLE_HOST: {{ default $base $g.externalRemotes.exampleServerHost | quote }}
INTRO_COURSE_HOST: {{ $g.externalRemotes.introCourseHost | quote }}
DEVOPS_CHALLENGE_HOST: {{ $g.externalRemotes.devopsChallengeHost | quote }}
{{- end -}}
