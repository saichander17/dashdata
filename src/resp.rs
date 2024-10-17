use std::str;

#[derive(Debug, Clone)]
pub enum RespValue {
    SimpleString(String),
    Error(String),
    Integer(i64),
    BulkString(String),
    Array(Vec<RespValue>),
    Null,
}

impl RespValue {
    pub fn parse(input: &[u8]) -> Result<Self, &'static str> {
        println!("Parsing input: {:?}", input);
        match input.get(0) {
            Some(b'+') => {
                println!("+Matched");
                let s = str::from_utf8(&input[1..input.len()-2]).map_err(|_| "Invalid UTF-8")?;
                Ok(RespValue::SimpleString(s.to_string()))
            },
            Some(b'-') => {
                println!("-Matched");
                let s = str::from_utf8(&input[1..input.len()-2]).map_err(|_| "Invalid UTF-8")?;
                Ok(RespValue::Error(s.to_string()))
            },
            Some(b':') => {
                println!(":Matched");
                let s = str::from_utf8(&input[1..input.len()-2]).map_err(|_| "Invalid UTF-8")?;
                let i = s.parse::<i64>().map_err(|_| "Invalid integer")?;
                Ok(RespValue::Integer(i))
            },
            Some(b'$') => {
                println!("$Matched");
                if &input[1..3] == b"-1" {
                    Ok(RespValue::Null)
                } else {
                    let len_end = input.iter().position(|&c| c == b'\r').ok_or("Invalid bulk string")?;
                    let len = str::from_utf8(&input[1..len_end]).map_err(|_| "Invalid UTF-8")?;
                    let len = len.parse::<usize>().map_err(|_| "Invalid length")?;
                    let s = str::from_utf8(&input[len_end+2..len_end+2+len]).map_err(|_| "Invalid UTF-8")?;
                    Ok(RespValue::BulkString(s.to_string()))
                }
            },
            Some(b'*') => {
                let len_end = input.iter().position(|&c| c == b'\r').ok_or("Invalid array")?;
                let len = str::from_utf8(&input[1..len_end]).map_err(|_| "Invalid UTF-8")?;
                let len = len.parse::<usize>().map_err(|_| "Invalid length")?;
                let mut values = Vec::new();
                let mut pos = len_end + 2;
                for _ in 0..len {
                    let value = RespValue::parse(&input[pos..])?;
                    values.push(value);
                    pos += input[pos..].iter().position(|&c| c == b'\r').ok_or("Invalid array")? + 2;
                }
                Ok(RespValue::Array(values))
            },
            _ => {
                println!("Nothing Matched");
                Err("Invalid RESP value")
            },
        }
    }

    pub fn serialize(&self) -> Vec<u8> {
        match self {
            RespValue::SimpleString(s) => format!("+{}\r\n", s).into_bytes(),
            RespValue::Error(s) => format!("-{}\r\n", s).into_bytes(),
            RespValue::Integer(i) => format!(":{}\r\n", i).into_bytes(),
            RespValue::BulkString(s) => format!("${}\r\n{}\r\n", s.len(), s).into_bytes(),
            RespValue::Array(values) => {
                let mut result = format!("*{}\r\n", values.len()).into_bytes();
                for value in values {
                    result.extend(value.serialize());
                }
                result
            },
            RespValue::Null => "$-1\r\n".as_bytes().to_vec(),
        }
    }
}
