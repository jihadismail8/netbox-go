package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"netbox-go/internal/adapters/postgres/bootstrap"
)

const privateSchemaFixture = "testdata/private_schema_contract.json"

type privateSchemaContract struct {
	Counts    privateSchemaCounts     `json:"counts"`
	Tables    []privateTableContract  `json:"tables"`
	Sequences []privateSequenceRecord `json:"sequences"`
}

type privateSchemaCounts struct {
	Tables           int `json:"tables"`
	Columns          int `json:"columns"`
	Constraints      int `json:"constraints"`
	PrimaryKeys      int `json:"primary_keys"`
	ForeignKeys      int `json:"foreign_keys"`
	Checks           int `json:"checks"`
	Indexes          int `json:"indexes"`
	PrimaryIndexes   int `json:"primary_indexes"`
	UniqueIndexes    int `json:"unique_indexes"`
	NonUniqueIndexes int `json:"non_unique_indexes"`
	Sequences        int `json:"sequences"`
}

type privateTableContract struct {
	Name        string                    `json:"name"`
	Columns     []privateColumnContract   `json:"columns"`
	Constraints []privateConstraintRecord `json:"constraints"`
	Indexes     []privateIndexRecord      `json:"indexes"`
}

type privateColumnContract struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Nullable bool   `json:"nullable"`
	Default  string `json:"default"`
	Sequence string `json:"sequence,omitempty"`
}

type privateConstraintRecord struct {
	Name              string   `json:"name"`
	Kind              string   `json:"kind"`
	Columns           []string `json:"columns"`
	Deferrable        bool     `json:"deferrable"`
	InitiallyDeferred bool     `json:"initially_deferred"`
	Validated         bool     `json:"validated"`
	ReferencedTable   string   `json:"referenced_table,omitempty"`
	ReferencedColumns []string `json:"referenced_columns,omitempty"`
	OnUpdate          string   `json:"on_update,omitempty"`
	OnDelete          string   `json:"on_delete,omitempty"`
	Expression        string   `json:"expression,omitempty"`
}

type privateIndexRecord struct {
	Name             string   `json:"name"`
	Unique           bool     `json:"unique"`
	Primary          bool     `json:"primary"`
	Method           string   `json:"method"`
	Keys             []string `json:"keys"`
	Included         []string `json:"included,omitempty"`
	Predicate        string   `json:"predicate,omitempty"`
	Valid            bool     `json:"valid"`
	Ready            bool     `json:"ready"`
	NullsNotDistinct bool     `json:"nulls_not_distinct"`
}

type privateSequenceRecord struct {
	Name   string `json:"name"`
	Table  string `json:"table"`
	Column string `json:"column"`
}

