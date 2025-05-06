# Function Naming Convention Guide (API) 

## Overview 
This document defines a standardized naming convention for CRUD functions across different layers of a module:  
- 1. **Repository** (Data access layer)  
- 2. **Service** (Business logic layer)  
- 3. **Handler** (API/Controller layer)  

Consistent naming improves code readability, maintainability, and team collaboration.  

---

## 1. Repository Layer 
Responsible for direct database/API interactions. 

| **Operation**  | **Correct Naming**      | **Avoid**               | **Example**           |
|---------------|------------------------|------------------------|-----------------------|
| **Create**    | `Create[Entity]`       | `Add[Entity]`          | `CreateUser`         |
| **Read (Single)** | `Get[Entity]ByID`  | `Retrieve[Entity]`, `Find[Entity]` | `GetUserByID`     |
| **Read (All)** | `GetAll[Entities]`  | `List[Entities]`, `FetchAll[Entities]` | `GetAllUsers`  |
| **Update**    | `Update[Entity]`       | `Modify[Entity]`       | `UpdateUser`        |
| **Delete**    | `Delete[Entity]`       | `Remove[Entity]`       | `DeleteUser`        |

**Key Notes:**  
- **`Create`** (not `Add`) emphasizes entity creation.  
- **`Get[Entity]ByID`** ensures consistency for single-record queries.  
- **`GetAll[Entities]`** (plural) clearly indicates multiple records.  

---

## 2. Service Layer 
Handles business logic and orchestrates repository calls.  

| **Operation**  | **Correct Naming**      | **Example**           |
|---------------|------------------------|-----------------------|
| **Create**    | `Create[Entity]`       | `CreateUser`         |
| **Read (Single)** | `Get[Entity]`      | `GetUser`           |
| **Read (All)** | `GetAll[Entities]`  | `GetAllUsers`       |
| **Update**    | `Update[Entity]`       | `UpdateUser`        |
| **Delete**    | `Delete[Entity]`       | `DeleteUser`        |

**Key Notes:**  
- Simpler than repositories (e.g., `GetUser` instead of `GetUserByID`).  
- Service methods may include business validation before calling repositories.  

---

## 3. Handler Layer 
API endpoints or controllers that call services. 

| **Operation**  | **Correct Naming**          | **Example**               |
|---------------|----------------------------|---------------------------|
| **Create**    | `Create[Entity]Handler`    | `CreateUserHandler`       |
| **Read (Single)** | `Get[Entity]Handler`    | `GetUserHandler`         |
| **Read (All)** | `GetAll[Entities]Handler` | `GetAllUsersHandler`     |
| **Update**    | `Update[Entity]Handler`    | `UpdateUserHandler`      |
| **Delete**    | `Delete[Entity]Handler`    | `DeleteUserHandler`      |

**Key Notes:**  
- Suffix with `Handler` for clarity (e.g., REST/gRPC endpoints).  
- Translates HTTP requests to service calls (e.g., extracts `ID` from URL).  

---

## Additional Best Practices
1. **Consistency is key**: Stick to one naming style across the codebase.  
2. **Avoid synonyms**: Use `Get` everywhere (not `Fetch`, `Retrieve`, etc.).  
3. **Pluralization**: Use `[Entities]` for lists (e.g., `GetAllUsers`).  
4. **Language-specific**: If using a non-English codebase, document deviations.  

---

## Examples 
*The following functions params are only used for example purposes.*

### **Repository**  
```go  
func CreateUser(ctx context.Context, user *User) error { ... }  
func GetUserByID(ctx context.Context, id string) (*User, error) { ... }
```

### **Service**  
```go  
func CreateUser(user *User) error { ... }  
func GetUser(id string) (*User, error) { ... }  
```

### **Handler**  
```go  
func CreateUserHandler(w http.ResponseWriter, r *http.Request) { ... }  
func GetUserHandler(w http.ResponseWriter, r *http.Request) { ... }  
```