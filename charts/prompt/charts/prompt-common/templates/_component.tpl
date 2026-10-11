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
{{- $backend := eq .kind "backend" -}}
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
      annotations:
        checksum/appconfig: {{ include "prompt.appConfigData" $root | sha256sum }}
        checksum/appsecrets: {{ $g.appSecrets | toYaml | sha256sum }}
        {{- if $backend }}
        checksum/db: {{ include "prompt.dbSecretInputs" (dict "root" $root "dbKey" .dbKey) | sha256sum }}
        {{- end }}
    spec:
      {{- with $g.image.pullSecrets }}
      imagePullSecrets:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      securityContext:
        seccompProfile:
          type: RuntimeDefault
        {{- if $backend }}
        {{- /* The Go images build on distroless static (root variant, no USER), so the
               UID has to come from here or the kubelet refuses runAsNonRoot. */}}
        runAsNonRoot: true
        runAsUser: {{ $g.podSecurity.runAsUser }}
        runAsGroup: {{ $g.podSecurity.runAsGroup }}
        {{- end }}
      {{- if $backend }}
      initContainers:
        - name: wait-for-db
          image: {{ $g.dbWaitImage }}
          command: ["sh", "-ec"]
          args:
            - |
              until psql -qtAX -c 'select 1' >/dev/null 2>&1; do
                echo "waiting for database $PGDATABASE at $PGHOST:$PGPORT..."; sleep 2;
              done
          env:
            - name: PGHOST
              value: {{ include "prompt.dbHost" (dict "root" $root) | quote }}
            - name: PGPORT
              value: {{ include "prompt.dbPort" (dict "root" $root) | quote }}
            - name: PGSSLMODE
              value: {{ include "prompt.dbSslMode" (dict "root" $root) | quote }}
            - name: PGUSER
              valueFrom:
                secretKeyRef:
                  name: {{ include "prompt.dbSecretName" (dict "root" $root "dbKey" .dbKey) }}
                  key: DB_USER
            - name: PGPASSWORD
              valueFrom:
                secretKeyRef:
                  name: {{ include "prompt.dbSecretName" (dict "root" $root "dbKey" .dbKey) }}
                  key: DB_PASSWORD
            - name: PGDATABASE
              valueFrom:
                secretKeyRef:
                  name: {{ include "prompt.dbSecretName" (dict "root" $root "dbKey" .dbKey) }}
                  key: DB_NAME
            - name: HOME
              value: /tmp
          resources:
            requests:
              cpu: 10m
              memory: 32Mi
            limits:
              memory: 64Mi
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop: ["ALL"]
            runAsNonRoot: true
            readOnlyRootFilesystem: true
          volumeMounts:
            - name: tmp
              mountPath: /tmp
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
            {{- if $backend }}
            - secretRef:
                name: {{ include "prompt.appSecretName" $root }}
            - secretRef:
                name: {{ include "prompt.dbSecretName" (dict "root" $root "dbKey" .dbKey) }}
            {{- end }}
          {{- if $backend }}
          env:
            - name: HOME
              value: /tmp
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
            {{- if $backend }}
            capabilities:
              drop: ["ALL"]
            runAsNonRoot: true
            readOnlyRootFilesystem: true
            {{- else }}
            {{- /* Frontends run stock nginx as root: it binds :80 and drops privileges
                   itself, so it needs NET_BIND_SERVICE, SETUID/SETGID and DAC_OVERRIDE.
                   Dropping capabilities here requires an unprivileged nginx base. */}}
            {{- end }}
          {{- if $backend }}
          volumeMounts:
            - name: tmp
              mountPath: /tmp
          {{- end }}
      {{- if $backend }}
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
