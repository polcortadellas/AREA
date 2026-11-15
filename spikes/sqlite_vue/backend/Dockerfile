FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /spike .

FROM alpine:3.20
RUN adduser -D app && mkdir /data && chown app /data
USER app
ENV DB_PATH=/data/spike.db
COPY --from=build /spike /spike
EXPOSE 8080
CMD ["/spike"]
