FROM golang:1.15 AS build

RUN apt-get -qq update && apt-get -yqq install upx

ENV GO111MODULE on
ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

WORKDIR /src

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY . .
RUN go build \
  -a \
  -trimpath \
  -ldflags "-s -w -extldflags '-static'" \
  -o /bin/moneysaver \
  .

RUN strip /bin/moneysaver
RUN upx -q -9 /bin/moneysaver


FROM scratch

ENV PORT 8080

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /bin/moneysaver /bin/moneysaver

ENTRYPOINT ["/bin/moneysaver"]
