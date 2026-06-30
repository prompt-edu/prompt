{{/*
prompt.workload renders a Deployment + Service for one component.
dict keys:
  root       subchart root context (.)
  suffix     resource name suffix, e.g. "core-server"
  component  label app name, e.g. "core-server"
  comp       component values map (image, replicas, port, ...)
  kind       "backend" | "frontend"
  dbKey      (backend) key into global.postgresql.databases for the DB secret
*/}}
{{- define "prompt.workload" -}}
{{- $root := .root -}}
{{- $comp := .comp -}}
{{- $name := include "prompt.name" (dict "root" $root "suffix" .suffix) -}}
{{- $g := $root.Values.global -}}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ $name }}
  labels:
    {{- include "prompt.labels" (dict "root" $root "component" .component) | nindent 4 }}
spec:
  replicas: {{ $comp.replicas | default 1 }}
  selector:
    matchLabels:
      {{- include "prompt.selectorLabels" (dict "root" $root "component" .component) | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "prompt.selectorLabels" (dict "root" $root "component" .component) | nindent 8 }}
    spec:
      {{- with $g.image.pullSecrets }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- if eq .kind "backend" }}
      initContainers:
        - name: wait-for-db
          image: busybox:1.36
          command:
            - sh
            - -c
            - |
              until nc -z {{ include "prompt.dbWaitHost" (dict "root" $root) }} {{ include "prompt.dbWaitPort" (dict "root" $root) }}; do
                echo "waiting for database..."; sleep 2;
              done
      {{- end }}
      containers:
        - name: {{ .component }}
          image: {{ include "prompt.image" (dict "root" $root "image" $comp.image "tag" $comp.tag) }}
          imagePullPolicy: {{ $g.image.pullPolicy }}
          ports:
            - name: http
              containerPort: {{ $comp.port }}
          envFrom:
            - configMapRef:
                name: {{ include "prompt.appConfigName" $root }}
            {{- if eq .kind "backend" }}
            - secretRef:
                name: {{ include "prompt.appSecretName" $root }}
            - secretRef:
                name: {{ include "prompt.dbSecretName" (dict "root" $root "dbKey" .dbKey) }}
            {{- end }}
          {{- if eq .kind "backend" }}
          startupProbe:
            tcpSocket:
              port: http
            periodSeconds: 10
            failureThreshold: 60
          livenessProbe:
            tcpSocket:
              port: http
            periodSeconds: 15
          readinessProbe:
            tcpSocket:
              port: http
            periodSeconds: 10
          {{- else }}
          startupProbe:
            httpGet:
              path: /
              port: http
            periodSeconds: 5
            failureThreshold: 24
          livenessProbe:
            httpGet:
              path: /
              port: http
            periodSeconds: 20
          readinessProbe:
            httpGet:
              path: /
              port: http
            periodSeconds: 10
          {{- end }}
          resources:
            {{- toYaml ($comp.resources | default (dict "requests" (dict "cpu" "50m" "memory" "64Mi") "limits" (dict "memory" "256Mi"))) | nindent 12 }}
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop: ["ALL"]
            {{- if eq .kind "backend" }}
            runAsNonRoot: true
            readOnlyRootFilesystem: true
            {{- end }}
          {{- if and (eq .kind "backend") }}
          volumeMounts:
            - name: tmp
              mountPath: /tmp
          {{- end }}
      {{- if eq .kind "backend" }}
      volumes:
        - name: tmp
          emptyDir: {}
      {{- end }}
---
apiVersion: v1
kind: Service
metadata:
  name: {{ $name }}
  labels:
    {{- include "prompt.labels" (dict "root" $root "component" .component) | nindent 4 }}
spec:
  type: ClusterIP
  selector:
    {{- include "prompt.selectorLabels" (dict "root" $root "component" .component) | nindent 4 }}
  ports:
    - name: http
      port: {{ $comp.port }}
      targetPort: http
{{- end -}}