func TestPrivateSchemaContractPostgres(t *testing.T) {
	dsn := strings.TrimSpace(os.Getenv("NETBOX_TEST_POSTGRES_DSN"))
	if dsn == "" {
		if os.Getenv("NETBOX_REQUIRE_TEST_POSTGRES") == "1" {
			t.Fatal("NETBOX_TEST_POSTGRES_DSN is required")
		}
		t.Skip("NETBOX_TEST_POSTGRES_DSN is not set")
	}

	admin, err := gorm.Open(
		postgres.Open(dsn),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	if err != nil {
		t.Fatalf("open PostgreSQL: %v", err)
	}
	adminSQL, err := admin.DB()
	if err != nil {
		t.Fatalf("get PostgreSQL admin connection: %v", err)
	}
	t.Cleanup(func() {
		if err := adminSQL.Close(); err != nil {
			t.Errorf("close PostgreSQL admin connection: %v", err)
		}
	})
	schema := fmt.Sprintf("netbox_private_schema_%d", time.Now().UnixNano())
	if err := admin.Exec(`CREATE SCHEMA "` + schema + `"`).Error; err != nil {
		t.Fatalf("create isolated schema: %v", err)
	}
	t.Cleanup(func() {
		if err := admin.Exec(`DROP SCHEMA "` + schema + `" CASCADE`).Error; err != nil {
			t.Errorf("drop isolated schema: %v", err)
		}
	})

	db, err := gorm.Open(
		postgres.Open(postgresTestDSN(t, dsn, schema)),
		&gorm.Config{Logger: logger.Default.LogMode(logger.Silent)},
	)
	if err != nil {
		t.Fatalf("open isolated PostgreSQL schema: %v", err)
	}
	scopedSQL, err := db.DB()
	if err != nil {
		t.Fatalf("get isolated PostgreSQL connection: %v", err)
	}
	t.Cleanup(func() {
		if err := scopedSQL.Close(); err != nil {
			t.Errorf("close isolated PostgreSQL connection: %v", err)
		}
	})
	registry, err := modelRegistry()
	if err != nil {
		t.Fatalf("build startup registry: %v", err)
	}
	privateEntries := make([]bootstrap.Entry, 0, 22)
	for _, entry := range registry.Entries() {
		if strings.HasPrefix(entry.Name, "go_") {
			privateEntries = append(privateEntries, entry)
		}
	}
	privateRegistry, err := bootstrap.NewRegistry(privateEntries...)
	if err != nil {
		t.Fatalf("build private startup registry: %v", err)
	}
	if _, err := bootstrap.Run(t.Context(), db, privateRegistry); err != nil {
		t.Fatalf("bootstrap private schema: %v", err)
	}

	actual := inspectPrivateSchema(t, db, schema)
	expected := loadPrivateSchemaContract(t)
	wantCounts := privateSchemaCounts{
		Tables:           22,
		Columns:          197,
		Constraints:      47,
		PrimaryKeys:      22,
		ForeignKeys:      24,
		Checks:           1,
		Indexes:          99,
		PrimaryIndexes:   22,
		UniqueIndexes:    22,
		NonUniqueIndexes: 55,
		Sequences:        18,
	}
	if expected.Counts != wantCounts {
		t.Fatalf("reviewed fixture counts = %+v, want %+v", expected.Counts, wantCounts)
	}
	if actual.Counts != wantCounts {
		t.Fatalf("PostgreSQL private schema counts = %+v, want %+v", actual.Counts, wantCounts)
	}
	if !reflect.DeepEqual(actual.Sequences, expected.Sequences) {
		t.Fatalf("private sequence ownership mismatch:\n%s", schemaDifference(t, expected.Sequences, actual.Sequences))
	}
	if len(actual.Tables) != len(expected.Tables) {
		t.Fatalf("private table count = %d, want %d", len(actual.Tables), len(expected.Tables))
	}
	for index := range expected.Tables {
		if !reflect.DeepEqual(actual.Tables[index], expected.Tables[index]) {
			t.Fatalf(
				"private table contract mismatch at position %d:\n%s",
				index,
				schemaDifference(t, expected.Tables[index], actual.Tables[index]),
			)
		}
	}
}

func loadPrivateSchemaContract(t *testing.T) privateSchemaContract {
	t.Helper()
	contents, err := os.ReadFile(privateSchemaFixture)
	if err != nil {
		t.Fatalf("read %s: %v", privateSchemaFixture, err)
	}
	decoder := json.NewDecoder(strings.NewReader(string(contents)))
	decoder.DisallowUnknownFields()
	var contract privateSchemaContract
	if err := decoder.Decode(&contract); err != nil {
		t.Fatalf("decode %s: %v", privateSchemaFixture, err)
	}
	return contract
}

func inspectPrivateSchema(t *testing.T, db *gorm.DB, schema string) privateSchemaContract {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get PostgreSQL connection: %v", err)
	}
	contract := privateSchemaContract{
		Tables: inspectPrivateTables(t, sqlDB, schema),
	}
	tablePositions := make(map[string]int, len(contract.Tables))
	for index, table := range contract.Tables {
		tablePositions[table.Name] = index
	}
	inspectPrivateColumns(t, sqlDB, schema, contract.Tables, tablePositions)
	inspectPrivateConstraints(t, sqlDB, schema, contract.Tables, tablePositions)
	inspectPrivateIndexes(t, sqlDB, schema, contract.Tables, tablePositions)
	contract.Sequences = inspectPrivateSequences(t, sqlDB, schema)
	contract.Counts = countPrivateSchema(contract)
	return contract
}

