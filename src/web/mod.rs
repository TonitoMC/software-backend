use axum::{
    extract::{Json, State},
    http::StatusCode,
    response::IntoResponse,
    routing::post,
    Router,
};
use bcrypt::{hash, verify, DEFAULT_COST};
use jsonwebtoken::{encode, EncodingKey, Header};
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tokio::sync::Mutex;
use chrono::{Utc, Duration};
use dotenv::dotenv;
use std::env;

use crate::model::{LoginRequest, LoginResponse};

#[derive(Serialize)]
struct Claims {
    sub: String,
    exp: usize,
}

#[derive(Deserialize)]
struct RegisterRequest {
    email: String,
    password: String,
}

#[derive(Serialize)]
struct ErrorResponse {
    error: String,
}

#[derive(Serialize)]
#[serde(untagged)]
enum ApiResponse {
    Success(LoginResponse),
    Error(ErrorResponse),
}

pub fn create_routes(
    users_db: Arc<Mutex<std::collections::HashMap<String, String>>>,
) -> Router {
    Router::new()
        .route("/login", post(login_handler))
        .route("/register", post(register_handler))
        .with_state(users_db)
}

async fn login_handler(
    State(users_db): State<Arc<Mutex<std::collections::HashMap<String, String>>>>,
    Json(payload): Json<LoginRequest>,
) -> impl IntoResponse {
    dotenv().ok();
    let secret_key = env::var("JWT_SECRET").expect("JWT_SECRET debe estar configurado");

    let users_db = users_db.lock().await;

    if let Some(stored_password) = users_db.get(&payload.email) {
        if verify(&payload.password, stored_password).unwrap() {
            let claims = Claims {
                sub: payload.email.clone(),
                exp: (Utc::now() + Duration::hours(1)).timestamp() as usize,
            };

            let token = match encode(
                &Header::default(),
                &claims,
                &EncodingKey::from_secret(secret_key.as_ref()),
            ) {
                Ok(t) => t,
                Err(_) => {
                    return (
                        StatusCode::INTERNAL_SERVER_ERROR,
                        Json(ApiResponse::Error(ErrorResponse {
                            error: "Error al generar el token".to_string(),
                        })),
                    );
                }
            };

            let response = LoginResponse { token };
            return (StatusCode::OK, Json(ApiResponse::Success(response)));
        }
    }

    (
        StatusCode::UNAUTHORIZED,
        Json(ApiResponse::Error(ErrorResponse {
            error: "Credenciales inválidas".to_string(),
        })),
    )
}

async fn register_handler(
    State(users_db): State<Arc<Mutex<std::collections::HashMap<String, String>>>>,
    Json(payload): Json<RegisterRequest>,
) -> impl IntoResponse {
    let mut users_db = users_db.lock().await;

    if payload.email.is_empty() || !payload.email.contains('@') {
        return (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::Error(ErrorResponse {
                error: "El correo electrónico no es válido".to_string(),
            })),
        );
    }

    if payload.password.len() < 6 {
        return (
            StatusCode::BAD_REQUEST,
            Json(ApiResponse::Error(ErrorResponse {
                error: "La contraseña debe tener al menos 6 caracteres".to_string(),
            })),
        );
    }

    if users_db.contains_key(&payload.email) {
        return (
            StatusCode::CONFLICT,
            Json(ApiResponse::Error(ErrorResponse {
                error: "El usuario ya existe".to_string(),
            })),
        );
    }

    let hashed_password = hash(&payload.password, DEFAULT_COST).unwrap();
    users_db.insert(payload.email.clone(), hashed_password);

    (
        StatusCode::CREATED,
        Json(ApiResponse::Success(LoginResponse {
            token: "Usuario registrado exitosamente".to_string(),
        })),
    )
}