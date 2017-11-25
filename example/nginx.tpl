{{ define "upstream_server" -}}
  {{- $labels   := .Container.Config.Labels -}}
  {{- $hostname := index $labels "service.web.backend.hostname" -}}
  {{- $port     := coalesce (index $labels "service.web.backend.port") .Port -}}

  {{- if $hostname -}}
    server {{ $hostname }}:{{ $port }}{{ if (ne .Status "running") }} down{{end}};
  {{- else -}}
    {{- $networkName := index $labels "service.web.docker.network" -}}
    {{- if $networkName -}}
      {{ $network := index .NetworkSettings.Networks $networkName }}
      {{ if $network -}}
        server {{ $network.IPAddress }}:{{ $port }}{{ if (ne .Status "running") }} down{{end}};
      {{- end -}}
    {{- else -}}
      server {{ .NetworkSettings.IPAddress }}:{{ $port }}{{ if (ne .Status "running") }} down{{end}};
    {{- end -}}
  {{- end }}
{{- end }}

{{ range $i, $c := .allcontainers -}}

  {{- $labels       := $c.Config.Labels -}}
  {{- $service_name := coalesce (index $labels "service.name") (index $labels "com.docker.swarm.service.name") -}}

  {{- if (eq (index $labels "service.web.enable") "true") -}}
  upstream {{ $service_name }} {
    {{ $port   := $c.NetworkSettings | network_first_hostport }}
    {{ $status := coalesce $c.State.Status "down" }}
    {{ template "upstream_server" (dict "Container" $c "Port" $port "NetworkSettings" $c.NetworkSettings "Status" $status) }}
  }

  {{ if false }}
  server {
    listen 80;
    server_name: {{ coalesce (index $labels "service.web.frontend.hostname") (printf "%s.localhost" $service_name) }};
    return: 301 https://$host$request_uri
  }
  {{ end }}

  server {
    listen: {{ indexor $labels "service.web.frontend.port" "80" }};
    server_name: {{ coalesce (index $labels "service.web.frontend.hostname") (printf "%s.localhost" $service_name) }};

    localtion / {
      proxy_http_version 1.1;
      proxy_pass "http://{{ $service_name }}";
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

  {{ $c | jsonbeauty }}

  {{- end }}
{{- end }}