#Start with the base go image
FROM golang:1.13 AS builder

#Copy the code and run the build
WORKDIR /app 
COPY * ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o hive_poller .

#Build the actual image
FROM alpine
RUN apk --no-cache add ca-certificates
COPY --from=builder /hive_poller .
CMD ["/hive_poller"]