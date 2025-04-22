// filepath: d:\Escritorio\Ing Software PRY\CRM\software-backend\src\model\mod.rs
use serde::{Deserialize, Serialize};

#[derive(Deserialize)]
pub struct LoginRequest {
    pub email: String,
    pub password: String,
}

#[derive(Serialize)]
pub struct LoginResponse {
    pub token: String,
}