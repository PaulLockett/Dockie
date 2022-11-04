# Dockie (WIP)

Dockie is a server tool to pull data from the twitter API for a given set of users into a querable database.

## Endpoints

### "localhost:8080/"

- Get: Gets the checkpoint of all twitter accounts seen and metadata importatant to runs of "/Refresh"

- Put: Takes a Json body to add twitter accounts that will be used in the "/Refresh" data pulls

### "localhost:8080/Refresh"

- Get: Starts a job that pulls the followers of accounts passed in by the put "/" endpoint as well as the accounts they follow.

**Note:** The data collected by "/Refresh" fill a surrealdb database that can be queried through it's own sql endpoint.

# WIP more instructions to come

to run the server code, you need to install:
- Go 1.19
- surrealdb

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

The server will be running on port 8080 and will be ready to be queried.

## Example Results:

### Find all the usernames of the followers of @Afropolitan that have greater than 10k followers
1. do a put "/" request with the body:
```Json
["2484650978"] # this is @Afropolitan's twitter ID
```
2. Start a refresh with the "/Refresh" endpoint
3. Wait ~5min for data injestion to finish on the Surrealdb server
4. Generate follower graph with some identifier of the account that was given in the put with surrealdb's sql endpoint
```SQL
 let $FollowersIDs = (select FollowerID from follow_map where UserID ="user_data:2484650978" );

 let $FriendsIDs = (select UserID from follow_map where FollowerID ="user_data:2484650978" );

 relate  $FollowersIDs ->follow-> user_data:2484650978 ;

 relate  user_data:2484650978 ->follow->  $FriendsIDs ;
 ```
 5. Query the database using the graph to filter for the correct cohort
 ```SQL
 select * from user_data where ( public_metrics.followers_count > 10000  and ->follow) order by public_metrics.followers_count Desc ;
 -- This Assumes that the @Afrpolitan account is the only one with a generated follow graph
 ```
 6. Publish your resulting JSON response : https://pastebin.com/7y0rvfkw
 
