use cdrs::authenticators::StaticPasswordAuthenticator;
use cdrs::cluster::session::{new as new_session, Session};
use cdrs::cluster::{ClusterTcpConfig, NodeTcpConfigBuilder, TcpConnectionPool};
use cdrs::load_balancing::RoundRobin;
use cdrs::query::*;
use cdrs::frame::Frame;

use cdrs::frame::IntoBytes;
use cdrs::types::from_cdrs::FromCDRSByName;
use cdrs::types::prelude::*;

use crate::xc::common::*;


#[derive(Clone, Debug, PartialEq)]
pub struct {{ .TableNameRust }} {
    {{range .Columns -}}
    pub {{ .ColumnNameRust }}: {{ .TypeRust }},   // {{ .ColumnName }}    {{ .Kind }}  {{ .Position }}
    {{end}}
    _exists: bool,
    _deleted: bool,
}

impl {{ .TableNameRust }} {
    pub fn deleted(&self) -> bool {
        self._deleted
    }

    pub fn exists(&self) -> bool {
        self._exists
    }

    pub fn delete(&mut self, session: &CurrentSession) -> cdrs::error::Result<Frame> {
        let mut deleter = Tweet_Deleter::new();
        {{ range $i, $col := .PartitionColumns }}
            {{- if (eq $i 0)}}
        //deleter.{{$col.ColumnNameRust}}_eq(&self.{{$col.ColumnNameRust}});
            {{- else}}
        //deleter.and_{{$col.ColumnNameRust}}_eq(&self.{{$col.ColumnNameRust}});
            {{end}}
        {{- end }}

        {{- range .ClusterColumns }}
        //deleter.and_{{.ColumnNameRust}}_eq(&self.{{.ColumnNameRust}});
        {{- end }}

        deleter.delete(session)
    }

}

{{- $deleterType := printf "%s%s_Deleter" .PrefixHidden .TableNameRust}}
{{- $updaterType := printf "%s%s_Updater" .PrefixHidden .TableNameRust}}
{{- $selectorType := printf "%s%s_Selector" .PrefixHidden .TableNameRust}}



#[derive(Default, Debug)]
pub struct {{ $deleterType}} {
    wheres: Vec<WhereClause>,
    delete_cols: Vec<&'static str>,
}


impl {{ $deleterType}} {
    pub fn new() -> Self {
        {{ $deleterType}}::default()
    }

    //each column delete
{{- range .Columns }}
    pub fn delete_{{ .ColumnNameRust }}(&mut self) -> &Self {
        self.delete_cols.push("{{.ColumnName}}");
        self
    }
{{ end }}

    pub fn delete(&mut self, session: &CurrentSession) -> cdrs::error::Result<Frame> {
        let del_col = self.delete_cols.join(", ");

        let  mut where_str = vec![];
        let mut where_arr = vec![];

        for w in &self.wheres {
            where_str.push(w.condition);
            where_arr.push(w.args.clone())
        }

        let where_str = where_str.join("");

        let cql_query = format!("DELETE {} FROM {{.TableSchemeOut}} WHERE {}", del_col, where_str);
        //let cql_query = "DELETE " + del_col + " FROM {{.TableSchemeOut}} WHERE " + where_str ;

        let query_values = QueryValues::SimpleValues(where_arr);

        session.query_with_values(cql_query, query_values)
    }

    {{ .GetRustWheresTmplOut }}

}


{{$table := . }}

