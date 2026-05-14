FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/chats-svc ./cmd

FROM alpine:3.22

COPY --from=build /out/chats-svc /usr/local/bin/chats-svc

EXPOSE 9992

ENTRYPOINT ["chats-svc"]
