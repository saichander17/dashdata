use std::io::{Error, ErrorKind};

#[derive(Debug, Clone, PartialEq)]
pub enum RespValue {
    SimpleString(String),
    Error(String),
    Integer(i64),
    BulkString(Option<Vec<u8>>),
    Array(Vec<RespValue>),
}

impl RespValue {
    pub fn serialize(&self) -> Vec<u8> {
        match self {
            RespValue::SimpleString(s) => format!("+{}\r\n", s).into_bytes(),
            RespValue::Error(s) => format!("-{}\r\n", s).into_bytes(),
            RespValue::Integer(i) => format!(":{}\r\n", i).into_bytes(),
            RespValue::BulkString(Some(data)) => {
                let mut result = format!("${}\r\n", data.len()).into_bytes();
                result.extend(data);
                result.extend(b"\r\n");
                result
            }
            RespValue::BulkString(None) => b"$-1\r\n".to_vec(),
            RespValue::Array(arr) => {
                let mut result = format!("*{}\r\n", arr.len()).into_bytes();
                for value in arr {
                    result.extend(value.serialize());
                }
                result
            }
        }
    }

    pub fn parse(input: &[u8]) -> Result<(RespValue, usize), Error> {
        match input.first() {
            Some(b'*') => {
                let mut pos = 1;
                let count = parse_integer(&input[pos..])?;
                pos += find_crlf(&input[pos..])? + 2;
                let mut values = Vec::with_capacity(count as usize);
                for _ in 0..count {
                    let (value, consumed) = RespValue::parse(&input[pos..])?;
                    values.push(value);
                    pos += consumed;
                }
                Ok((RespValue::Array(values), pos))
            },
            Some(b'$') => {
                let mut pos = 1;
                let len = parse_integer(&input[pos..])?;
                pos += find_crlf(&input[pos..])? + 2;
                if len == -1 {
                    Ok((RespValue::BulkString(None), pos))
                } else {
                    let end = pos + len as usize;
                    let data = input[pos..end].to_vec();
                    Ok((RespValue::BulkString(Some(data)), end + 2))
                }
            },
            Some(b'+') => parse_simple_string(input),
            Some(b'-') => parse_error(input),
            Some(b':') => parse_integer_value(input),
            _ => Err(Error::new(ErrorKind::InvalidData, "Invalid RESP data")),
        }
    }
}

fn unescape(s: &str) -> String {
    s.replace("\\r", "\r").replace("\\n", "\n")
}

fn parse_integer(input: &[u8]) -> Result<i64, Error> {
    let end = find_crlf(input)?;
    let num_str = std::str::from_utf8(&input[..end])
        .map_err(|_| Error::new(ErrorKind::InvalidData, "Invalid UTF-8 in integer"))?;
    let unescaped = unescape(num_str.trim());
    unescaped.parse::<i64>()
        .map_err(|e| Error::new(ErrorKind::InvalidData, format!("Invalid integer: {}", e)))
}

fn find_crlf(input: &[u8]) -> Result<usize, Error> {
    input.iter().position(|&b| b == b'\n')
        .map(|pos| if pos > 0 && input[pos - 1] == b'\r' { pos - 1 } else { pos })
        .ok_or_else(|| Error::new(ErrorKind::InvalidData, "Line ending not found"))
}

fn parse_simple_string(input: &[u8]) -> Result<(RespValue, usize), Error> {
    let end = find_crlf(input)?;
    let s = String::from_utf8_lossy(&input[1..end]).to_string();
    Ok((RespValue::SimpleString(s), end + 2))
}

fn parse_error(input: &[u8]) -> Result<(RespValue, usize), Error> {
    let end = find_crlf(input)?;
    let s = String::from_utf8_lossy(&input[1..end]).to_string();
    Ok((RespValue::Error(s), end + 2))
}

fn parse_integer_value(input: &[u8]) -> Result<(RespValue, usize), Error> {
    let end = find_crlf(input)?;
    let num = parse_integer(&input[1..end + 1])?;
    Ok((RespValue::Integer(num), end + 2))
}
