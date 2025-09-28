# route-planner-api

This is the backend for the Route Planner project. It provides:

- User authentication (signup/login with JWT)
- Storing user-created trails
- Fetching trails and trailheads from the Overpass API with caching
- Tile caching for performance

---

## **Setup**

### **1. Install Go**
Ensure you have Go 1.21+ installed.

### **2. Install PostgreSQL**
- MacOS: `brew install postgresql`
- Linux: `sudo apt install postgresql postgresql-contrib`
- Windows: Use the [PostgreSQL installer](https://www.postgresql.org/download/windows/)

---

### **3. Create the Database**
```sql
CREATE DATABASE routeplanner;
CREATE USER myuser WITH PASSWORD 'mypassword';
GRANT ALL PRIVILEGES ON DATABASE routeplanner TO myuser;
\c routeplanner
CREATE EXTENSION postgis;
```

Update `.env` with your DB credentials:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=myuser
DB_PASS=mypassword
DB_NAME=routeplanner
JWT_SECRET=supersecretkey
```

---

### **4. Install Go dependencies**
```bash
go mod tidy
```

---

## **Running Locally**

```bash
go run main.go
```

Server runs on `http://localhost:8080`.

---

## **API Endpoints**

### **Authentication**
- `POST /api/auth/signup` – Create a new user
- `POST /api/auth/login` – Login and get JWT

### **User Trails**
- `GET /api/trails` – Get all trails for the logged-in user
- `POST /api/trails` – Save a new trail (JWT required)

### **Tiles**
- `GET /api/tiles/:z/:x/:y` – Fetch cached or fresh tile data from Overpass API

---

## **Frontend Integration**

- Use the JWT token in the `Authorization` header for authenticated endpoints:

```js
fetch("http://localhost:8080/api/trails", {
  method: "POST",
  headers: {
    "Authorization": `Bearer ${jwt}`,
    "Content-Type": "application/json"
  },
  body: JSON.stringify({ name: "My Trail", geojson: trailGeoJSON })
});
```

- Fetch tiles:

```js
fetch(`http://localhost:8080/api/tiles/${z}/${x}/${y}`)
  .then(res => res.json())
  .then(data => {
    // Use trail_geojson and trailhead_geojson
  });
```

---

## **Optional Improvements**
- Pre-fetch popular tiles
- Rate limiting for Overpass API
- HTTPS for production
- Unit tests for handlers and caching logic
- Redis for in-memory tile caching

---

## **Deployment**
- Dockerize the Go backend for services like Render, DigitalOcean, or AWS ECS
- Set environment variables for DB and JWT
- Can use Supabase for authentication and database if desired