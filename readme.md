# MySQL Punch

This is a small CLI tool for testing load on mysql servers. It runs in a loop, sending new rows to a table one after another, and also concurrently. The maximum number of records to send is the max of a 32-bit unsigned integer (4,294,967,295) though memory usage of the host running this tool increases somewhat linearly to the number of records sent, as it saves a record of durations of time that it took to get a response from mysql for each row sent, for some post-run timing statistics. For this reason, it is recommended to not run this on the same host running the mysql server.

## Installation

This tool currently is not packaged. To run mysqlpunch you will need to build it, so ensure you have Golang installed locally.

```bash
# Clone the repo
git clone https://github.com/Chia-Network/mysqlpunch.git
cd mysqlpunch

# Build mysqlpunch
go build
```

That will create an executable named `mysqlpunch` in the current directory.

## Usage

```
Usage:
  mysqlpunch [flags]

Flags:
      --create-db               When set to true, this will handle creating the database in your mysql server. (defaults to false)
  -h, --help                    help for mysqlpunch
      --log-level string        Log verbosity. Should be one of: panic, fatal, error, warn, info, debug, trace. (default "info")
      --max-concurrent uint32   The max number of records to send concurrently (in individual requests.) (default 1)
      --mysql-database string   The mysql database to use.
      --mysql-host string       The hostname to connect to for the mysql db.
      --mysql-password string   A password for the corresponding mysql username, see the --mysql-user flag.
      --mysql-user string       A mysql username to authenticate as, requires a password, see the --mysql-password flag.
      --records uint32          The number of rows or records to send.
      --reset                   When set to true, this resets the mysqlpunch table at the beginning of a run, deleting all records in it and resetting the ID counter. (defaults to false)
```

ex:
```bash
mysqlpunch --mysql-host="10.0.0.15" --mysql-database="mysqlpunch" --mysql-user="<redacted>" --mysql-password="<redacted>" --records=10000 --max-concurrent=40
```

Other flags like `--reset`, `--log-level`, and `--create-db` are shown in the usage text above.
