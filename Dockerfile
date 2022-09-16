FROM mtr.devops.telekom.de/mcsps/golang:1.18 as builder

WORKDIR /app
ADD . /app

RUN go mod download && go mod tidy && go build -o main main.go

FROM gcr.io/distroless/base
LABEL org.opencontainers.image.authors="f.kloeker@telekom.de"
LABEL version="1.0.0"
LABEL description="Create DemoApp for OTC RDS Operator"

WORKDIR /

USER nonroot:nonroot
COPY --from=builder --chown=nonroot:nonroot /app/main /
COPY --from=builder --chown=nonroot:nonroot /app/kodata /var/run/ko

ENV KO_DATA_PATH=/var/run/ko
EXPOSE 8080
ENTRYPOINT ["/main"]
