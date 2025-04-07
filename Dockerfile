FROM rust:1.86.0 as builder

WORKDIR /usr/src/software-backend
COPY . .

RUN cargo build --release

EXPOSE 3000
CMD ["./target/release/software-backend"]


