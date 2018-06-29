FROM golang:1.8
WORKDIR /go/src/gateway
COPY . . 
RUN go-wrapper download
RUN go-wrapper install
#CMD ["/go/src/app/main"]
#EXPOSE 3000
ENTRYPOINT ["go-wrapper", "run","./config.json"]
