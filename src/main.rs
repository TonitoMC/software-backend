use axum::{Router};
use axum::http::Method;
use std::sync::Arc;
use tokio::sync::Mutex;
use tower_http::cors::{Any, CorsLayer};
use hyper::header::HeaderValue; // Importar HeaderValue
use sqlx::PgPool;
use dotenv::dotenv;
use std::env;

mod ctx; // Módulo para la base de datos simulada
mod web; // Módulo para las rutas web
mod model; // Módulo para las estructuras de datos

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    dotenv().ok();

    let database_url = env::var("DATABASE_URL").expect("DATABASE_URL no está configurado");
    let pool = PgPool::connect(&database_url).await?;

    println!("Conexión a la base de datos exitosa");

    let users_db = Arc::new(Mutex::new(ctx::init_users_db())); // ctx es el módulo para la base de datos

    let cors = CorsLayer::new()
        .allow_origin("http://localhost:8080".parse::<HeaderValue>().unwrap()) // Permitir solicitudes desde el frontend
        .allow_methods([Method::GET, Method::POST])
        .allow_headers(Any);

    let app = Router::new()
        .merge(web::create_routes(users_db.clone()))
        .layer(cors);

    let listener = tokio::net::TcpListener::bind("0.0.0.0:4000").await.unwrap();
    println!("Server running on http://localhost:4000");
    axum::serve(listener, app).await.unwrap();

    Ok(())
}