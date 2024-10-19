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

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_set_and_get() {
        let store = Store::new();
        store.set("key1".to_string(), "value1".to_string());
        assert_eq!(store.get("key1"), Some("value1".to_string()));
    }

    #[test]
    fn test_get_non_existent() {
        let store = Store::new();
        assert_eq!(store.get("non_existent"), None);
    }

    #[test]
    fn test_del() {
        let store = Store::new();
        store.set("key2".to_string(), "value2".to_string());
        assert_eq!(store.get("key2"), Some("value2".to_string()));
        store.delete("key2");
        assert_eq!(store.get("key2"), None);
    }

    #[test]
    fn test_empty_key() {
        // ToDo: We should return an error for empty_key just as a safety net
        let store = Store::new();
        store.set("".to_string(), "empty_key".to_string());
        assert_eq!(store.get(""), Some("empty_key".to_string()));
    }

    #[test]
    fn test_large_value() {
        // ToDo: We should have a limit on the value and key sizes. But why do we need them? Think!!
        let store = Store::new();
        let large_value = "a".repeat(1_000_000);
        store.set("large_key".to_string(), large_value.clone());
        assert_eq!(store.get("large_key"), Some(large_value));
    }

    #[test]
    fn test_concurrent_operations() {
        use std::sync::Arc;
        use std::thread;

        let store = Arc::new(Store::new());
        let threads: Vec<_> = (0..10)
            .map(|i| {
                let store = Arc::clone(&store);
                thread::spawn(move || {
                    store.set(format!("key{}", i), format!("value{}", i));
                    assert_eq!(store.get(&format!("key{}", i)), Some(format!("value{}", i)));
                })
            })
            .collect();

        for thread in threads {
            thread.join().unwrap();
        }
    }

    #[test]#[test]
    fn test_consistency() {
        let store = Store::new();
        for i in 0..100 {
           store.set(format!("key{}", i), format!("value{}", i));
        }
        for i in 0..100 {
           assert_eq!(store.get(&format!("key{}", i)), Some(format!("value{}", i)));
        }
        for i in 0..50 {
           store.delete(&format!("key{}", i));
        }
        for i in 0..100 {
           if i < 50 {
               assert_eq!(store.get(&format!("key{}", i)), None);
           } else {
               assert_eq!(store.get(&format!("key{}", i)), Some(format!("value{}", i)));
           }
        }
    }

    #[test]
    fn test_performance() {
        // ToDo: Determine the accurate number of operations per second based on hardware?
        use std::time::Instant;
        let store = Store::new();
        let start = Instant::now();
        // This number is currently based on my macbook under normal usage
        let number_of_operations = 900_000;
        for i in 0..number_of_operations {
           store.set(format!("key{}", i), format!("value{}", i));
        }
        let duration = start.elapsed();
        println!("Time taken to insert {:?} items: {:?}", number_of_operations, duration);
        assert!(duration.as_secs() < 1, "Insertion took too long");
    }
}
