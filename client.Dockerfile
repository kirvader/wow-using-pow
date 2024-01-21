FROM golang:1.19

WORKDIR /

COPY ./ .

RUN go mod download

RUN go build -mod=mod -o start_client ./cmd/client/*.go

CMD [ "./start_client" ]