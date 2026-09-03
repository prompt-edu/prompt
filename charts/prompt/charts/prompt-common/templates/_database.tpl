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
