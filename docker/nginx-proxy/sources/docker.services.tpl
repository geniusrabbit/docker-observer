{{ define "upstream_server" -}}
  {{- $labels   := .Container.Config.Labels -}}
  {{- $netSett  := .Container.NetworkSettings -}}
  {{- $status   := coalesce .Container.State.Status "down" -}}
  {{- $hostname := index $labels "service.web.backend.hostname" -}}
  {{- $hostport := .Container.FirstPort -}}
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

  {{- $labels         := $it.Config.Labels -}}
  {{- $service_name   := coalesce (index $labels "service.name") (index $labels "com.docker.swarm.service.name") -}}
  {{- $basic          := index $labels "service.web.frontend.auth.basic" -}}
  {{- $basicTitle     := coalesce (index $labels "service.web.frontend.auth.basic_title") "Administrator’s area" -}}
  {{- $hostname       := coalesce (index $labels "service.web.frontend.hostname") (printf "%s.localhost" $service_name) -}}
  {{- $aliases        := split (index $labels "service.web.frontend.hostalias") "," -}}
  {{- $max_body_size  := index $labels "service.web.frontend.client_max_body_size" -}}

  {{- if (eq (index $labels "service.web.enable") "true") -}}
  upstream {{ $service_name }} {
    {{- template "upstream_server" (dict "Container" $it) }}
  }

  {{ if false }}
  server {
    listen 80;
    server_name {{ $hostname }} {{ join $aliases " " }};
    return 301 https://$host$request_uri
  }
  {{ end }}

  server {
    listen {{ indexor $labels "service.web.frontend.port" "80" }};
    server_name {{ $hostname }} {{ join $aliases " " }};

    {{ if eq $service_name "registry" -}}
    # required to avoid HTTP 411: see Issue #1486 (https://github.com/moby/moby/issues/1486)
    chunked_transfer_encoding on;

    location /v2/ {
      # Do not allow connections from docker 1.5 and earlier
      # docker pre-1.6.0 did not properly set the user agent on ping, catch "Go *" user agents
      if ($http_user_agent ~ "^(docker\/1\.(3|4|5(?!\.[0-9]-dev))|Go ).*$" ) {
        return 404;
      }

      {% if registry_login is defined and registry_login != '' %}
      # To add basic authentication to v2 use auth_basic setting.
      auth_basic "Access can";
      auth_basic_user_file /etc/nginx/conf.d/docker-registry.htpasswd;
      {% endif %}

      ## If $docker_distribution_api_version is empty, the header will not be added.
      ## See the map directive above where this variable is defined.
      add_header 'Docker-Distribution-Api-Version' '{{ registry_api_version }}' always;

      proxy_pass                          http://127.0.0.1:{{ registry_port }};
      proxy_set_header  Host              $http_host;   # required for docker client's sake
      proxy_set_header  X-Real-IP         $remote_addr; # pass on real client's IP
      proxy_set_header  X-Forwarded-For   $proxy_add_x_forwarded_for;
      proxy_set_header  X-Forwarded-Proto $scheme;
      proxy_read_timeout                  900;
    }
    {{- else -}}
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

      {{ if $max_body_size -}}
      client_max_body_size {{ $max_body_size }};
      {{- end }}

      {{ if $basic -}}
      auth_basic           "{{ $basicTitle }}";
      auth_basic_user_file /etc/nginx/vhost.d/.{{ $service_name }}.htpasswd; 
      {{- end }}
    }

    location /.well-known/ {
      root /var/www/html/;
    }
    {{- end }}
  }
  {{- end }}
{{- end }}