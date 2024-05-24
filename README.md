# pinkfishplatform

# Workbook Management System

## Overview

The Workbook Management System is designed to manage workbook records for users, allowing creation, retrieval, sharing, and updating of workbooks stored in AWS DynamoDB. The core components include the Client Application, WorkBookHandler, WorkBookRecordService, WorkBookRepository, and DynamoDB.

## System Architecture
### HTTP ENDPOINTS
```mermaid
flowchart TD
    subgraph HTTP_ENDPOINTS
        A(Create WorkBook Endpoint) --> B{Check User ID}
        B -- User ID Missing --> C1[Return 400 Bad Request with Error]
        B -- User ID Present --> D{Parse Request Body}
        D -- Invalid Data --> C2[Return 400 Bad Request with Error]
        D -- Valid Data --> E{Validate Python Code}
        E -- Code Too Large --> C3[Return 413 Payload Too Large with Error]
        E -- Code Valid --> F[Insert WorkBook Record]
        F -- Insertion Failed --> C4[Return 500 Internal Server Error with Error]
        F -- Insertion Successful --> G[Return 200 OK]
        
        H(Retrieve WorkBooks Endpoint) --> I{Check User ID}
        I -- User ID Missing --> C1
        I -- User ID Present --> J[Retrieve WorkBook Records]
        J -- Retrieval Failed --> C5[Return 500 Internal Server Error with Error]
        J -- Retrieval Successful --> K[Return 200 OK with WorkBook Records]
        
        L(Retrieve Shared WorkBooks Endpoint) --> M{Check User ID}
        M -- User ID Missing --> C1
        M -- User ID Present --> N[Retrieve Shared WorkBook Records]
        N -- Retrieval Failed --> C5
        N -- Retrieval Successful --> O[Return 200 OK with Shared WorkBook Records]
        
        P(Share WorkBook Endpoint) --> Q{Check User ID}
        Q -- User ID Missing --> C1
        Q -- User ID Present --> R{Parse Request Body}
        R -- Invalid Data --> C2
        R -- Valid Data --> S[Share WorkBook with Users]
        S -- Sharing Failed --> C4
        S -- Sharing Successful --> T[Return 200 OK]
    end
```
### SYSTEM FLOW DIAGRAMM
```mermaid
sequenceDiagram
    participant Client
    participant WorkBookHandler
    participant WorkBookRecordService
    participant WorkBookRepository
    participant DynamoDB

    Client ->> WorkBookHandler: CreateWorkBook
    activate WorkBookHandler
    WorkBookHandler ->> WorkBookRecordService: InsertToWorkBookRecord
    activate WorkBookRecordService
    WorkBookRecordService ->> WorkBookRepository: InsertNewWorkBookRecord
    activate WorkBookRepository
    WorkBookRepository ->> DynamoDB: PutItem
    activate DynamoDB
    DynamoDB -->> WorkBookRepository: PutItemResponse
    deactivate DynamoDB
    WorkBookRepository -->> WorkBookRecordService: InsertNewWorkBookRecordResponse
    deactivate WorkBookRepository
    WorkBookRecordService -->> WorkBookHandler: InsertToWorkBookRecordResponse
    deactivate WorkBookRecordService
    WorkBookHandler -->> Client: CreateWorkBookResponse
    deactivate WorkBookHandler

    Client ->> WorkBookHandler: RetrieveWorkBooks
    activate WorkBookHandler
    WorkBookHandler ->> WorkBookRecordService: RetrieveFromWorkBookRecords
    activate WorkBookRecordService
    WorkBookRecordService ->> WorkBookRepository: RetrieveWorkBookRecords
    activate WorkBookRepository
    WorkBookRepository ->> DynamoDB: Query
    activate DynamoDB
    DynamoDB -->> WorkBookRepository: QueryResponse
    deactivate DynamoDB
    WorkBookRepository -->> WorkBookRecordService: RetrieveWorkBookRecordsResponse
    deactivate WorkBookRepository
    WorkBookRecordService -->> WorkBookHandler: RetrieveFromWorkBookRecordsResponse
    deactivate WorkBookRecordService
    WorkBookHandler -->> Client: RetrieveWorkBooksResponse
    deactivate WorkBookHandler

    Client ->> WorkBookHandler: RetrieveSharedWorkBooks
    activate WorkBookHandler
    WorkBookHandler ->> WorkBookRecordService: RetrieveSharedWorkBookRecords
    activate WorkBookRecordService
    WorkBookRecordService ->> WorkBookRepository: RetrieveSharedWorkBookRecords
    activate WorkBookRepository
    WorkBookRepository ->> DynamoDB: Query with Filter
    activate DynamoDB
    DynamoDB -->> WorkBookRepository: QueryResponse
    deactivate DynamoDB
    WorkBookRepository -->> WorkBookRecordService: RetrieveSharedWorkBookRecordsResponse
    deactivate WorkBookRepository
    WorkBookRecordService -->> WorkBookHandler: RetrieveSharedWorkBookRecordsResponse
    deactivate WorkBookRecordService
    WorkBookHandler -->> Client: RetrieveSharedWorkBooksResponse
    deactivate WorkBookHandler

    Client ->> WorkBookHandler: ShareWorkBook
    activate WorkBookHandler
    WorkBookHandler ->> WorkBookRecordService: AddNewUserToWorkBook
    activate WorkBookRecordService
    WorkBookRecordService ->> WorkBookRepository: AddSharedUser
    activate WorkBookRepository
    WorkBookRepository ->> DynamoDB: GetItem
    activate DynamoDB
    DynamoDB -->> WorkBookRepository: GetItemResponse
    deactivate DynamoDB
    WorkBookRepository ->> WorkBookRepository: CheckIfUserAlreadyHasAccess
    WorkBookRepository ->> DynamoDB: UpdateItem
    activate DynamoDB
    DynamoDB -->> WorkBookRepository: UpdateItemResponse
    deactivate DynamoDB
    WorkBookRepository -->> WorkBookRecordService: AddSharedUserResponse
    deactivate WorkBookRepository
    WorkBookRecordService -->> WorkBookHandler: AddNewUserToWorkBookResponse
    deactivate WorkBookRecordService
    WorkBookHandler -->> Client: ShareWorkBookResponse
    deactivate WorkBookHandler
```


