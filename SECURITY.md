VERDICT: BLOCKED

## Scanner-Lage
Für dieses Go-Projekt wurden keine Security-Scanner ausgeführt bzw. als nicht anwendbar markiert. Das ist eine Prüflücke, aber kein eigenständiger Befund. Meine Bewertung erfolgt aus der sichtbaren Code- und Laufzeitanalyse.

## Sicherheitsreport

### 1. Fehlende Authentifizierung/Autorisierung für die gesamte Feature-Flag-API — HOCH
**Betroffene Stellen:**  
- `main.go` — HTTP-Server ohne Auth-Middleware  
- `internal/handlers/handlers.go` — alle `/flags*`-Routen ungeschützt registriert  
- `internal/handlers/create.go`, `mutate.go`, `read.go`, `evaluate.go`

**Risiko:**  
Jeder, der den Port erreichen kann, darf Flag-Definitionen anlegen, ändern, löschen und auslesen. Ein Angreifer kann damit die Business-Logik steuern: Flags aktivieren/deaktivieren, Rollout-Prozentsätze auf 100 setzen oder Flags gezielt entfernen. Dies ist ein vollständiger Integritäts- und Informationsverlust für den Feature-Flag-Dienst. Da Mutationen kritische Anwendungskonfiguration betreffen, ist der fehlende Zugriffsschutz als hohes Risiko einzustufen.

**Konkreter Fix:**  
- Eine Auth-Middleware vor alle `/flags*`-Routen schalten, die einen konfigurierten API-Key (z. B. `Authorization: Bearer <key>` oder `X-API-Key`) prüft.  
  Beispiel:
  ```go
  apiKey := os.Getenv("FLAG_API_KEY")
  if apiKey == "" {
      log.Fatal("FLAG_API_KEY must be set")
  }
  mux.Handle("/flags", authMiddleware(apiKey, handlers...))
  ```
- `/healthz` darf ohne Auth erreichbar bleiben, damit Health-Checks funktionieren.  
- Der Server sollte standardmäßig auf `127.0.0.1:8080` binden; ein öffentlicher Bind nur durch explizite Konfiguration, nicht als Default.  
- Alternativ: Der Dienst muss zwingend hinter einem vorgelagerten Authentifizierungs-/Netzwerk-Gateway betrieben werden. Das ist im Code nicht erzwingbar und sollte dokumentiert plus per Default-Portbindung abgesichert werden.

---

### 2. Unbegrenzter In-Memory-Store ermöglicht Ressourcenerschöpfung — HOCH
**Betroffene Stellen:**  
- `internal/store/store.go` — `Store.Create` ohne Mengenbegrenzung  
- `internal/handlers/create.go` — Body-Limit 1 MiB pro Request, aber keine Begrenzung der Gesamtanzahl oder Gesamtgröße

**Risiko:**  
Ein Angreifer kann massenhaft `POST /flags` mit vielen eindeutigen Keys und großen Beschreibungen senden. Jede Beschreibung darf bis zu ca. 1 MiB groß sein; der Store wächst unbegrenzt, bis der Prozess “out of memory” läuft. Zusammen mit der fehlenden Authentifizierung ist dies ein einfacher DoS-Vektor.

**Konkreter Fix:**  
- Konfigurierbares Maximum für die Anzahl der Flags einführen, z. B. `maxFlags int` in `Store.New`.  
- In `Store.Create` bei Erreichen des Limits einen Fehler wie `ErrTooManyFlags` zurückgeben; im Handler daraus eine `429 Too Many Requests` oder `507 Insufficient Storage` machen.  
- Optional zusätzlich ein einfaches Rate-Limit pro Client/IP für schreibende Routen (`POST`, `PUT`, `DELETE`), um massenhafte Mutationen zu drosseln.

---

### 3. Server bindet an alle Interfaces und ohne TLS — MITTEL
**Betroffene Stelle:**  
- `main.go` — `Addr: ":8080"`, kein `TLSConfig`

**Risiko:**  
`:8080` bindet an alle Netzwerk-Interfaces. Sofern keine Firewall vorgelagert ist, ist der Dienst öffentlich im Klartext erreichbar. In Kombination mit Finding 1 können Angreifer die gesamte API unverschlüsselt manipulieren. Auch ohne Auth ist Klartext-Übertragung bei einem Konfigurationsdienst ein unnötiges Risiko.

**Konkreter Fix:**  
- Standard-Adresse auf `127.0.0.1:8080` setzen, öffentlichen Bind nur über eine Umgebungsvariable wie `ADDR` zulassen.  
- TLS entweder direkt im Server (`ListenAndServeTLS`) oder über einen TLS-terminierenden Reverse Proxy erzwingen.  
- Sicherstellen, dass bei einem Proxy-Setup nicht blind `X-Forwarded-*`-Header vertraut werden, sofern Client-IP-abhängige Limits genutzt werden.

---

### 4. POST /flags prüft Content-Type nicht MIME-konform — NIEDRIG
**Betroffene Stelle:**  
- `internal/handlers/create.go` — `if ct := r.Header.Get("Content-Type"); ct != "application/json"`

**Risiko:**  
Requests mit gültigem Content-Type wie `application/json; charset=utf-8` werden mit `415 Unsupported Media Type` abgelehnt. Das ist primär ein Kompatibilitätsproblem, kein ausnutzbarer Sicherheitsbruch. Es kann aber dazu führen, dass Clients Workarounds mit exakt gleichem String verwenden; ein Sicherheitsrisiko entsteht daraus nicht unmittelbar.

**Konkreter Fix:**  
Wie bereits in `mutate.go` vorhanden:  
```go
mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
if err != nil || mediaType != "application/json" { ... }
```
Dadurch bleiben strenge Prüfung und Standardkonformität erhalten.

---

## Zusammenfassung
Die umgesetzten Härtungsmaßnahmen (Timeouts, MaxBytesReader, Content-Type-Prüfung, Logging ohne Query-String) sind korrekt und erfüllen die vorgegebenen Security-ACs. Die schwerwiegende Lücke ist die vollständig fehlende Zugriffskontrolle: Ohne Authentifizierung kann jeder erreichbare Client die Feature-Flags uneingeschränkt verändern. Das ist für einen Konfigurationsdienst mit direkter Auswirkung auf Anwendungsverhalten ein hohes Risiko. Zusätzlich besteht ein einfacher DoS-Vektor durch unbegrenzten Speicherverbrauch. Deshalb: **BLOCKED**.