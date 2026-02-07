# MariaDB Backend

Beads now uses MariaDB as its default storage backend. This provides a robust, server-based database that supports multi-user access and can be shared across multiple machines.

## Prerequisites

Before using Beads with MariaDB, you need to have a MariaDB server installed and running.

### Installing MariaDB

**On Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install mariadb-server
sudo systemctl start mariadb
sudo systemctl enable mariadb
```

**On macOS (with Homebrew):**
```bash
brew install mariadb
brew services start mariadb
```

**On Windows:**
Download and install MariaDB from the [official website](https://mariadb.org/download/).

### Securing MariaDB (Recommended)

After installation, run the security script to set a root password:
```bash
sudo mysql_secure_installation
```

### Creating a Database User (Optional)

For better security, create a dedicated user for Beads:
```bash
sudo mysql -u root -p
```

Then in the MariaDB shell:
```sql
CREATE USER 'beads'@'localhost' IDENTIFIED BY 'your_secure_password';
CREATE DATABASE beads;
GRANT ALL PRIVILEGES ON beads.* TO 'beads'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

## Configuration

### Using Environment Variables

The recommended way to configure MariaDB connection is through environment variables:

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `BEADS_MARIADB_HOST` | MariaDB server host | `127.0.0.1` |
| `BEADS_MARIADB_PORT` | MariaDB server port | `3306` |
| `BEADS_MARIADB_USER` | MariaDB username | `root` |
| `BEADS_MARIADB_PASSWORD` | MariaDB password | (empty) |
| `BEADS_MARIADB_DATABASE` | Database name | `beads` |

Example:
```bash
export BEADS_MARIADB_HOST=127.0.0.1
export BEADS_MARIADB_PORT=3306
export BEADS_MARIADB_USER=beads
export BEADS_MARIADB_PASSWORD=your_secure_password
export BEADS_MARIADB_DATABASE=beads
```

### Using Command Line Flags

You can also specify connection details when initializing:

```bash
bd init --prefix myproject \
  --mariadb-host 127.0.0.1 \
  --mariadb-port 3306 \
  --mariadb-user beads \
  --mariadb-database beads
```

**Note:** For security reasons, passwords should be set via the `BEADS_MARIADB_PASSWORD` environment variable rather than command line flags.

### Using the Configuration File

After initialization, connection settings are stored in `.beads/metadata.json`:

```json
{
  "database": "beads",
  "backend": "mariadb",
  "mariadb_host": "127.0.0.1",
  "mariadb_port": 3306,
  "mariadb_user": "beads",
  "mariadb_database": "beads"
}
```

You can edit this file directly to change connection settings.

## Initializing Beads with MariaDB

Once MariaDB is running and configured, initialize Beads:

```bash
cd your-project
bd init --prefix myproject
```

Beads will automatically:
1. Connect to the MariaDB server
2. Create the `beads` database if it doesn't exist
3. Create all required tables
4. Set up the issue tracking schema

## Verifying the Connection

After initialization, verify everything is working:

```bash
bd list
bd doctor
```

## Troubleshooting

### Connection Refused

If you see "connection refused" errors:

1. Check that MariaDB is running:
   ```bash
   sudo systemctl status mariadb
   ```

2. Verify the server is listening on the expected port:
   ```bash
   sudo netstat -tlnp | grep 3306
   ```

3. Check firewall settings if connecting to a remote server.

### Access Denied

If you see "access denied" errors:

1. Verify your credentials are correct
2. Ensure the user has permissions on the database
3. Check if the user can connect from your host

### Database Does Not Exist

Beads will automatically create the database, but if you see errors:

1. Connect to MariaDB manually and verify permissions:
   ```bash
   mysql -u your_user -p -e "CREATE DATABASE beads;"
   ```

## Using a Different Backend

If you prefer to use SQLite (local file-based storage):

```bash
bd init --backend sqlite --prefix myproject
```

Or for Dolt (version-controlled database):

```bash
bd init --backend dolt --prefix myproject
```

## Migration from SQLite

If you have an existing SQLite database and want to migrate to MariaDB:

1. Export your issues to JSONL:
   ```bash
   bd export > issues_backup.jsonl
   ```

2. Reinitialize with MariaDB:
   ```bash
   rm -rf .beads
   bd init --prefix myproject
   ```

3. Import your issues:
   ```bash
   bd import < issues_backup.jsonl
   ```