### System Components

1. **Client Application**: Interacts with the system via HTTP endpoints.
2. **WorkBookHandler**: Processes HTTP requests and coordinates with the WorkBookRecordService.
3. **WorkBookRecordService**: Contains the business logic for managing workbooks and interacts with the WorkBookRepository.
4. **WorkBookRepository**: Interfaces with DynamoDB to perform CRUD operations.
5. **DynamoDB**: AWS DynamoDB is used as the database for storing workbook records.

### Detailed Component Design

#### 1. Client Application

The client application interacts with the system via HTTP endpoints. It sends requests to create, retrieve, share, and update workbook records.

#### 2. WorkBookHandler

The WorkBookHandler processes incoming HTTP requests and coordinates with the WorkBookRecordService. It includes the following endpoints:

- **CreateWorkBook**: Handles requests to create a new workbook.
- **RetrieveWorkBooks**: Handles requests to retrieve a user's workbooks.
- **RetrieveSharedWorkBooks**: Handles requests to retrieve workbooks shared with the user.
- **ShareWorkBook**: Handles requests to share a workbook with other users.

```mermaid
graph TD
    A[Client Application] --> B[WorkBookHandler]    
    subgraph WorkBookHandler
        B1[CreateWorkBook] --> C[WorkBookRecordService]
        B2[RetrieveWorkBooks] --> C
        B3[RetrieveSharedWorkBooks] --> C
        B4[ShareWorkBook] --> C
    end
```
### WorkBookRecordService
This service contains the business logic for managing workbooks. It interacts with the WorkBookRepository to perform database operations. Key methods include:

