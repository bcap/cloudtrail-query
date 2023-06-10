FROM alpine as build

# install base build tools
RUN apk add bash build-base go 

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download -x

COPY . .

RUN go build -o cloudtrail-query ./cmd

#
# final exported image
#
FROM alpine
WORKDIR /app
COPY --from=build /app/cloudtrail-query cloudtrail-query

VOLUME [ "/root/.aws" ]

ENTRYPOINT [ "/app/cloudtrail-query" ]