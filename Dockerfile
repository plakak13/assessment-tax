# Build stage
FROM golang:1.21 as build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -o ./out/assessment-tax .

#Final stage
FROM alpine:3.19.1
COPY --from=build app/out/assessment-tax /app/assessment-tax

CMD [ "/app/assessment-tax" ]