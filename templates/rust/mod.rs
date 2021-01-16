{{range .Tables}}
pub mod {{ .TableName }};
{{- end}}

{{range .Tables}}
pub use {{ .TableName }}::*;
{{- end}}