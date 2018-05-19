package cwalker

import (
	"github.com/gocql/gocql"
	"log"
	"ms/sun/shared/helper"
)

func describeKeyspace(keyspace string, gen *GenOut, cluster *gocql.ClusterConfig) {
    cluster.Keyspace = keyspace
    session, err := cluster.CreateSession()
    if err != nil {
        log.Fatal(err)
    }
    defer session.Close()

   // iter := session.Query(`DESCRIBE ` + keyspace + ";").String()
}

func loadTables(keyspace string, gen *GenOut, cluster *gocql.ClusterConfig) {
	cluster.Keyspace = keyspace
	session, err := cluster.CreateSession()
	if err != nil {
		log.Fatal(err)
	}
	defer session.Close()

	iter := session.Query(`SELECT * FROM system_schema.tables where keyspace_name = ?`, keyspace).Iter()
	//iter := session.Query(`SELECT * FROM system_schema.tables`).Iter()
	//iter := session.Query(`SELECT * FROM columns `).Iter()
	m := make(map[string]interface{}, 100)
	for iter.MapScan(m) {
		helper.PertyPrint(m)
		t := &Table{
			TableName: (m["table_name"]).(string),
			Keyspace:  (m["keyspace_name"]).(string),
		}
		gen.TablesExtracted = append(gen.TablesExtracted, t)
		m = make(map[string]interface{})
	}
}

func loadColumns(gen *GenOut, cluster *gocql.ClusterConfig) {
	for _, table := range gen.TablesExtracted {
		cluster.Keyspace = table.Keyspace
		session, _ := cluster.CreateSession()
		defer session.Close()

		iter := session.Query(`SELECT * FROM system_schema.columns where  keyspace_name = ? AND table_name = ? `, table.Keyspace, table.TableName).Iter()
		m := make(map[string]interface{}, 100)
		for iter.MapScan(m) {
			helper.PertyPrint(m)
			t := &Column{
				ColumnName: (m["column_name"]).(string),
				TypeCql:    (m["type"]).(string),
			}
			if k, ok := (m["kind"]).(string); ok {
				t.Kind = k
			}
			switch t.Kind {
			case "partition_key":
				t.IsPartition = true
			case "clustering":
				t.IsClustering = true
			case "regular":
				t.IsRegular = true
			}
			table.Columns = append(table.Columns, t)
			m = make(map[string]interface{})
		}

		/*for _, col := range table.Columns {
			if col.IsPartition {
				table.PartitionColumns = append(table.PartitionColumns, col)
			}

			if col.IsClustering {
				table.ClusterColumns = append(table.ClusterColumns, col)
			}
		}*/

	}
}
