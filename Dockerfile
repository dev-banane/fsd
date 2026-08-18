FROM debian:bookworm-slim AS build

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        cmake \
        g++ \
        make \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src
COPY CMakeLists.txt ./
COPY fsd ./fsd

RUN cmake -S . -B build -DCMAKE_BUILD_TYPE=Release \
    && cmake --build build --parallel \
    && strip build/fsd

FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        curl \
        gosu \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --system fsd \
    && useradd --system --gid fsd --home-dir /data --create-home \
        --shell /usr/sbin/nologin fsd \
    && mkdir -p /opt/fsd

COPY --from=build /src/build/fsd /usr/local/bin/fsd
COPY docker/fsd.conf /opt/fsd/fsd.conf
COPY unix/motd.txt /opt/fsd/motd.txt
COPY unix/adminhelp.txt /opt/fsd/help.txt
COPY docker/entrypoint.sh /usr/local/bin/entrypoint.sh

RUN chmod 755 /usr/local/bin/fsd /usr/local/bin/entrypoint.sh \
    && sed -i 's/\r$//' /usr/local/bin/entrypoint.sh

ENV TZ=UTC \
    FSD_HOME=/data \
    FSD_CLIENTPORT=6809

WORKDIR /data
VOLUME ["/data"]

# Client connections
EXPOSE 6809/tcp
# System management
EXPOSE 3010/tcp
# Server-to-server FSD links
EXPOSE 3011/tcp

STOPSIGNAL SIGTERM
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
