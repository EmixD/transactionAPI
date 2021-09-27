# transactionAPI

This is a sample Golang/gin/MongoDB API for managing the deposits and transactions of different users.

The ENV file should contain:

MONGODB_URL;
DBNAME;
COLLECTION_USERS_NAME;
COLLECTION_DEPOSITS_NAME;
COLLECTION_TRANSACTIONS_NAME;

The collections are assumed to be empty at the server startup.

Files:
main.go - general startup and shutdown;
db.go - everything related to database;
api.go - the API functions themselves;
structs.go - The structs used by the API.

The server is configured to shutdown gracefully on SIGINT.