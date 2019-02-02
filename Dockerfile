FROM golang:1.10
RUN mkdir /go/src/gateway -p
ADD . /go/src/gateway/
WORKDIR /go/src/gateway 
#RUN go get "gateway/lib"
#RUN go get "github.com/rs/xid"
RUN go get "github.com/Jeffail/gabs"
RUN go get github.com/gorilla/websocket
#RUN cp -r ./lib /go/src/gateway/lib 
#RUN go get "github.com/jinzhu/gorm"
#RUN go get "github.com/go-sql-driver/mysql"
RUN go build -o pool-api-gateway .

ENTRYPOINT ["/go/src/gateway/pool-api-gateway", "./pool.config.json"]
