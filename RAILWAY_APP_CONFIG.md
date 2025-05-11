# Railway Application Configuration

## Connecting Your Application to MySQL

After setting up the MySQL service in Railway, you need to configure your main application service to connect to it properly. This document explains how to set this up.

## Required Environment Variables

Railway automatically generates environment variables for your MySQL service. Make sure your main application service has access to these variables by linking the services in Railway.

Here are the environment variables your application needs:

```
MYSQLUSER=root
MYSQLPASSWORD=${MYSQL_ROOT_PASSWORD}
MYSQLHOST=${RAILWAY_PRIVATE_DOMAIN}
MYSQLPORT=3306
MYSQLDATABASE=railway
```

## Verifying Connection

To verify your application is connecting to the database properly:

1. Check the application logs after deployment
2. Look for these lines which indicate successful connection:
   ```
   Database connection established successfully
   Connected to MySQL version: 8.x.x
   ```

## Schema Initialization

If your tables are not automatically created, you may need to manually initialize the schema. Use the Railway CLI to connect to your MySQL service:

```bash
railway connect --service your-mysql-service-name
```

Then run the schema initialization SQL:

```sql
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    cash_balance DECIMAL(15, 2) DEFAULT 10000.00,
    is_admin BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Add more table definitions as needed...
```

## Troubleshooting

If you encounter database connection issues:

1. **Check Environment Variables**: Make sure all required environment variables are correctly set in your main service.

2. **Network Issues**: Ensure your services are in the same Railway project and network.

3. **Credentials**: Verify username and password are correct.

4. **Schema Issues**: Check if tables exist and are properly created.

5. **Connection Limits**: MySQL has connection limits. Ensure your application isn't creating too many connections.

## Testing MySQL Connection

You can test the MySQL connection directly with this simple command in your Railway service terminal:

```bash
mysql -h$MYSQLHOST -u$MYSQLUSER -p$MYSQLPASSWORD -P$MYSQLPORT $MYSQLDATABASE -e "SHOW TABLES;"
```

This should display all tables in your database if the connection is successful.