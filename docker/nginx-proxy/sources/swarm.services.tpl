{{ define "upstream_server_service" -}}
  {{- $labels   := .Service.Service.Spec.Labels -}}
  {{- $network  := .Service.Service.Spec.TaskTemplate.Networks | first -}}
  {{- $port     := index $labels "service.web.backend.port" -}}

  server {{ $network.Aliases | first }}{{ if $port }}:{{ $port }}{{ end }};
{{- end }}

{{ range $i, $it := .allservices -}}

  {{- $srv            := $it.Service -}}
  {{- $labels         := $srv.Spec.Labels -}}
  {{- $service_name   := coalesce (index $labels "service.name") (index $labels "com.docker.swarm.service.name") $srv.Spec.Name -}}
  {{- $basic          := index $labels "service.web.frontend.auth.basic" -}}
  {{- $basicTitle     := coalesce (index $labels "service.web.frontend.auth.basic_title") "Administrator’s area" -}}
  {{- $hostname       := coalesce (index $labels "service.web.frontend.hostname") (printf "%s.localhost" $service_name) -}}
  {{- $max_body_size  := index $labels "service.web.frontend.client_max_body_size" -}}

  {{- if (eq (index $labels "service.web.enable") "true") }}

  {{ if gt $it.LiveCount 0 -}}
  upstream {{ $service_name }} {
    {{ template "upstream_server_service" (dict "Service" $it) }}
  }
  {{- end }}

  server {
    listen {{ indexor $labels "service.web.frontend.port" "80" }};
    server_name {{ $hostname }};

    location / {
      # HTTP 1.1 support
      proxy_http_version 1.1;
      proxy_buffering off;
      proxy_set_header Host $http_host;
      proxy_set_header Upgrade $http_upgrade;
      proxy_set_header Connection $proxy_connection;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header X-Forwarded-Proto $proxy_x_forwarded_proto;
      proxy_set_header X-Forwarded-Ssl $proxy_x_forwarded_ssl;
      proxy_set_header X-Forwarded-Port $proxy_x_forwarded_port;

      {{ if eq $it.LiveCount 0 -}}
      root   /usr/share/nginx/html;
      index  index.html index.htm;
      {{- else -}}
      proxy_pass http://{{ $service_name }};
      {{- end }}

      {{ if $max_body_size -}}
      client_max_body_size {{ $max_body_size }};
      {{- end }}

      {{ if $basic -}}
      auth_basic           "{{ $basicTitle }}";
      auth_basic_user_file /etc/nginx/vhost.d/.{{ $service_name }}.htpasswd; 
      {{- end }}
    }

    {{ if eq $it.LiveCount 0 -}}
    # redirect server error pages to the static page /50x.html
    #
    error_page   500 502 503 504  /50x.html;
    location = /50x.html {
        root   /usr/share/nginx/html;
    }
    {{- end }}

    location /.well-known/ {
      root /var/www/html/;
    }
  }

  {{- end -}}
{{- end }}