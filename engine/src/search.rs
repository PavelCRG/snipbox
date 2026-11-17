use fuzzy_matcher::{skim::SkimMatcherV2, FuzzyMatcher};
use serde::{Deserialize, Serialize};

#[derive(Deserialize)]
struct Command {
    name: String,
    #[serde(default)]
    description: String,
    #[serde(default)]
    tags: Option<Vec<String>>,
    #[serde(default)]
    command: String,
    #[serde(default)]
    is_favorite: bool,
    #[serde(default)]
    usage_count: u64,
}
#[derive(Serialize, Deserialize)]
struct Match {
    index: usize,
    score: i64,
    indices: Vec<usize>,
}

pub fn search(json: &str, query: &str) -> Result<String, String> {
    let commands: Vec<Command> = serde_json::from_str(json).map_err(|e| e.to_string())?;
    let matcher = SkimMatcherV2::default();
    let mut found: Vec<(usize, i64, bool, u64, Vec<usize>)> = commands
        .iter()
        .enumerate()
        .filter_map(|(i, c)| {
            if query.is_empty() {
                return Some((i, 0, c.is_favorite, c.usage_count, vec![]));
            }
            let fields = [&c.name, &c.description, &c.command];
            let best = fields
                .into_iter()
                .chain(c.tags.iter().flatten())
                .filter_map(|s| matcher.fuzzy_indices(s, query))
                .max_by_key(|(score, _)| *score);
            let (score, _) = best?;
            let name_hits = matcher
                .fuzzy_indices(&c.name, query)
                .map(|(_, v)| v)
                .unwrap_or_default();
            Some((i, score, c.is_favorite, c.usage_count, name_hits))
        })
        .collect();
    found.sort_by(|a, b| b.2.cmp(&a.2).then(b.1.cmp(&a.1)).then(b.3.cmp(&a.3)));
    serde_json::to_string(
        &found
            .into_iter()
            .map(|(index, score, _, _, indices)| Match {
                index,
                score,
                indices,
            })
            .collect::<Vec<_>>(),
    )
    .map_err(|e| e.to_string())
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn ranks_favorites_and_returns_highlight_positions() {
        let json = r#"[{"name":"git status","is_favorite":false},{"name":"git stash","is_favorite":true}]"#;
        let results: Vec<Match> = serde_json::from_str(&search(json, "git").unwrap()).unwrap();
        assert_eq!(results[0].index, 1);
        assert!(!results[0].indices.is_empty());
    }
}
