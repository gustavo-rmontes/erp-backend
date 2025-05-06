# Module Structure - Directories and Files

## Overview
This document defines the standard structure for organizing Go modules, with two variations based on module complexity:
1. **Full structure** (for modules with many files)
2. **Simplified structure** (for small modules)

---

## 1. Full Structure
Recommended for modules with many `.go` files, including separated test directories.
```text
./backend
└── internal
    └── modules
        └── [module_name]          # Ex: sales, inventory, users
            ├── models
            │   ├── [entity].go    # Domain models
            │   ├── enums.go       # Related enumerations
            │   └── tests/        # Model tests (if needed)
            ├── repository
            │   ├── [entity]_repository.go
            │   └── tests/         # Repository tests
            ├── service
            │   ├── [entity]_service.go
            │   └── tests/        # Service tests
            └── handler
                ├── [entity]_handler.go
                └── tests/         # Handler tests
```

Example for Sales Module
```text
./backend
└── internal
    └── modules
        └── sales
            ├── models
            │   ├── quotation.go
            │   ├── sales_order.go
            │   ├── purchase_order.go
            │   ├── delivery.go
            │   ├── invoice.go
            │   ├── enums.go
            │   └── tests/
            │       ├── quotation_test.go
            │       ├── sales_order_test.go
            │       └── invoice_test.go
            ├── repository
            │   ├── quotation_repository.go
            │   ├── sales_order_repository.go
            │   └── invoice_repository.go
            ├── service
            │   ├── quotation_service.go
            │   └── sales_order_service.go
            └── handler
                ├── quotation_handler.go
                └── sales_order_handler.go
```
## 2. Simplified Structure
For small modules (less than 3-4 files per layer), tests can remain in the same directory:
``` text
./backend
└── internal
    └── modules
        └── [module_name]
            ├── models
            │   ├── [entity].go
            │   └── [entity]_test.go  # Tests with models
            ├── repository
            │   ├── [entity]_repository.go
            │   └── [entity]_repository_test.go
            ├── service
            │   ├── [entity]_service.go
            │   └── [entity]_service_test.go
            └── handler
                ├── [entity]_handler.go
                └── [entity]_handler_test.go
```

## Best Practices
### 1. Consistency 
- Maintain the same structure across all project modules

### 2. Testing
- Use /tests/ directory when having 4+ test files
- Keep _test.go files in the same directory for small modules

### 3. Naming
- Always use snake_case for filenames
- File prefix should match the primary entity