func inspectPrivateTables(t *testing.T, db *sql.DB, schema string) []privateTableContract {
	t.Helper()
	rows, err := db.QueryContext(t.Context(), `
		SELECT table_record.relname
		FROM pg_class table_record
		JOIN pg_namespace namespace_record ON namespace_record.oid = table_record.relnamespace
		WHERE namespace_record.nspname = $1
		  AND table_record.relkind IN ('r', 'p')
		  AND left(table_record.relname, 3) = 'go_'
		ORDER BY table_record.relname`, schema)
	if err != nil {
		t.Fatalf("query private tables: %v", err)
	}
	defer closePrivateRows(t, rows, "private-table")
	var tables []privateTableContract
	for rows.Next() {
		var table privateTableContract
		if err := rows.Scan(&table.Name); err != nil {
			t.Fatalf("scan private table: %v", err)
		}
		table.Columns = []privateColumnContract{}
		table.Constraints = []privateConstraintRecord{}
		table.Indexes = []privateIndexRecord{}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate private tables: %v", err)
	}
	return tables
}

func inspectPrivateColumns(
	t *testing.T,
	db *sql.DB,
	schema string,
	tables []privateTableContract,
	positions map[string]int,
) {
	t.Helper()
	rows, err := db.QueryContext(t.Context(), `
		SELECT
			t.relname,
			a.attname,
			pg_catalog.format_type(a.atttypid, a.atttypmod),
			NOT a.attnotnull,
			COALESCE(pg_get_expr(d.adbin, d.adrelid, true), ''),
			COALESCE((
				SELECT sequence.relname
				FROM pg_depend dependency
				JOIN pg_class sequence ON sequence.oid = dependency.objid
				JOIN pg_namespace sequence_namespace ON sequence_namespace.oid = sequence.relnamespace
				WHERE dependency.refobjid = t.oid
				  AND dependency.refobjsubid = a.attnum
				  AND dependency.deptype IN ('a', 'i')
				  AND sequence.relkind = 'S'
				  AND sequence_namespace.nspname = $1
				ORDER BY sequence.relname
				LIMIT 1
			), '')
		FROM pg_class t
		JOIN pg_namespace n ON n.oid = t.relnamespace
		JOIN pg_attribute a ON a.attrelid = t.oid
		LEFT JOIN pg_attrdef d ON d.adrelid = t.oid AND d.adnum = a.attnum
		WHERE n.nspname = $1
		  AND t.relkind IN ('r', 'p')
		  AND left(t.relname, 3) = 'go_'
		  AND a.attnum > 0
		  AND NOT a.attisdropped
		ORDER BY t.relname, a.attnum`, schema)
	if err != nil {
		t.Fatalf("query private columns: %v", err)
	}
	defer closePrivateRows(t, rows, "private-column")
	for rows.Next() {
		var tableName, defaultExpression, sequence string
		var column privateColumnContract
		if err := rows.Scan(
			&tableName,
			&column.Name,
			&column.Type,
			&column.Nullable,
			&defaultExpression,
			&sequence,
		); err != nil {
			t.Fatalf("scan private column: %v", err)
		}
		position, ok := positions[tableName]
		if !ok {
			t.Fatalf("column belongs to unexpected private table %q", tableName)
		}
		column.Sequence = sequence
		column.Default = normalizePrivateDefault(schema, sequence, defaultExpression)
		tables[position].Columns = append(tables[position].Columns, column)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate private columns: %v", err)
	}
}

