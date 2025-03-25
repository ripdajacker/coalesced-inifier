FROM golang:1.24

WORKDIR /app

COPY . ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /coalesced-inifier

EXPOSE 8080
CMD ["/coalesced-inifier", "web"]
