FROM rancher/agent-base:v0.3.0

RUN apt-get update && apt-get install -y \
    curl \
    vim-tiny \
    iputils-ping \
    netcat \
    software-properties-common && \
    add-apt-repository ppa:vbernat/haproxy-1.6 && \
    apt-get update && apt-get install -y haproxy

RUN mkdir -p /etc/haproxy/
COPY healthcheck /usr/bin/
COPY haproxy_reload.sh /usr/bin
COPY haproxy_template.cfg /etc/haproxy/haproxy_template.cfg

ENV TINI_VERSION v0.10.0
ADD https://github.com/krallin/tini/releases/download/${TINI_VERSION}/tini /tini
RUN chmod +x /tini
ENTRYPOINT ["/rancher-entrypoint.sh", "/tini", "--"]

CMD ["healthcheck"]
