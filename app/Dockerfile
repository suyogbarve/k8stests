FROM alpine:latest

ADD ./greeter_server /app/greeter_server

WORKDIR /app

CMD ["./greeter_server"]
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
EXPOSE 50051
