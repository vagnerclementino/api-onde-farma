FROM golang:1.25 AS builder
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='-s -w' -o /out/bootstrap ./cmd/lambda

FROM public.ecr.aws/lambda/provided:al2023
COPY --from=builder /out/bootstrap /var/runtime/bootstrap
CMD ["bootstrap"]
