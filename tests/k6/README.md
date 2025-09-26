# Pruebas de rendimiento y seguridad básicas con k6

Esta carpeta contiene scripts de k6 para ejecutar pruebas de carga, estrés y comprobaciones rápidas de seguridad contra la API del backend. El entorno objetivo es GitHub Codespaces, pero también funcionan de forma local en Linux/macOS.

## 0) Prerrequisitos de Backend y BD

- URL base del backend: http://localhost:4000 (por defecto). Puedes sobrescribirla con la variable de entorno BASE_URL.
- Levanta el stack antes de correr las pruebas:
  - Docker: usa el docker-compose del repositorio para iniciar backend, base de datos y MinIO.
  - Asegúrate de que JWT_SECRET esté configurada en el backend (el compose lo hace en dev).
  - La base de datos se inicializa con `database/backup.sql`.

## 1) Instalar k6 en Codespaces

Desde la raíz del repositorio, en la terminal del Codespace:

- `software-backend/scripts/install-k6.sh`

Este script agrega el repositorio apt de Grafana e instala k6.

Verifica la instalación con:
- `k6 version`

## 2) Endpoints confirmados (según software-backend/internal/api/routes.go)

- GET `/healthz`
- POST `/login`
- POST `/register`
- GET `/appointments?start_time&end_time` (timestamps RFC3339)
- GET `/appointments/today`
- GET `/appointments/month?year&month`
- GET `appointments/day?date=YYYY-MM-DD` (nota: en el código falta la barra inicial; para fiabilidad, usa `/appointments/today` o el rango `/appointments`).
- DELETE `/appointments/:id`
- PUT `/appointments/:id`
- POST `/appointments`
- GET `/patients/search?q=`
- GET `/patients/:id`
- GET `/business-hours`
- GET `/consultations/patient/:patient_id`
- GET `/patients/:patientId/exams`
- POST `/exams/:examId/upload`
- GET `/exams/:examId/download`
- GET `/exams/pending`
- GET `/consultations/:consultation_id/diagnostics`
- POST `/consultations/:consultation_id/diagnostics`
- CRUD en `/api/consultations`
- GET `/api/questionnaires`
- GET `/api/questionnaires/:id/questions`

Observación: actualmente no se aplica un middleware JWT de forma global; varios endpoints podrían ser accesibles públicamente. Esto se comenta en la sección de Seguridad más abajo.

## 3) Puntos de entrada de pruebas

Todas las pruebas aceptan variables de entorno:
- `BASE_URL`: por defecto `http://localhost:4000`.
- `USERNAME`, `PASSWORD`: opcionales para el flujo de login; los tests también pueden auto‑registrar un usuario.

Ejemplos de ejecución (desde la raíz del repo):

- `k6 run --summary-export=summary.json --out json=results.json software-backend/tests/k6/smoke-test.js`
- `k6 run --summary-export=summary.json --out json=results.json software-backend/tests/k6/load-test.js`
- `k6 run --summary-export=summary.json --out json=results.json software-backend/tests/k6/stress-test.js`
- `k6 run --summary-export=summary.json --out json=results.json software-backend/tests/k6/security-probes.js`

Consulta `software-backend/scripts/run-k6.sh` para atajos de ejecución.

## 4) Escenarios

- `smoke-test.js`: verificación rápida (health, business-hours, appointments/today, search).
- `load-test.js`: carga sostenida con rampa por etapas (10 → 25 → 50 VUs) centrada en lecturas.
- `stress-test.js`: rampa agresiva para encontrar el punto de quiebre; monitoriza tasa de error y latencia.
- `security-probes.js`: comprobaciones básicas de SQLi, XSS, CORS y endpoints que deberían requerir autenticación.

## 5) Métricas y umbrales

Se definen umbrales para:
- `http_req_failed`: < 1% (carga), < 5% (estrés)
- `http_req_duration` p(95): < 400 ms (carga), < 1000 ms (estrés)
- Checks (status 2xx): > 95%

Ajusta estos valores según el entorno.

## 6) Resultados y reporting

- Cada ejecución puede exportar:
  - `--summary-export=summary.json`: resumen de alto nivel
  - `--out json=results.json`: muestras por iteración

Para visualizar:
- Usa `jq` para explorar JSON
- Extensiones de k6 o paneles de Grafana si se permite k6 Cloud.

Además, los scripts generan automáticamente `summary.html` y `summary.txt` (ver `summary.js`).

## 7) Seguridad: notas y checklist

- Revisa el control de acceso: en `main.go` no se aplica globalmente el middleware JWT. Define grupos protegidos si corresponde.
- Valida CORS en producción: actualmente permite solo `http://localhost:8080`. Para Codespaces u otros dominios, actualiza `AllowOrigins`.
- Prueba entradas para SQLi y XSS (ver `security-probes.js`). El servidor debería responder con 4xx y sin trazas internas.
- Verifica que los endpoints que deban requerir autenticación realmente la exijan (prueba con y sin `Authorization`).

## 8) Resolución de problemas (Troubleshooting)

- Errores 500 bajo estrés pueden indicar límites de conexiones a BD: ajusta el pool o añade índices.
- Si `/patients/:id` devuelve 404 en las pruebas, primero usa `/patients/search` para obtener un ID válido.
- MinIO gestiona archivos de exámenes; por defecto evitamos subir/descargar en estos tests, a menos que lo habilites explícitamente.

## 9) Siguientes pasos

- Añadir un job en CI (GitHub Actions) para ejecutar smoke/load de forma nocturna y subir artefactos.
- Incorporar pruebas de “soak” (1–2 horas de carga sostenida) para evaluar estabilidad.
- Conectar k6 a InfluxDB + Grafana para dashboards más ricos si es necesario.
