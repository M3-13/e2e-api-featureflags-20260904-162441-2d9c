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

## Authentifizierung

Alle Endpunkte außer `GET /healthz` verlangen einen API-Key. Setzen Sie den Key
über die Umgebungsvariable `FLAG_API_KEY` und senden Sie ihn mit jedem Request im
`Authorization`-Header:

```
FLAG_API_KEY=<dein-key> go run .
```

```
GET /flags
Authorization: Bearer <dein-key>
```

Ist `FLAG_API_KEY` nicht gesetzt, erzeugt der Server einen zufälligen
ephemeren Key und loggt ihn einmal als Warnung — der Server startet niemals
ungeschützt.

## Betrieb & TLS

Standardmäßig bindet der Server an `127.0.0.1:8080`. Über die Umgebungsvariable
`ADDR` lässt sich die Bind-Adresse überschreiben:

```
ADDR=0.0.0.0:8080 go run .
```

Der Dienst stellt selbst kein TLS zur Verfügung. Öffentlich betreiben Sie ihn
ausschließlich hinter einem TLS-terminierenden Reverse Proxy oder API-Gateway,
das HTTPS abschließt und die Verbindung intern weiterleitet.

## Rate-Limiting

Der Dienst implementiert kein eigenes Rate-Limiting. Für den öffentlichen
Betrieb wird ein vorgelagertes API-Gateway mit Rate-Limiting empfohlen.

## Datenschutz / Verarbeitete Daten

Der Service verarbeitet den `user`-Parameter ausschließlich transient zur
Berechnung der Rollout-Entscheidung. Die Nutzerkennung wird weder persistiert
noch geloggt. In `key` und `description` dürfen keine personenbezogenen Daten
(PII) gespeichert werden. Rechtsgrundlage der Verarbeitung ist berechtigtes
Interesse bzw. Auftragsverarbeitung.

## Dependencies/SBOM

Das Projekt nutzt ausschließlich die Go-Standardbibliothek. Es sind keine
externen Drittanbieter-Abhängigkeiten enthalten.

## Updates & Version

Aktuelle Version: `1.0.0` (Konstante `ServiceVersion` in `main.go`). Für ein
Update die Versionsnummer erhöhen, `go build ./...` ausführen und das erzeugte
Binary neu deployen.

## Features

- Thread-sicherer In-Memory-Store für Feature-Flags
- Deterministische Rollout-Entscheidung (FNV-64a) für stabile Ergebnisse je Nutzer
- Einheitliches JSON-Fehlerformat über alle Endpunkte
- Zugriffs-Logging-Middleware (Methode, Pfad, Status, Dauer)
- API-Key-Authentifizierung (Bearer-Token) für alle Endpunkte außer `/healthz`
- Gehärteter HTTP-Server mit Read-/Write-/Idle-Timeout und begrenzter Header-Größe
- Sichere Defaults: Bind an `127.0.0.1`, kein ungeschützter Start
