# Feature-Flag-Service (Go REST-API)

Eine REST-API in Go, die Feature-Flags in einem thread-sicheren In-Memory-Store
verwaltet und deterministische Nutzer-Rollout-Entscheidungen liefert. Das Projekt
kommt ohne externe Web-Framework-Abhängigkeiten aus und nutzt ausschließlich die
Standardbibliothek (`net/http`).

## Tech-Stack

- **Sprache:** Go (1.22+)
- **Framework:** `net/http` (Standardbibliothek, `http.ServeMux` mit Methoden- und Wildcard-Patterns)
- **Concurrency:** `sync.Mutex` / `sync.RWMutex`
- **Tests:** `testing` + `net/http/httptest`

## Installation

Das Projekt hat keine externen Abhängigkeiten. Voraussetzung ist lediglich eine
Go-Installation (1.22 oder neuer).

```
git clone <repository-url>
cd <repository>
```

## Start (Entwicklung)

```
go run .
```

Der Server startet auf Port `8080`.

## Build (Produktion)

```
go build ./...
```

Das erzeugte Binary startet den Server auf Port `8080`.

## Tests

```
go test ./...
```

## Endpunkte

Alle Antworten sind JSON mit `Content-Type: application/json`. Fehlerantworten
haben durchgängig die Form `{"error": "<meldung>"}`.

| Methode | Pfad | Beschreibung |
|---|---|---|
| `POST` | `/flags` | Legt ein Flag an (`{"key","enabled","description?","rollout_percent?"}`) → `201` |
| `GET` | `/flags` | Listet alle Flags → `200` |
| `GET` | `/flags/{key}` | Liefert ein Flag → `200` |
| `PUT` | `/flags/{key}` | Aktualisiert ein Flag partiell → `200` |
| `DELETE` | `/flags/{key}` | Entfernt ein Flag → `204` |
| `GET` | `/flags/{key}/evaluate?user={id}` | Rollout-Entscheidung für einen Nutzer → `200` |
| `GET` | `/healthz` | Health-Check → `200` mit `{"status":"ok"}` |

### Beispiel: Health-Check

```
GET /healthz
```

Antwort:

```json
{"status":"ok"}
```

## Features

- Thread-sicherer In-Memory-Store für Feature-Flags
- Deterministische Rollout-Entscheidung (FNV-64a) für stabile Ergebnisse je Nutzer
- Einheitliches JSON-Fehlerformat über alle Endpunkte
- Zugriffs-Logging-Middleware (Methode, Pfad, Status, Dauer)
- Gehärteter HTTP-Server mit Read-/Write-/Idle-Timeout und begrenzter Header-Größe
