{{ define "upstream_server" -}}
  {{- $labels   := .Container.Config.Labels -}}
  {{- $netSett  := .Container.NetworkSettings -}}
  {{- $status   := coalesce .Container.State.Status "down" -}}
  {{- $hostname := index $labels "service.web.backend.hostname" -}}
  {{- $hostport := $netSett | network_first_hostport -}}
  {{- $port     := coalesce (index $labels "service.web.backend.port") .Port -}}

  {{- if $hostname -}}
    server {{ $hostname }}:{{ $port }}{{ if (ne $status "running") }} down{{end}};
  {{- else -}}
    {{- $networkName := index $labels "service.web.docker.network" -}}
    {{- if $networkName -}}
      {{- $network := index $netSett.Networks $networkName -}}
      {{- if $network -}}
        server {{ $network.IPAddress }}:{{ $port }}{{ if (ne $status "running") }} down{{end}};
      {{- end -}}
    {{- else -}}
      server {{ $netSett.IPAddress }}:{{ $port }}{{ if (ne $status "running") }} down{{end}};
    {{- end -}}
  {{- end -}}
{{- end }}

{{ range $i, $it := .allcontainers -}}

  {{- $labels       := $it.Config.Labels -}}
  {{- $service_name := coalesce (index $labels "service.name") (index $labels "com.docker.swarm.service.name") -}}
  {{- $basic        := index $labels "service.web.frontend.auth.basic" -}}
  {{- $basicTitle   := coalesce (index $labels "service.web.frontend.auth.basic_title") "Administrator’s area" -}}
  {{- $hostname     := coalesce (index $labels "service.web.frontend.hostname") (printf "%s.localhost" $service_name) -}}


  {{- if (eq (index $labels "service.web.enable") "true") -}}
  upstream {{ $service_name }} {
    {{- template "upstream_server" (dict "Container" $it) }}
  }

  {{ if false }}
  server {
    listen 80;
    server_name {{ $hostname }};
    return 301 https://$host$request_uri
  }
  {{ end }}

  server {
    listen {{ indexor $labels "service.web.frontend.port" "80" }};
    server_name {{ $hostname }};

    location / {
      proxy_http_version 1.1;
      proxy_pass "http://{{ $service_name }}";
      proxy_set_header Connection $connection_upgrade;
      proxy_set_header Upgrade $http_upgrade;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header X-Forwarded-Proto $scheme;
      proxy_set_header Proxy $proxy_host;

      {{ if $basic -}}
      auth_basic           "{{ $basicTitle }}";
      auth_basic_user_file /etc/nginx/vhost.d/.{{ $service_name }}.htpasswd; 
      {{- end }}
    }

    location /.well-known/ {
      root /var/www/html/;
    }
  }
  {{- end }}
{{- end }}