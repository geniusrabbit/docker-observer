FROM nginx

LABEL maintainer="GeniusRabbit Dmitry Ponomarev: Load balancer"
LABEL service.role="balancer"

VOLUME ["/etc/nginx/certs", "/etc/nginx/dhparam", "/etc/nginx/vhost.d"]

EXPOSE 80 443

RUN echo "include /etc/nginx/vhost.d/*.conf;" > /etc/nginx/conf.d/dynamic.conf

COPY ./nginx.conf /etc/nginx/
COPY ./network_internal.conf /etc/nginx/
COPY ./docker-entrypoint.sh /app/docker-entrypoint.sh
COPY ./.build/docker-observer /app/docker-observer
COPY ./sources /app/sources

ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["/app/docker-observer", "--basedir", "/app/sources", "--config", "config.json", "--debug", "1"]
