{{/* Infrastructure-local derivations of shared endpoints/URLs. */}}

{{- define "infra.baseUrl" -}}
{{- printf "%s://%s" .Values.global.scheme .Values.global.host -}}
{{- end -}}

{{- define "infra.keycloakHost" -}}
{{- $kc := .Values.global.keycloak -}}
{{- if eq $kc.mode "in-cluster" -}}
{{ printf "%s://auth.%s" .Values.global.scheme .Values.global.host }}
{{- else -}}
{{ $kc.externalHost }}
{{- end -}}
{{- end -}}

{{- define "infra.s3Endpoint" -}}
{{- $os := .Values.global.objectStorage -}}
{{- if eq $os.mode "in-cluster" -}}
{{ printf "http://%s-seaweedfs-s3:8333" .Release.Name }}
{{- else -}}
{{ $os.external.endpoint }}
{{- end -}}
{{- end -}}

{{- define "infra.s3PublicEndpoint" -}}
{{- $os := .Values.global.objectStorage -}}
{{- if eq $os.mode "in-cluster" -}}
{{ printf "%s://s3.%s" .Values.global.scheme .Values.global.host }}
{{- else -}}
{{ $os.external.publicEndpoint }}
{{- end -}}
{{- end -}}
