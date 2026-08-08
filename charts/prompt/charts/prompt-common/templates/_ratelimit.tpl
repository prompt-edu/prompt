{{/*
prompt.ratelimit renders an Envoy Gateway BackendTrafficPolicy attaching a local
rate limit to a component's HTTPRoute. Needs both the Envoy provider (the CRD
only exists there) and the rate-limiting toggle.
dict keys: root, suffix, component, routeName
*/}}
{{- define "prompt.ratelimit" -}}
{{- $root := .root -}}
{{- $gw := $root.Values.global.gateway -}}
{{- if and (eq $gw.provider "envoygateway") $gw.rateLimiting.enabled -}}
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
            requests: {{ $gw.rateLimiting.average }}
            unit: Second
{{- end -}}
{{- end -}}
