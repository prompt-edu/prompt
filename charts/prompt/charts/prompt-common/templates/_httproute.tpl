{{/*
prompt.httproute renders a Gateway API HTTPRoute attached to the apex listener.
dict keys:
  root        subchart root context (.)
  suffix      resource name suffix
  component   label app name
  path        path prefix to match (e.g. /assessment or /assessment/api)
  strip       bool - rewrite the matched prefix to "/" (client micro-frontends)
  serviceName backend Service name
  servicePort backend Service port
*/}}
{{- define "prompt.httproute" -}}
{{- $root := .root -}}
{{- $g := $root.Values.global -}}
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: {{ include "prompt.name" (dict "root" $root "suffix" .suffix) }}
  labels:
    {{- include "prompt.labels" (dict "root" $root "component" .component) | nindent 4 }}
spec:
  parentRefs:
    - name: {{ $g.gateway.name }}
      sectionName: apex
  hostnames:
    - {{ $g.host | quote }}
  rules:
    - matches:
        - path:
            type: PathPrefix
            value: {{ .path | quote }}
      {{- if .strip }}
      filters:
        - type: URLRewrite
          urlRewrite:
            path:
              type: ReplacePrefixMatch
              replacePrefixMatch: /
      {{- end }}
      backendRefs:
        - name: {{ .serviceName }}
          port: {{ .servicePort }}
{{- end -}}
