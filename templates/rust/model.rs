use cdrs::authenticators::StaticPasswordAuthenticator;
use cdrs::cluster::session::{new as new_session, Session};
use cdrs::cluster::{ClusterTcpConfig, NodeTcpConfigBuilder, TcpConnectionPool};
use cdrs::load_balancing::RoundRobin;
// use cdrs::query::*;
use cdrs::query::{QueryValues,QueryExecutor};
use cdrs::frame::Frame;
use cdrs::types::value::ValueType;

use cdrs::frame::IntoBytes;
use cdrs::types::from_cdrs::FromCDRSByName;
use cdrs::types::prelude::*;
use cdrs::types::ByName;
use std::collections::HashMap;
use std::result::Result; // override prelude Result

//use cdrs::error::{Error as CWError};
use cdrs::frame::frame_error::CDRSError;
use cdrs::Error as DriverError;
use crate::xc::common::*;

{{- $deleterType := printf "%s%s_Deleter" .PrefixHidden .TableNameRust}}
{{- $updaterType := printf "%s%s_Updater" .PrefixHidden .TableNameRust}}
{{- $selectorType := printf "%s%s_Selector" .PrefixHidden .TableNameRust}}

#[derive(Default, Clone, Debug, PartialEq)]
pub struct {{ .TableNameRust }} {
    {{- range .Columns}}
    pub {{ .ColumnNameRust }}: {{ .TypeRust }},   // {{ .ColumnName }}    {{ .Kind }}  {{ .Position }}
    {{- end}}
}

impl {{ .TableNameRust }} {
    pub fn save(&self, session: impl FCQueryExecutor) -> Result<(),CWError> {
        let mut columns = vec![];
        let mut values :Vec<Value> = vec![];

        {{ .GetRustModelSavePartial }}

        if columns.len() == 0 {
            return Err(CWError::InvalidCQL)
        }

        let cql_columns = columns.join(", ");
        let mut cql_question = "?,".repeat(columns.len());
        cql_question.remove(cql_question.len()-1);

        let cql_query = format!("INSERT INTO {{ .TableSchemeOut }} ({}) VALUES ({})", cql_columns, cql_question);

        println!("{} - {}", &cql_query, &cql_question);

        session.query_with_values(cql_query, values)?;

        Ok(())
    }

    pub fn delete(&self, session: impl FCQueryExecutor) -> Result<(), CWError> {
        let mut deleter = {{$deleterType}}::new();

    {{- range $i, $col := .PartitionColumns }}
      {{if (eq $i 0) }}
        deleter.{{$col.ColumnNameRust}}_eq({{.RustBorrowSign}}self.{{$col.ColumnNameRust}});
    	{{- else -}}
        deleter.and_{{$col.ColumnNameRust}}_Eq({{.RustBorrowSign}}self.{{$col.ColumnNameRust}});
      {{- end}}
    {{ end -}}

    {{- range .ClusterColumns }}
        deleter.and_{{.ColumnNameRust}}_eq({{.RustBorrowSign}}self.{{.ColumnNameRust}});
    {{- end }}

        let res = deleter.delete(session)?;

        Ok(())
    }

}

fn _get_where(wheres: Vec<WhereClause>) ->  (String, Vec<Value>) {
    let mut values = vec![];
    let  mut where_str = vec![];

    for w in wheres {
        where_str.push(w.condition);
        values.push(w.args)
    }
    let cql_where = where_str.join(" ");

    (cql_where, values)
}

#[derive(Default, Debug)]
pub struct {{ $selectorType }} {
    wheres: Vec<WhereClause>,
    select_cols: Vec<&'static str>,
    order_by: Vec<&'static str>,
    limit: u32,
    allow_filter: bool,
}

impl {{ $selectorType }} {
    pub fn new() -> Self {
        {{ $selectorType }}::default()
    }

    pub fn limit(&mut self, size: u32) -> &mut Self {
        self.limit = size;
        self
    }

    pub fn allow_filtering(&mut self, allow: bool) -> &mut Self {
        self.allow_filter = allow;
        self
    }

    pub fn select_all(&mut self) -> &mut Self {
        // Default is select *
        self
    }

    //each column select
    {{- range .Columns }}
    pub fn select_{{ .ColumnNameRust }}(&mut self) -> &mut Self {
        self.select_cols.push("{{.ColumnName}}");
        self
    }
    {{ end }}

    pub fn _to_cql(&self) ->  (String, Vec<Value>)  {
        let cql_select = if self.select_cols.is_empty() {
            "*".to_string()
        } else {
            self.select_cols.join(", ")
        };

        let mut cql_query = format!("SELECT {} FROM {{.TableSchemeOut}}", cql_select);

        let (cql_where, where_values) = _get_where(self.wheres.clone());

        if where_values.len() > 0 {
            cql_query.push_str(&format!(" WHERE {}",&cql_where));
        }

        if self.order_by.len() > 0 {
            let cql_orders = self.order_by.join(", ");
            cql_query.push_str( &format!(" ORDER BY {}", &cql_orders));
        };

        if self.limit != 0  {
            cql_query.push_str(&format!(" LIMIT {} ", self.limit));
        };

        if self.allow_filter  {
            cql_query.push_str(" ALLOW FILTERING");
        };

        (cql_query, where_values)
    }

