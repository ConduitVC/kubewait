FROM busybox
WORKDIR /bin
COPY kubewait /bin/
ENTRYPOINT ["./kubewait"]
