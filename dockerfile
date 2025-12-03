###########################
# Builder Stage
###########################
FROM golang:1.24.9-alpine AS builder

ARG SERVICE

WORKDIR /activeOperations

COPY go.mod go.sum ./
RUN go env -w GOPROXY="https://goproxy.io,https://proxy.golang.org,https://goproxy.cn,direct"
RUN go mod download

COPY . .

# 复制 agent 特有资源到 builder（只复制一次）
RUN mkdir -p /source
COPY bash /source/bash

# 编译指定微服务
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/${SERVICE}


###########################
# Runtime Stage
###########################
FROM alpine:latest

ARG SERVICE

WORKDIR /activeOperations

COPY --from=builder /activeOperations/app .
COPY ./config.yaml ./config.yaml

COPY --from=builder /source /source

# 只有 agent 服务才需要文档与 bash 资源
RUN if [ "$SERVICE" = "agent" ]; then \
        apk add --no-cache git bash && \
        cp -r /source/bash /activeOperations/bash ; \
    fi

EXPOSE 8080

CMD ["./app"]
