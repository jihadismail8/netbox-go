FROM alpine:latest
MAINTAINER zhufuyi "g.zhufuyi@gmail.com"

# add grpc_health_probe for health check of grpc services
COPY grpc_health_probe /bin/grpc_health_probe
RUN chmod +x /bin/grpc_health_probe

COPY configs/ /app/configs/
COPY netbox_go /app/netbox_go
RUN chmod +x /app/netbox_go

# grpc and http port
EXPOSE 8282 8283


WORKDIR /app

CMD ["./netbox_go", "-c", "configs/netbox_go.yml"]
# if you use the Configuration Center, netbox_go.yml is changed to the Configuration Center configuration.
#CMD ["./netbox_go", "-c", "configs/netbox_go.yml", "-enable-cc"]
