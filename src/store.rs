use std::collections::HashMap;
use std::sync::{Arc, RwLock};

pub struct Store {
    data: Arc<RwLock<HashMap<String, String>>>,
}

impl Store {
    pub fn new() -> Self {
        Store {
            data: Arc::new(RwLock::new(HashMap::new())),
        }
    }

    pub fn get(&self, key: &str) -> Option<String> {
        self.data.read().unwrap().get(key).cloned()
    }

    pub fn set(&self, key: String, value: String) {
        self.data.write().unwrap().insert(key, value);
    }

    pub fn delete(&self, key: &str) {
        self.data.write().unwrap().remove(key);
    }
}
