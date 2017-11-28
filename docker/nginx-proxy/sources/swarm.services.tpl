{{ define "upstream_server_service" -}}
  {{- $labels   := .Service.Spec.Labels -}}
  {{- $network  := .Service.Spec.TaskTemplate.Networks | first -}}
  {{- $port     := index $labels "service.web.backend.port" -}}

  server {{ $network.Aliases | first }}{{ if $port }}:{{ $port }}{{ end }};
{{- end }}

{{ range $i, $it := .allservices -}}

  {{- $labels       := $it.Spec.Labels -}}
  {{- $service_name := coalesce (index $labels "service.name") (index $labels "com.docker.swarm.service.name") $it.Spec.Name -}}
  {{- $basic        := index $labels "service.web.frontend.auth.basic" -}}
  {{- $basicTitle   := coalesce (index $labels "service.web.frontend.auth.basic_title") "Administrator’s area" -}}
  {{- $hostname     := coalesce (index $labels "service.web.frontend.hostname") (printf "%s.localhost" $service_name) -}}

  {{- if (eq (index $labels "service.web.enable") "true") }}
  upstream {{ $service_name }} {
    {{ template "upstream_server_service" (dict "Service" $it) }}
  }

  server {
    listen {{ indexor $labels "service.web.frontend.port" "80" }};
    server_name {{ $hostname }};

    location / {
      proxy_http_version 1.1;
      proxy_pass "http://{{ $service_name }}";
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
  {{- end -}}
{{- end }}