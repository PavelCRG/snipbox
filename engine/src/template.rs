use once_cell::sync::Lazy;
use regex::Regex;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

static VAR: Lazy<Regex> =
    Lazy::new(|| Regex::new(r"\{\{([a-zA-Z_][a-zA-Z0-9_]*)(?::([^{}]*))?\}\}").unwrap());
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
        dangerous: is_dangerous(command),
    })
    .map_err(|e| e.to_string())
}

fn is_dangerous(command: &str) -> bool {
    // shortcut: shell-neutral tokenization misses language-specific syntax, upgrade when shell metadata is available.
    command_segments(command)
        .iter()
        .any(|segment| is_dangerous_segment(segment))
}

fn command_segments(command: &str) -> Vec<Vec<String>> {
    let mut result = Vec::new();
    let mut tokens = Vec::new();
    let mut token = String::new();
    let mut quote = None;
    let mut chars = command.chars().peekable();
    while let Some(ch) = chars.next() {
        if let Some(delimiter) = quote {
            if ch == delimiter {
                quote = None;
            } else if ch == '\\' && chars.peek() == Some(&delimiter) {
                if let Some(escaped) = chars.next() {
                    token.push(escaped);
                }
            } else {
                token.push(ch);
            }
            continue;
        }
        match ch {
            '\'' | '"' => quote = Some(ch),
            '$' if chars.peek() == Some(&'(') => {
                chars.next();
                if !token.is_empty() {
                    tokens.push(std::mem::take(&mut token));
                }
                if !tokens.is_empty() {
                    result.push(std::mem::take(&mut tokens));
                }
            }
            ')' => {
                if !token.is_empty() {
                    tokens.push(std::mem::take(&mut token));
                }
                if !tokens.is_empty() {
                    result.push(std::mem::take(&mut tokens));
                }
            }
            '#' if token.is_empty() => {
                while let Some(comment) = chars.next() {
                    if comment == '\n' {
                        break;
                    }
                }
                if !tokens.is_empty() {
                    result.push(std::mem::take(&mut tokens));
                }
            }
            ';' | '|' | '&' | '\n' => {
                if !token.is_empty() {
                    tokens.push(std::mem::take(&mut token));
                }
                if !tokens.is_empty() {
                    result.push(std::mem::take(&mut tokens));
                }
            }
            ch if ch.is_whitespace() => {
                if !token.is_empty() {
                    tokens.push(std::mem::take(&mut token));
                }
            }
            _ => token.push(ch),
        }
    }
    if !token.is_empty() {
        tokens.push(token);
    }
    if !tokens.is_empty() {
        result.push(tokens);
    }
    result
}

fn is_dangerous_segment(tokens: &[String]) -> bool {
    if tokens.is_empty() {
        return false;
    }
    let mut command = 0;
    while command < tokens.len()
        && matches!(
            tokens[command].to_ascii_lowercase().as_str(),
            "sudo" | "env" | "command" | "exec" | "nohup" | "time"
        )
    {
        command += 1;
    }
    while command < tokens.len()
        && tokens[command].contains('=')
        && !tokens[command].starts_with('-')
    {
        command += 1;
    }
    let Some(executable) = tokens.get(command) else {
        return false;
    };
    let executable = executable
        .rsplit(['/', '\\'])
        .next()
        .unwrap_or(executable)
        .to_ascii_lowercase();
    let executable = executable.strip_suffix(".exe").unwrap_or(&executable);
    let args = &tokens[command + 1..];
    match executable {
        "rm" => {
            let mut recursive = false;
            let mut force = false;
            for arg in args {
                let arg = arg.to_ascii_lowercase();
                recursive |= arg == "--recursive"
                    || (arg.starts_with('-') && !arg.starts_with("--") && arg.contains('r'));
                force |= arg == "--force"
                    || (arg.starts_with('-') && !arg.starts_with("--") && arg.contains('f'));
            }
            recursive && force
        }
        "git"
            if args
                .first()
                .is_some_and(|arg| arg.eq_ignore_ascii_case("reset")) =>
        {
            args.iter().any(|arg| arg.eq_ignore_ascii_case("--hard"))
        }
        "git"
            if args
                .first()
                .is_some_and(|arg| arg.eq_ignore_ascii_case("clean")) =>
        {
            args.iter()
                .any(|arg| arg.starts_with('-') && arg[1..].contains('f'))
        }
        "mkfs" => true,
        _ if executable.starts_with("mkfs.") => true,
        "format" | "format.com" => args.first().is_some_and(|arg| {
            arg.as_bytes().get(1) == Some(&b':') && arg.as_bytes()[0].is_ascii_alphabetic()
        }),
        "drop" => args.first().is_some_and(|arg| {
            matches!(
                arg.to_ascii_lowercase().as_str(),
                "table" | "database" | "schema"
            )
        }),
        "remove-item" => true,
        "clear-content" => true,
        "stop-process" => args.iter().any(|arg| arg.eq_ignore_ascii_case("-force")),
        "docker" => match args.first().map(|arg| arg.to_ascii_lowercase()).as_deref() {
            Some("compose") => {
                args.iter().any(|arg| arg.eq_ignore_ascii_case("down"))
                    && args
                        .iter()
                        .any(|arg| matches!(arg.as_str(), "-v" | "--volumes"))
            }
            Some("system" | "image" | "container" | "volume") => {
                args.iter().any(|arg| arg.eq_ignore_ascii_case("prune"))
            }
            _ => false,
        },
        "kubectl" => args
            .first()
            .is_some_and(|arg| arg.eq_ignore_ascii_case("delete")),
        _ => false,
    }
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

    #[test]
    fn flags_destructive_commands_without_matching_quoted_examples() {
        for command in [
            "rm -rf /tmp/cache",
            "rm -fr /tmp/cache",
            "rm --recursive --force /tmp/cache",
            "git reset --hard HEAD",
            "git clean -fdx",
            "mkfs.ext4 /dev/sdb1",
            "format D:",
            "DROP TABLE sessions",
            "Remove-Item C:\\temp -Recurse -Force",
            "Remove-Item C:\\temp",
            "sudo rm -rf /tmp/cache",
            "command rm -rf /tmp/cache",
            "echo $(rm -rf /tmp/cache)",
            "docker compose down -v",
        ] {
            let parsed: Parsed = serde_json::from_str(&parse(command).unwrap()).unwrap();
            assert!(parsed.dangerous, "expected a warning for: {command}");
        }

        for command in [
            "echo \"rm -rf /tmp/cache\"",
            "git reset --soft HEAD",
            "echo \"drop table sessions\"",
            "Write-Output 'format is a command library'",
        ] {
            let parsed: Parsed = serde_json::from_str(&parse(command).unwrap()).unwrap();
            assert!(!parsed.dangerous, "unexpected warning for: {command}");
        }

        let rendered: String = serde_json::from_str(
            &render("echo {{value}}", r#"{"value":"safe; rm -rf /tmp/cache"}"#).unwrap(),
        )
        .unwrap();
        let parsed: Parsed = serde_json::from_str(&parse(&rendered).unwrap()).unwrap();
        assert!(
            parsed.dangerous,
            "rendered template should be checked after substitution"
        );
    }
}
