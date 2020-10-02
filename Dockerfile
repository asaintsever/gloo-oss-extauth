FROM golang:1.14.4 AS build

COPY . /gloo-oss-extauth
RUN cd /gloo-oss-extauth && make build OFFLINE=true

FROM centos:7.8.2003

WORKDIR /opt

COPY --from=build /gloo-oss-extauth/target/gloo-oss-extauth /opt/

ENV PATH=$PATH:/opt
ENTRYPOINT ["gloo-oss-extauth"]