    pub fn _get_rows_with_size(&mut self,session: impl FCQueryExecutor, size: i64) -> Result<Vec<{{ .TableNameRust }}>, CWError>   {

        let(cql_query, query_values) = self._to_cql();

        println!("{} - {:?}", &cql_query, &query_values);

        let query_result = session
            .query_with_values(cql_query,query_values)?
            .get_body()?
            .into_rows();

        let db_raws = match query_result {
            Some(rs) => {
                if size > 0 {
                    if rs.len() == size as usize {
                        rs
                    } else {
                        let min = (size as usize).min(rs.len());
                        rs[0..min].to_vec()
                    }
                } else {
                    rs
                }
            },
            None => return Err(CWError::NotFound)
        };

        let mut rows = vec![];

        for db_row in db_raws {
            let mut row = {{ .TableNameRust }}::default();
            {{range .Columns }}
                {{if (eq .TypeCql "blob")}}
            row.{{ .ColumnNameRust }} = db_row.by_name::<Blob>("{{ .ColumnName }}")?.unwrap_or(Blob::new(vec![])).into_vec();
                {{- else}}
            row.{{ .ColumnNameRust }} = db_row.by_name("{{ .ColumnName }}")?.unwrap_or_default();
                {{- end}}
            {{- end }}

            rows.push(row);
        }

        Ok(rows)
    }

    pub fn get_rows(&mut self, session: impl FCQueryExecutor) -> Result<Vec<{{ .TableNameRust }}>, CWError>{
        self._get_rows_with_size(session,-1)
    }

    pub fn get_row(&mut self, session: impl FCQueryExecutor) -> Result<{{ .TableNameRust }}, CWError>{
        let rows = self._get_rows_with_size(session,1)?;

        let opt = rows.get(0);
        match opt {
            Some(row) => Ok(row.to_owned()),
            None => Err(CWError::NotFound)
        }
    }

    {{ .GetRustSelectorOrders }}

    {{ .GetRustWheresTmplOut }}

    {{ .GetRustWhereInsTmplOut }}

}


#[derive(Default, Debug)]
pub struct {{ $deleterType}} {
    wheres: Vec<WhereClause>,
    delete_cols: Vec<&'static str>,
}

#[derive(Default, Debug)]
pub struct {{ $updaterType}} {
    wheres: Vec<WhereClause>,
    updates: HashMap<&'static str, Value>,
}

impl {{ $updaterType}} {
    pub fn new() -> Self {
        {{ $updaterType}}::default()
    }

    pub fn update(&mut self,session: impl FCQueryExecutor) -> cdrs::error::Result<Frame>  {
        if self.updates.is_empty() {
            return Err(cdrs::error::Error::General("empty".to_string()));
        }

        // Update columns building
        let mut all_vals = vec![];
        let mut col_updates = vec![];

        for (col,val) in self.updates.clone() {
            all_vals.push(val);
            col_updates.push(col);
        }
        let cql_update = col_updates.join(",");

        // Where columns building
        let  mut where_str = vec![];

        for w in self.wheres.clone() {
            where_str.push(w.condition);
            all_vals.push(w.args)
        }
        let cql_where = where_str.join(" ");

        // Build final query
        let mut cql_query = if self.wheres.is_empty() {
            format!("UPDATE {{.TableSchemeOut}} SET {}", cql_update)
        } else {
            format!("UPDATE {{.TableSchemeOut}} SET {} WHERE {}", cql_update, cql_where)
        };

        let query_values = QueryValues::SimpleValues(all_vals);
        println!("{} - {:?}", &cql_query, &query_values);

        session.query_with_values(cql_query, query_values)
    }

    {{ .GetRustUpdaterFnsOut }}

    {{ .GetRustWheresTmplOut }}

    {{ .GetRustWhereInsTmplOut }}
}

impl {{ $deleterType}} {
    pub fn new() -> Self {
        {{ $deleterType}}::default()
    }

    //each column delete
    {{- range .Columns }}
    pub fn delete_{{ .ColumnNameRust }}(&mut self) -> &mut Self {
        self.delete_cols.push("{{.ColumnName}}");
        self
    }
    {{ end }}

    pub fn delete(&mut self, session: impl FCQueryExecutor) -> Result<(),CWError> {
        let del_col = self.delete_cols.join(", ");

        let  mut where_str = vec![];
        let mut where_arr = vec![];

        for w in self.wheres.clone() {
            where_str.push(w.condition);
            where_arr.push(w.args)
        }

        let where_str = where_str.join(" ");

        let cql_query = format!("DELETE {} FROM {{.TableSchemeOut}} WHERE {}", del_col, where_str);
        //let cql_query = "DELETE " + del_col + " FROM {{.TableSchemeOut}} WHERE " + where_str ;

        let query_values = QueryValues::SimpleValues(where_arr);
        println!("{} - {:?}", &cql_query, &query_values);

        session.query_with_values(cql_query, query_values)?;

        Ok(())
    }

    {{ .GetRustWheresTmplOut }}

    {{ .GetRustWhereInsTmplOut }}
}

{{ .GetRustPrimaryGetter }}

{{$table := . }}
