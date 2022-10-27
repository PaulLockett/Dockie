to run the server code, you need to install:
- Go 1.19
- surrealdb

# WIP more instructions to come

## Local development

add a twitter bearer token to .env and a api key ( for request to the servers X-API-KEY header) to .env

### Run these commands to start and query the local server:

```bash
 surreal.exe start --log debug --user root --pass root memory
```

the database will be running on port 8000

### Then when the data base is ready to run:
    
```bash
go run cmd.go
```

in the server directory.

The server will be running on port 8080.