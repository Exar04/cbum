FROM golang AS builder

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM scratch 

COPY --from=builder /app/main .

EXPOSE 9000 

CMD ["./main"]