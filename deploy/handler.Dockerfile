FROM golang:1.22.5 as build

WORKDIR /usr/src/app

COPY . .

RUN go build -o ./main ./cmd/handler/main.go


FROM ubuntu

WORKDIR /usr/app
RUN mkdir "config"

COPY --from=build /usr/src/app/main ./main
COPY --from=build /usr/src/app/config/base.yaml ./config/base.yaml

CMD ["./main"]