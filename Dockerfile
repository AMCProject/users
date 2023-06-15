FROM golang:1.19 as builder

WORKDIR /src/app
ADD . /src/app

RUN go build -ldflags="-extldflags=-static" -mod=mod -o /bin/app cmd/user/main.go

FROM gcr.io/distroless/base-debian11
COPY --from=builder /bin/app /bin/app
COPY internal/config/.env ./internal/config/.env

CMD ["/bin/app"]