1. **InsertToWorkBookRecord**: Validates and inserts a new workbook record.
2. **RetrieveFromWorkBookRecords**: Retrieves all workbook records for a user.
3. **RetrieveSharedWorkBookRecords**: Retrieves workbooks shared with a user.
4. **AddNewUserToWorkBook**: Adds a new user to the shared workbook's access list.

```mermaid
graph TD
    C[WorkBookRecordService]    
    subgraph WorkBookRecordService
        C1[InsertToWorkBookRecord] --> D[WorkBookRepository]
        C2[RetrieveFromWorkBookRecords] --> D
        C3[RetrieveSharedWorkBookRecords] --> D
        C4[AddNewUserToWorkBook] --> D
    end
```

### WorkBookRepository
The repository interfaces with DynamoDB to perform CRUD operations. It includes:

1. **InsertNewWorkBookRecord**: Inserts a new workbook record into the DynamoDB table.
2. **RetrieveWorkBookRecords**: Retrieves workbook records based on user ID.
3. **RetrieveSharedWorkBookRecords**: Retrieves shared workbook records using filter expressions.
4. **AddSharedUser**: Updates a workbook record to add a new user to the shared access list.

```mermaid
graph TD
    D[WorkBookRepository]    
    subgraph WorkBookRepository
        D1[InsertNewWorkBookRecord] --> E[(DynamoDB)]
        D2[RetrieveWorkBookRecords] --> E
        D3[RetrieveSharedWorkBookRecords] --> E
        D4[AddSharedUser] --> E
    end
    
    subgraph DynamoDB
        E1[Insert Item]
        E2[Query Items]
        E3[Update Item]
        E4[Get Item]
    end

```


## Detailed Sequence of Operations

### Create WorkBook Operation

1. **Client Application** sends a request to **CreateWorkBook** endpoint in **WorkBookHandler**.
2. **WorkBookHandler** calls **InsertToWorkBookRecord** in **WorkBookRecordService**.
3. **WorkBookRecordService** calls **InsertNewWorkBookRecord** in **WorkBookRepository**.
4. **WorkBookRepository** uses DynamoDB **PutItem** operation to insert the new record.

### Retrieve WorkBooks Operation

1. **Client Application** sends a request to **RetrieveWorkBooks** endpoint in **WorkBookHandler**.
2. **WorkBookHandler** calls **RetrieveFromWorkBookRecords** in **WorkBookRecordService**.
3. **WorkBookRecordService** calls **RetrieveWorkBookRecords** in **WorkBookRepository**.
4. **WorkBookRepository** uses DynamoDB **Query** operation to retrieve the records.

### Retrieve Shared WorkBooks Operation

1. **Client Application** sends a request to **RetrieveSharedWorkBooks** endpoint in **WorkBookHandler**.
2. **WorkBookHandler** calls **RetrieveSharedWorkBookRecords** in **WorkBookRecordService**.
3. **WorkBookRecordService** calls **RetrieveSharedWorkBookRecords** in **WorkBookRepository**.
4. **WorkBookRepository** uses DynamoDB **Query** operation with a filter expression to retrieve the shared records.

### Share WorkBook Operation

1. **Client Application** sends a request to **ShareWorkBook** endpoint in **WorkBookHandler**.
2. **WorkBookHandler** calls **AddNewUserToWorkBook** in **WorkBookRecordService**.
3. **WorkBookRecordService** calls **AddSharedUser** in **WorkBookRepository**.
4. **WorkBookRepository** retrieves the workbook using DynamoDB **GetItem** operation.
5. **WorkBookRepository** updates the workbook's **SharedWith** attribute using DynamoDB **UpdateItem** operation.

## Conclusion

The detailed design and sequence of operations provide a comprehensive view of how the system components interact to manage workbook records in DynamoDB. Each component's responsibilities are clearly defined, ensuring a maintainable and scalable system architecture.
