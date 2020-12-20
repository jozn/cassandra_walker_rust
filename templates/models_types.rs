use cdrs::query::*;

use cdrs::frame::IntoBytes;
use cdrs::types::from_cdrs::FromCDRSByName;
use cdrs::types::prelude::*;



/*#[derive(Clone, Debug, PartialEq)]
struct RowStruct {
    key: i32,
    user: User,
    map: HashMap<String, User>,
    list: Vec<User>,
}
*/
{{range .Tables -}}

#[derive(Clone, Debug, PartialEq)]
struct {{ .TableNameGo }} {
	{{range .Columns -}}
	pub {{ .ColumnNameRust }}: {{ .TypeRust }},   // {{ .ColumnName }}    {{ .Kind }}  {{ .Position }}
	{{end}}
	_exists: bool,
	_deleted: bool,
}
/*
:= &xc.{{ .TableNameRust }} {
	{{- range .Columns }}
	{{ .ColumnNameRust }}: {{.TypeDefaultRust}},
	{{- end }}
*/
{{end}}

/*
// logs tables
type LogTableCql struct{
    {{range .Tables }}
    {{ .TableNameGo }} bool
    {{- end}}
}

var LogTableCqlReq = LogTableCql{
	{{- range .Tables }}
    {{ .TableNameGo }}: true ,
    {{- end}}
}

*/
