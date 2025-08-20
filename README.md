## 🩺 Consultation Diagnostics & Treatments API

These endpoints allow you to **add** and **retrieve** diagnostics (with treatments) for a given consultation in a single request.

---

### **GET** `/consultations/:consultation_id/diagnostics`

**Description:**  
Fetch all diagnostics for a consultation, including their treatments.

**Example Request:**
```bash
curl -X GET "http://localhost:4000/consultations/1/diagnostics" \
  -H "Accept: application/json"
```

**Example Response:**
```json
[
  {
    "id": 1,
    "name": "Hipertensión ocular",
    "recommendation": "Control cada 6 meses, medición PIO",
    "consultation_id": 1,
    "treatments": [
      {
        "id": 5,
        "diagnostic_id": 1,
        "active_component": "Timolol maleato",
        "presentation": "Solución oftálmica",
        "dosage": "0.5",
        "frequency": "12:00:00",
        "duration": "720:00:00"
      }
    ]
  }
]
```

---

### **POST** `/consultations/:consultation_id/diagnostics`

**Description:**  
Add one or more diagnostics (with optional treatments) to a consultation in a single request.  
All inserts are wrapped in a transaction — if one fails, nothing is saved.

**Request Body:**  
- **Array** of diagnostics
- Each diagnostic can have a `treatments` array
- All fields are strings except `consultation_id` (from URL)

**Example Request:**
```bash
curl -X POST "http://localhost:4000/consultations/1/diagnostics" \
  -H "Content-Type: application/json" \
  -d '[
    {
      "name": "Migraine",
      "recommendation": "Avoid bright lights and loud noises",
      "treatments": [
        {
          "active_component": "Ibuprofen",
          "presentation": "Tablet",
          "dosage": "400mg",
          "frequency": "Every 8 hours",
          "duration": "5 days"
        }
      ]
    }
  ]'
```

**Example Response:**
```json
{
  "status": "created"
}
```

---

### Notes
- `dosage`, `frequency`, and `duration` are stored as **TEXT** for flexibility (e.g., `"400mg"`, `"Every 8 hours"`, `"As needed"`).
- You can send multiple diagnostics in one POST.
- `treatments` can be an empty array if no treatments are needed.
- GET always returns diagnostics **with** their treatments nested.
