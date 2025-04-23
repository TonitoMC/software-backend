# Etapa 1: compilación
FROM rust:1.86.0 as builder
WORKDIR /usr/src/software-backend
COPY . .
RUN cargo build --release

# Etapa 2: imagen final liviana
FROM debian:buster-slim

# Instalamos dependencias necesarias para ejecutar el binario
RUN apt-get update && apt-get install -y libpq-dev postgresql-client && apt-get clean


WORKDIR /app

# Copiamos el binario compilado
COPY --from=builder /usr/src/software-backend/target/release/software-backend .

# Añadimos script para esperar a la base de datos
COPY wait-for-db.sh /usr/local/bin/wait-for-db.sh
RUN chmod +x /usr/local/bin/wait-for-db.sh

EXPOSE 4000

ENV DATABASE_URL=postgres://postgres:admin123@db:5432/oftalcrm

# Llamamos el script que espera y luego ejecuta el backend
CMD ["wait-for-db.sh", "db:5432", "--", "./software-backend"]
