# Demo Application

This is a demo application showcasing the usage of the `go-di` dependency injection library. The application provides a simple REST API for managing TODO items.

## Endpoints

### 1. Create a TODO
**URL:** `/todos`  
**Method:** `POST`  
**Description:** Creates a new TODO item.  
**Request Body:**
```json
{
  "title": "Sample TODO",
  "description": "This is a sample TODO item."
}
```
**Example:**
```bash
curl -X POST http://localhost:8080/todos \
-H "Content-Type: application/json" \
-d '{ "title": "Sample TODO", "description": "This is a sample TODO item." }'
```

### 2. Get All TODOs
**URL:** `/todos`  
**Method:** `GET`  
**Description:** Retrieves a list of all TODO items.  
**Example:**
```bash
curl -X GET http://localhost:8080/todos
```

### 3. Mark TODO as Done
**URL:** `/todos/{id}/done`  
**Method:** `PATCH`  
**Description:** Marks a TODO item as done.  
**Path Parameter:** `id` - The ID of the TODO item to mark as done.  
**Request Body:**
```json
{
  "id": 1715211870
}
```
**Example:**
```bash
curl -X PATCH http://localhost:8080/todos/1715211870/done \
-H "Content-Type: application/json" \
-d '{ "id": 1715211870 }'
```