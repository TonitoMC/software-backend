#!/bin/sh

host="$1"
port="$2"
shift 2

echo "Esperando a que la base de datos esté lista en $host:$port..."

until pg_isready -h "$host" -p "$port" -U postgres; do
  >&2 echo "La base de datos no está disponible aún — esperando..."
  sleep 1
done

echo "La base de datos está lista — ejecutando comando..."
exec "$@"