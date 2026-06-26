{{/*
prompt.database renders a CloudNativePG Database CRD for one phase, owned by the
phase but referencing the shared cluster + role (created in infrastructure).
Rendered only when postgresql.mode == in-cluster.
dict keys: root, dbKey
*/}}
{{- define "prompt.database" -}}
{{- $root := .root -}}
{{- $pg := $root.Values.global.postgresql -}}
{{- if eq $pg.mode "in-cluster" -}}
{{- $db := index $pg.databases .dbKey -}}
apiVersion: postgresql.cnpg.io/v1
kind: Database
metadata:
  name: {{ include "prompt.name" (dict "root" $root "suffix" (printf "db-%s" .dbKey)) }}
  labels:
    {{- include "prompt.labels" (dict "root" $root "component" (printf "db-%s" .dbKey)) | nindent 4 }}
spec:
  cluster:
    name: {{ $pg.clusterName }}
  name: {{ $db.dbName }}
  owner: {{ $db.owner }}
  ensure: present
  databaseReclaimPolicy: retain
{{- end -}}
{{- end -}}

{{/*
prompt.ratelimit renders an Envoy Gateway BackendTrafficPolicy attaching a local
rate limit to a component's HTTPRoute. Gated on the configured provider.
dict keys: root, suffix, component, routeName
*/}}
{{- define "prompt.ratelimit" -}}
{{- $root := .root -}}
{{- $rl := $root.Values.global.gateway.rateLimiting -}}
{{- if eq $rl.provider "envoygateway" -}}
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: BackendTrafficPolicy
metadata:
  name: {{ include "prompt.name" (dict "root" $root "suffix" (printf "%s-ratelimit" .suffix)) }}
  labels:
    {{- include "prompt.labels" (dict "root" $root "component" .component) | nindent 4 }}
spec:
  targetRefs:
    - group: gateway.networking.k8s.io
      kind: HTTPRoute
      name: {{ .routeName }}
  rateLimit:
    type: Local
    local:
      rules:
        - limit:
            requests: {{ $rl.average }}
            unit: Second
{{- end -}}
{{- end -}}