func inspectPrivateConstraints(
	t *testing.T,
	db *sql.DB,
	schema string,
	tables []privateTableContract,
	positions map[string]int,
) {
	t.Helper()
	rows, err := db.QueryContext(t.Context(), `
		SELECT
			t.relname,
			constraint_record.conname,
			constraint_record.contype::text,
			constraint_record.condeferrable,
			constraint_record.condeferred,
			constraint_record.convalidated,
			COALESCE((
				SELECT string_agg(attribute.attname, ',' ORDER BY key_position.ordinality)
				FROM unnest(constraint_record.conkey) WITH ORDINALITY AS key_position(attnum, ordinality)
				JOIN pg_attribute attribute
				  ON attribute.attrelid = constraint_record.conrelid
				 AND attribute.attnum = key_position.attnum
			), ''),
			COALESCE(referenced_table.relname, ''),
			COALESCE((
				SELECT string_agg(attribute.attname, ',' ORDER BY key_position.ordinality)
				FROM unnest(constraint_record.confkey) WITH ORDINALITY AS key_position(attnum, ordinality)
				JOIN pg_attribute attribute
				  ON attribute.attrelid = constraint_record.confrelid
				 AND attribute.attnum = key_position.attnum
			), ''),
			constraint_record.confupdtype::text,
			constraint_record.confdeltype::text,
			COALESCE(pg_get_expr(constraint_record.conbin, constraint_record.conrelid, true), '')
		FROM pg_constraint constraint_record
		JOIN pg_class t ON t.oid = constraint_record.conrelid
		JOIN pg_namespace n ON n.oid = t.relnamespace
		LEFT JOIN pg_class referenced_table ON referenced_table.oid = constraint_record.confrelid
		WHERE n.nspname = $1
		  AND left(t.relname, 3) = 'go_'
		ORDER BY t.relname, constraint_record.conname`, schema)
	if err != nil {
		t.Fatalf("query private constraints: %v", err)
	}
	defer closePrivateRows(t, rows, "private-constraint")
	for rows.Next() {
		var tableName, kind, columns, referencedColumns, updateCode, deleteCode string
		var constraint privateConstraintRecord
		if err := rows.Scan(
			&tableName,
			&constraint.Name,
			&kind,
			&constraint.Deferrable,
			&constraint.InitiallyDeferred,
			&constraint.Validated,
			&columns,
			&constraint.ReferencedTable,
			&referencedColumns,
			&updateCode,
			&deleteCode,
			&constraint.Expression,
		); err != nil {
			t.Fatalf("scan private constraint: %v", err)
		}
		position, ok := positions[tableName]
		if !ok {
			t.Fatalf("constraint belongs to unexpected private table %q", tableName)
		}
		constraint.Columns = splitCatalogList(columns)
		switch kind {
		case "p":
			constraint.Kind = "primary_key"
		case "f":
			constraint.Kind = "foreign_key"
			constraint.ReferencedColumns = splitCatalogList(referencedColumns)
			constraint.OnUpdate = foreignKeyAction(updateCode)
			constraint.OnDelete = foreignKeyAction(deleteCode)
		case "c":
			constraint.Kind = "check"
		default:
			constraint.Kind = "postgres_" + kind
		}
		tables[position].Constraints = append(tables[position].Constraints, constraint)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate private constraints: %v", err)
	}
}

func inspectPrivateIndexes(
	t *testing.T,
	db *sql.DB,
	schema string,
	tables []privateTableContract,
	positions map[string]int,
) {
	t.Helper()
	rows, err := db.QueryContext(t.Context(), `
		SELECT
			t.relname,
			index_relation.relname,
			index_record.indisunique,
			index_record.indisprimary,
			index_record.indisvalid,
			index_record.indisready,
			index_record.indnullsnotdistinct,
			access_method.amname,
			COALESCE((
				SELECT string_agg(
					pg_get_indexdef(index_record.indexrelid, key_position, true),
					E'\x1f' ORDER BY key_position
				)
				FROM generate_series(1, index_record.indnkeyatts) key_position
			), ''),
			COALESCE((
				SELECT string_agg(
					pg_get_indexdef(index_record.indexrelid, included_position, true),
					E'\x1f' ORDER BY included_position
				)
				FROM generate_series(index_record.indnkeyatts + 1, index_record.indnatts) included_position
			), ''),
			COALESCE(pg_get_expr(index_record.indpred, index_record.indrelid, true), '')
		FROM pg_index index_record
		JOIN pg_class t ON t.oid = index_record.indrelid
		JOIN pg_namespace n ON n.oid = t.relnamespace
		JOIN pg_class index_relation ON index_relation.oid = index_record.indexrelid
		JOIN pg_am access_method ON access_method.oid = index_relation.relam
		WHERE n.nspname = $1
		  AND left(t.relname, 3) = 'go_'
		ORDER BY t.relname, index_relation.relname`, schema)
	if err != nil {
		t.Fatalf("query private indexes: %v", err)
	}
	defer closePrivateRows(t, rows, "private-index")
	for rows.Next() {
		var tableName, keys, included string
		var index privateIndexRecord
		if err := rows.Scan(
			&tableName,
			&index.Name,
			&index.Unique,
			&index.Primary,
			&index.Valid,
			&index.Ready,
			&index.NullsNotDistinct,
			&index.Method,
			&keys,
			&included,
			&index.Predicate,
		); err != nil {
			t.Fatalf("scan private index: %v", err)
		}
		position, ok := positions[tableName]
		if !ok {
			t.Fatalf("index belongs to unexpected private table %q", tableName)
		}
		index.Keys = strings.Split(keys, "\x1f")
		if included != "" {
			index.Included = strings.Split(included, "\x1f")
		}
		tables[position].Indexes = append(tables[position].Indexes, index)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate private indexes: %v", err)
	}
}

