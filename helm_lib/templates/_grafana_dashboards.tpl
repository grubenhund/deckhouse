{{- /* Usage: {{ include "helm_lib_grafana_dashboard_definitions_recursion" (list . <root dir> [current dir]) }} */ -}}
{{- /* returns dashboard-definintions from <root dir>/ */ -}}
{{- /* current dir is optional — used for recursion but you can use it for partially generating dashboards */ -}}
{{- define "helm_lib_grafana_dashboard_definitions_recursion" -}}
  {{- $context := index . 0 }}
  {{- $rootDir := index . 1 }}
  {{- $currentDir := "" }}
  {{- if gt (len .) 2 }} {{- $currentDir = index . 2 }} {{- else }} {{- $currentDir = $rootDir }} {{- end }}
  {{- $currentDirIndex := (sub ($currentDir | splitList "/" | len) 1) }}
  {{- $rootDirIndex := (sub ($rootDir | splitList "/" | len) 1) }}
  {{- $folderNamesIndex := (add1 $rootDirIndex) }}

  {{- range $path, $_ := $context.Files.Glob (print $currentDir "/*.json") }}
    {{- $fileName := ($path | splitList "/" | last ) }}
    {{- $definition := ($context.Files.Get $path) }}

    {{- $folder := (index ($currentDir | splitList "/") $folderNamesIndex | replace "-" " " | title) }}
    {{- $resourceName := (regexReplaceAllLiteral "\\.json$" $path "") }}
    {{- $resourceName = ($resourceName | replace " " "-" | replace "." "-" | replace "_" "-") }}
    {{- $resourceName = (slice ($resourceName | splitList "/") $folderNamesIndex | join "-") }}
    {{- $resourceName = (printf "%s-%s" $context.Chart.Name $resourceName) }}
---
apiVersion: deckhouse.io/v1alpha1
kind: GrafanaDashboardDefinition
metadata:
  name: {{ $resourceName }}
{{ include "helm_lib_module_labels" (list $context (dict "prometheus.deckhouse.io/grafana-dashboard" "")) | indent 2 }}
spec:
  folder: "{{ $folder }}"
  definition: |
{{ $definition | indent 4 }}

  {{- end }}

  {{- $subDirs := list }}
  {{- range $path, $_ := ($context.Files.Glob (print $currentDir "/**.json")) }}
    {{- $pathSlice := ($path | splitList "/") }}
    {{- $subDirs = append $subDirs (slice $pathSlice 0 (add $currentDirIndex 2) | join "/") }}
  {{- end }}

  {{- range $subDir := ($subDirs | uniq) }}
{{ include "helm_lib_grafana_dashboard_definitions_recursion" (list $context $rootDir $subDir) }}
  {{- end }}
{{- end }}


{{- /* Usage: {{ include "helm_lib_grafana_dashboard_definitions" . }} */ -}}
{{- /* returns dashboard-definintions from monitoring/grafana-dashboards/ */ -}}
{{- define "helm_lib_grafana_dashboard_definitions" -}}
  {{- $context := . }}
  {{- include "helm_lib_grafana_dashboard_definitions_recursion" (list $context "monitoring/grafana-dashboards") }}
{{- end }}
