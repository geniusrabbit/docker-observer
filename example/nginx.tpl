{{ define "upstream_server" -}}
  {{ if (index .Container.Config.Labels "service.hostname") -}}
    server {{ (index .Container.Config.Labels "service.hostname") }}:{{ .Port.HostPort }}{{ if (ne .Status "running") }} down{{end}};
  {{- else -}}
    server {{ .NetworkSettings.IPAddress }}:{{ .Port.HostPort }}{{ if (ne .Status "running") }} down{{end}};
  {{- end }}
{{- end }}

{{ range $i, $c := .allcontainers -}}
  {{ if (and (eq (index $c.Config.Labels "service.public") "true") (and $c.NetworkSettings.IPAddress (gt (len $c.NetworkSettings.Ports) 0))) -}}
  upstream {{ index $c.Config.Labels "service.name" }} {
    {{ range $k, $port := $c.NetworkSettings.Ports -}}
      {{- if $port -}}
        {{ $hostPort := $port | first }}
        {{ if $hostPort -}}
          {{ $status := coalesce $c.State.Status "down" }}
          {{ template "upstream_server" (dict "Container" $c "Port" $hostPort "NetworkSettings" $c.NetworkSettings "Status" $status) }}
        {{- end }}
      {{- end -}}
    {{- end }}
  }

  {{ if false }}
  server {
    listen 80;
    {{ if (index $c.Config.Labels "service.web.frontend.hostname") -}}
    server_name: {{ index $c.Config.Labels "service.web.frontend.hostname" }};
    {{- end }}
    return: 301 https://$host$request_uri
  }
  {{ end }}

  server {
    listen: {{ indexor $c.Config.Labels "service.web.port" "80" }};
    {{ if (index $c.Config.Labels "service.web.frontend.hostname") -}}
    server_name: {{ index $c.Config.Labels "service.web.frontend.hostname" }};
    {{- end }}

    localtion / {
      proxy_http_version 1.1;
      proxy_pass "http://{{ index $c.Config.Labels "service.name" }}";
      proxy_set_header Connection $connection_upgrade;
      proxy_set_header Upgrade $http_upgrade;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header X-Forwarded-Proto $scheme;
      proxy_set_header Proxy "";
    }

    location /.well-known/ {
      root: /var/www/html/;
    }
  }
  {{- end }}
{{- end }}