{{/*
prompt.server renders a backend component: Deployment + Service + HTTPRoute
(on its apiPath, no strip) + optional rate limit + CNPG Database.
dict keys: root, name (phase base, e.g. "core"), comp (.Values.server)
*/}}
{{- define "prompt.server" -}}
{{- $root := .root -}}
{{- $comp := .comp -}}
{{- $suffix := printf "%s-server" .name -}}
{{- $svc := include "prompt.name" (dict "root" $root "suffix" $suffix) -}}
{{- if $comp.enabled }}
{{ include "prompt.workload" (dict "root" $root "suffix" $suffix "component" $suffix "comp" $comp "kind" "backend" "dbKey" $comp.dbKey) }}
---
{{ include "prompt.httproute" (dict "root" $root "suffix" $suffix "component" $suffix "path" $comp.apiPath "strip" false "serviceName" $svc "servicePort" $comp.port) }}
{{- if $comp.rateLimited }}
---
{{ include "prompt.ratelimit" (dict "root" $root "suffix" $suffix "component" $suffix "routeName" $svc) }}
{{- end }}
{{- $db := include "prompt.database" (dict "root" $root "dbKey" $comp.dbKey) }}
{{- if trim $db }}
---
{{ $db }}
{{- end }}
{{- end }}
{{- end -}}

{{/*
prompt.client renders a frontend component: Deployment + Service + HTTPRoute
(on its path, with optional prefix strip).
dict keys: root, name (phase base), comp (.Values.client)
*/}}
{{- define "prompt.client" -}}
{{- $root := .root -}}
{{- $comp := .comp -}}
{{- $suffix := printf "%s-client" .name -}}
{{- $svc := include "prompt.name" (dict "root" $root "suffix" $suffix) -}}
{{- if $comp.enabled }}
{{ include "prompt.workload" (dict "root" $root "suffix" $suffix "component" $suffix "comp" $comp "kind" "frontend") }}
---
{{ include "prompt.httproute" (dict "root" $root "suffix" $suffix "component" $suffix "path" $comp.path "strip" $comp.stripPrefix "serviceName" $svc "servicePort" $comp.port) }}
{{- end }}
{{- end -}}
