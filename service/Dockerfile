FROM    golang:1.19-bullseye as builder
WORKDIR /app
COPY    . ./
RUN     go build -o service

FROM    debian:buster-slim
# PORT=443 used as a default.  
# Can override via GCE Environment Variables, etc.
ENV		PORT=443
RUN     set -x && \
		apt-get update && \
		DEBIAN_FRONTEND=noninteractive apt-get install -y \
			ca-certificates && \
			rm -rf /var/lib/apt/lists/*
COPY    --from=builder /app/service /app/service

CMD     ["/app/service"]