VERDICT: APPROVED

## Sicherheitsprüfung des Feature-Flag-Service

Es lag kein anwendbarer Security-Scanner-Output vor (`no applicable security scanners for this project type`). Das Fehlen eines Scanner-Ergebnisses wurde nicht als Schwachstelle gewertet; die Prüfung erfolgte manuell gegen den gelieferten Quellcode.

### Geprüfte Bereiche

- **Secrets:** Es wurden keine hartkodierten Schlüssel, Passwörter oder Token im Produktcode gefunden. `FLAG_API_KEY` wird in `main.go` ausschließlich aus der Umgebung gelesen; bei fehlendem Wert beendet sich der Dienst.
- **Injection/Inputs:** Keine SQL-, Command- oder Path-Injection erkennbar. JSON-Bodies werden größenbegrenzt, `Content-Type` wird erzwungen, ungültige Werte werden mit 400/413/415 und einheitlichem JSON-Fehlerobjekt beantwortet. Es gibt keine HTML-UI, daher keine direkte XSS-Fläche.
- **AuthN/AuthZ:** Alle produktiven Routen außer `/healthz` sind durch einen Bearer-Token geschützt. Der Vergleich erfolgt über `crypto/subtle.ConstantTimeCompare`. Ein Auth-Bypass ist nicht erkennbar.
- **Dependencies:** Das Projekt nutzt ausschließlich die Go-Standardbibliothek. Es sind keine externen, potenziell angreifbaren Pakete sichtbar.
- **Konfiguration/Transport:** `ReadTimeout`, `WriteTimeout`, `IdleTimeout` und `MaxHeaderBytes` sind gesetzt. `Cache-Control: no-store` wird für JSON-Antworten gesetzt. Der Logging-Middleware-Formatstring ist statisch; der Pfad wird als Argument übergeben und von Steuerzeichen bereinigt.

### Feststellungen

#### LOW-01: HTTP-Transport ohne TLS bei nicht-lokalem Bind

- **Schwere:** low
- **Betroffene Stelle:** `main.go:41-46`, `internal/middleware/auth.go`
- **Beschreibung:** Der Server startet ausschließlich mit `ListenAndServe`. Der Standard-Bind `127.0.0.1:8080` ist unkritisch. Wird `ADDR` jedoch auf `0.0.0.0:8080` oder eine öffentliche Adresse gesetzt, wird der Bearer-API-Key im Klartext übertragen.
- **Konkrete Empfehlung:** Für nicht-lokale Bindings einen TLS-terminierenden Reverse-Proxy vorschreiben oder optional `ListenAndServeTLS` mit Zertifikatspfaden aus der Umgebung implementieren. Loopback-Betrieb kann unverändert über HTTP laufen, damit lokale Entwicklung und Tests funktionsfähig bleiben.

#### LOW-02: Keine Mindeststärke für `FLAG_API_KEY`

- **Schwere:** low
- **Betroffene Stelle:** `main.go:17-20`
- **Beschreibung:** Es wird nur geprüft, dass der API-Key nicht leer ist. Ein sehr kurzer oder trivialer Schlüssel kann per Online-Brute-Force gegen die geschützten Endpunkte erraten werden.
- **Konkrete Empfehlung:** Mindestlänge (z. B. 32 Zeichen) und idealerweise eine hohe Entropie erzwingen. Erzeugung über ein Secret-Management dokumentieren; Rotation ermöglichen. Kein Hardcode im Repository.

#### LOW-03: Feldlängen für `key` und `description` nicht begrenzt

- **Schwere:** low
- **Betroffene Stelle:** `internal/handlers/create.go`, `internal/handlers/mutate.go`, `internal/model/model.go`
- **Beschreibung:** Das Body-Limit von 1 MiB ist korrekt umgesetzt. Da `key` und `description` aber keine eigenen Längenlimits haben, könnte ein authentifizierter Client im Extremfall bis zu 1000 Flags mit jeweils nahezu 1 MiB großen Feldern anlegen und so erheblichen Speicher belegen.
- **Konkrete Empfehlung:** Zusätzlich fachliche Limits validieren, z. B. `key` maximal 256 Zeichen und `description` maximal 4096 Zeichen. Bei Überschreitung 400 mit JSON-Fehlerobjekt zurückgeben. Das bestehende 1-MiB-Body-Limit bleibt unverändert bestehen.

#### LOW-04: Kein Rate-Limiting an API-/Auth-Schicht

- **Schwere:** low
- **Betroffene Stelle:** `main.go:32-34`, `internal/middleware/auth.go`
- **Beschreibung:** Fehlerhafte Authentifizierungsversuche und wiederholte API-Aufrufe werden nicht begrenzt. Dadurch sind Brute-Force-Versuche gegen den API-Key und im weiteren Sinne API-basierte Überlastungen nicht auf Anwendungsebene gedrosselt.
- **Konkrete Empfehlung:** Eine Rate-Limit-Middleware oder ein vorgeschaltetes Gateway einführen, z. B. Limits pro Client-IP oder API-Key. Alternativ bei wiederholten 401-Antworten eine kurze Verzögerung einbauen. Dabei Limits praxistauglich wählen, damit reguläre Nutzung der Flag-API nicht blockiert wird.

### Ergänzender Hinweis

Der Testcode `internal/handlers/evaluate_test.go` verwendet `unsafe`/Reflection, um die private Store-Map zu befüllen. Das ist ausschließlich Testcode, wird nicht deployed und begründet keine Produktschwachstelle.

### Fazit

Es wurden keine ausnutzbaren Schwachstellen hoher oder mittlerer Schwere festgestellt. Die geforderten Security-Maßnahmen aus dem Sprint sind im sichtbaren Produktcode umgesetzt. Die verbleibenden Punkte sind Härtungsempfehlungen mit niedrigem Risiko.