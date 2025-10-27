FROM debian:bookworm-slim

WORKDIR /app
COPY backend .
EXPOSE 8080
CMD ["./backend"]