func inspectPrivateSequences(t *testing.T, db *sql.DB, schema string) []privateSequenceRecord {
	t.Helper()
	rows, err := db.QueryContext(t.Context(), `
		SELECT
			sequence.relname,
			COALESCE(owned_table.relname, ''),
			COALESCE(owned_column.attname, '')
		FROM pg_class sequence
		JOIN pg_namespace n ON n.oid = sequence.relnamespace
		LEFT JOIN pg_depend dependency
		  ON dependency.objid = sequence.oid
		 AND dependency.deptype IN ('a', 'i')
		LEFT JOIN pg_class owned_table ON owned_table.oid = dependency.refobjid
		LEFT JOIN pg_attribute owned_column
		  ON owned_column.attrelid = owned_table.oid
		 AND owned_column.attnum = dependency.refobjsubid
		WHERE n.nspname = $1
		  AND sequence.relkind = 'S'
		  AND left(sequence.relname, 3) = 'go_'
		ORDER BY sequence.relname`, schema)
	if err != nil {
		t.Fatalf("query private sequences: %v", err)
	}
	defer closePrivateRows(t, rows, "private-sequence")
	var sequences []privateSequenceRecord
	for rows.Next() {
		var sequence privateSequenceRecord
		if err := rows.Scan(&sequence.Name, &sequence.Table, &sequence.Column); err != nil {
			t.Fatalf("scan private sequence: %v", err)
		}
		sequences = append(sequences, sequence)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate private sequences: %v", err)
	}
	return sequences
}

func normalizePrivateDefault(schema, sequence, expression string) string {
	expression = strings.TrimSpace(expression)
	if sequence != "" {
		unqualified := "nextval('" + sequence + "'::regclass)"
		qualified := "nextval('" + schema + "." + sequence + "'::regclass)"
		if expression == unqualified || expression == qualified {
			return "sequence"
		}
		return expression
	}
	switch expression {
	case "":
		return "none"
	case "''::character varying", "''::text":
		return "empty_string"
	case "true", "false":
		return expression
	default:
		return expression
	}
}

func closePrivateRows(t *testing.T, rows *sql.Rows, subject string) {
	t.Helper()
	if err := rows.Close(); err != nil {
		t.Errorf("close %s catalog rows: %v", subject, err)
	}
}

func foreignKeyAction(code string) string {
	switch code {
	case "a":
		return "NO ACTION"
	case "r":
		return "RESTRICT"
	case "c":
		return "CASCADE"
	case "n":
		return "SET NULL"
	case "d":
		return "SET DEFAULT"
	default:
		return "UNKNOWN(" + code + ")"
	}
}

func splitCatalogList(value string) []string {
	if value == "" {
		return []string{}
	}
	return strings.Split(value, ",")
}

func countPrivateSchema(contract privateSchemaContract) privateSchemaCounts {
	counts := privateSchemaCounts{
		Tables:    len(contract.Tables),
		Sequences: len(contract.Sequences),
	}
	for _, table := range contract.Tables {
		counts.Columns += len(table.Columns)
		counts.Constraints += len(table.Constraints)
		for _, constraint := range table.Constraints {
			switch constraint.Kind {
			case "primary_key":
				counts.PrimaryKeys++
			case "foreign_key":
				counts.ForeignKeys++
			case "check":
				counts.Checks++
			}
		}
		counts.Indexes += len(table.Indexes)
		for _, index := range table.Indexes {
			switch {
			case index.Primary:
				counts.PrimaryIndexes++
			case index.Unique:
				counts.UniqueIndexes++
			default:
				counts.NonUniqueIndexes++
			}
		}
	}
	return counts
}

func schemaDifference(t *testing.T, expected, actual any) string {
	t.Helper()
	expectedJSON, err := json.MarshalIndent(expected, "", "  ")
	if err != nil {
		t.Fatalf("encode expected private schema: %v", err)
	}
	actualJSON, err := json.MarshalIndent(actual, "", "  ")
	if err != nil {
		t.Fatalf("encode actual private schema: %v", err)
	}
	return "expected:\n" + string(expectedJSON) + "\nactual:\n" + string(actualJSON)
}
