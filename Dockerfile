FROM golang:1.8
USER root
RUN go get google.golang.org/grpc

WORKDIR /go/src/google.golang.org/grpc/examples/helloworld

RUN go install ./greeter_server

ADD ./app/Dockerfile /app/Dockerfile

RUN cp /go/bin/greeter_server /app/greeter_server

WORKDIR /app

RUN ls
RUN pwd


#CMD ls
CMD tar cvzf - ./*
