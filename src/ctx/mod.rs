use std::collections::HashMap;

pub fn init_users_db() -> HashMap<String, String> {
    let mut users_db = HashMap::new();
    users_db.insert("usuario@ejemplo.com".to_string(), "password123".to_string());
    users_db
}