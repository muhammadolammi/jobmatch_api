FROM debian:bookworm-slim
# Install CA certificates so Go can verify TLS
RUN apt-get update && apt-get install -y ca-certificates
WORKDIR /app
COPY backend .
EXPOSE 8080
CMD ["./backend"]