# Synchronize Table Schema

## How To Use

Edit flags in `run_me.go` and execute the command `go generate run_me.go`,
Synchronize-Table-Schema-Tool will depend on the table schema of this mcom
version and migrate tables.
The tool tries to alter or add columns or tables, and any table will **NOT** be
deleted even if it is deprecated.

If you set `--db-source-data` flag, Synchronize-Table-Schema-Tool will synchronize
the data from the specified schema(depends on `--db-source-data` flag) to
destination schema(depends on `--db-schema` flag).

## Features

- Migrate tables of the specified destination schema.
- Synchronize the data from source schema to destination schema.

## Caution

1. Migrating tables will **NOT**:
   - set/remove the default column value
   - alter the column to be nullable
   - alter the data type of the column
   - alter the length of the data type
   - drop the column
1. Synchronizing the data will delete all data and create the data from
the source.
