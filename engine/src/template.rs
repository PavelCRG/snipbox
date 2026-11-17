use once_cell::sync::Lazy;
use regex::Regex;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

static VAR: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"\{\{([a-zA-Z_][a-zA-Z0-9_]*)(?::([^{}]*))?\}\}").unwrap());
static DANGEROUS: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"(?i)(rm\s+-rf|reset\s+--hard|\bdrop\b|\bformat\b|\bmkfs)").unwrap());

#[derive(Serialize, Deserialize)]
struct Variable {
    name: String,
    #[serde(default)]
    default: String,
}
#[derive(Serialize, Deserialize)]
struct Parsed {
    variables: Vec<Variable>,
    dangerous: bool,
}

pub fn parse(command: &str) -> Result<String, String> {
    let variables = VAR
        .captures_iter(command)
        .map(|c| Variable {
            name: c[1].to_string(),
            default: c.get(2).map_or("", |m| m.as_str()).to_string(),
        })
        .collect();
    serde_json::to_string(&Parsed {
        variables,
        dangerous: DANGEROUS.is_match(command),
    })
    .map_err(|e| e.to_string())
}

pub fn render(command: &str, values_json: &str) -> Result<String, String> {
    let values: HashMap<String, String> =
        serde_json::from_str(values_json).map_err(|e| e.to_string())?;
    let rendered = VAR.replace_all(command, |c: &regex::Captures| {
        values
            .get(&c[1])
            .cloned()
            .or_else(|| c.get(2).map(|m| m.as_str().to_string()))
            .unwrap_or_default()
    });
    serde_json::to_string(rendered.as_ref()).map_err(|e| e.to_string())
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn parses_defaults_and_detects_danger() {
        let p: Parsed = serde_json::from_str(&parse("rm -rf {{path:/tmp}}").unwrap()).unwrap();
        assert!(p.dangerous);
        assert_eq!(p.variables[0].default, "/tmp");
    }
    #[test]
    fn renders_value_or_default() {
        assert_eq!(
            render("echo {{name:world}}", r#"{"name":"snipbox"}"#).unwrap(),
            "\"echo snipbox\""
        );
    }
}
