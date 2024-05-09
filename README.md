# go-parser
Parser for Ethereum blockchain


# To run application
 - go build -o myprogram main.go
 - ./myprogram 
 - At the prompt `Enter command (e.g: getCurrentBlock)` you can enter various commands like 
    `getCurrentBlock`
    `subscribeAddress 0xb794f5ea0ba39494ce839613fffba74279579268` 
    `getTransaction 0xb794f5ea0ba39494ce839613fffba74279579268`


## Note
- it has functions like getCurrentBlock, subsrcibeAddress and getTransactions
- Not scanning the all blocks in the entire chain, but can implement blocks scan since when address balance is greater than 0
- The MemoryStorage struct provides a basic in-memory storage for suubscribers. You can extend this by implementing persistent storage (e.g., using a database) by modifying the MemoryStorage methods.
- Error handling is simplified and no tests added for demonstration purposes. In production code, should handle errors more robustly and wrrite tests for all edge cases.

