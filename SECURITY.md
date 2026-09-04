VERDICT: BLOCKED

## Sicherheitsbericht

**Prüfumfang:** Go-Backend Feature-Flag-Service, NUR Standardbibliothek (`net/http`), keine anwendbaren Security-Scanner (Bandit/pip-audit/npm audit/semgrep nicht zutreffend). Es wurden ausschließlich die sichtbaren Quellcode-Dateien manuell bewertet; unbekannte Inhalte in nicht gezeigten Dateien (`AGENTS.md`, `README.md`, `SECURITY.md`, `COMPLIANCE.md`, `RUN.json`, `go.mod`) wurden nicht als verwundbar unterstellt.

---

### 1. HOCH — API-Schlüssel wird im Klartext geloggt

**Betroffene Datei/Stelle:** `main.go`, Zeile mit  
`log.Printf("WARNING: FLAG_API_KEY not set; generated ephemeral key %s", apiKey)`

**Befund:**  
Wenn die Umgebungsvariable `FLAG_API_KEY` nicht gesetzt ist, erzeugt der Server einen Bearer-Token und protokolliert ihn vollständig. Dieser Token ist die einzige Authentifizierung der gesamten API. Personen mit Lesezugriff auf Logdateien oder Log-Aggregationssysteme können damit während der Laufzeit des Prozesses alle Feature-Flags lesen, anlegen, ändern und löschen.

**Risiko:**  
- Credential-Leak in Logs (Bearer-Token = voller API-Zugriff).  
- Insbesondere in Container-/Cloud-Deployments werden stdout-Logs zentral gesammelt; der Schlüssel wird so persistent gespeichert.  
- Die Standardbindung `127.0.0.1:8080` mildert das Risiko nur bei rein lokaler Nutzung, ist aber kein Schutz bei üblichen Deployments mit `ADDR=0.0.0.0:8080`.

**Konkreter Fix:**  
Nicht den vollständigen Schlüssel loggen. Beispiel:

```go
if apiKey == "" {
    buf := make([]byte, 32)
    if _, err := rand.Read(buf); err != nil {
        log.Fatalf("failed to generate API key: %v", err)
    }
    apiKey = hex.EncodeToString(buf)
    log.Printf("WARNING: FLAG_API_KEY not set; using an ephemeral key. Set FLAG_API_KEY to a stable secret.")
}
```

Alternativ, sofern der automatisch generierte Schlüssel für den Betrieb zwingend bekannt sein muss, diesen in eine separate Datei mit Berechtigung `0600` schreiben und den Pfad loggen, nicht den Schlüssel selbst. Noch sicherer: bei fehlender `FLAG_API_KEY` den Start mit klarer Fehlermeldung verweigern (fail closed).

---

### 2. NIEDRIG — JSON-Body mit zusätzlichen Daten wird als gültig akzeptiert

**Betroffene Dateien/Stellen:**  
`internal/handlers/create.go`, `internal/handlers/mutate.go`  
jeweils der Aufruf `json.NewDecoder(r.Body).Decode(...)`

**Befund:**  
Nach dem ersten erfolgreichen `Decode` wird nicht geprüft, ob der Request-Body vollständig konsumiert wurde. Ein Request wie  
`{"key":"x","enabled":true} trailing-garbage`  
wird als gültig akzeptiert, obwohl es sich nicht um ein einzelnes, vollständig valides JSON-Objekt handelt.

**Risiko:**  
- Schwächere Eingabevalidierung als in den Acceptance Criteria angenommen („gültiges JSON“).  
- Kann in Kombination mit unsauberen Proxys oder zukünftigen Erweiterungen zu inkonsistentem Request-Handling führen.  
- Derzeit kein direkt ausnutzbarer RCE-/Injection-Befund, daher niedrig.

**Konkreter Fix:**  
Nach dem `Decode` prüfen, ob der Body beendet ist:

```go
dec := json.NewDecoder(r.Body)
if err := dec.Decode(&req); err != nil {
    // bestehende MaxBytesError-/400-Behandlung
}
if dec.More() {
    writeError(w, http.StatusBadRequest, "invalid JSON")
    return
}
```

Für `mutate.go` sinngemäß gleich.

---

### 3. NIEDRIG — Klartext-HTTP bei externem `ADDR`

**Betroffene Datei/Stelle:** `main.go`, Serverkonfiguration und `ADDR`-Auswertung

**Befund:**  
Der Server verwendet kein TLS. Der Bearer-Token wird im `Authorization`-Header übertragen. Der Standard `127.0.0.1:8080` ist unkritisch, sobald `ADDR` jedoch auf eine nicht-lokale Schnittstelle gesetzt wird, kann der Token im Netzwerk mitgelesen werden.

**Risiko:**  
- Abhören des Bearer-Tokens bei Deployment ohne TLS-Terminierung.  
- Kein Codefehler im engeren Sinne, aber eine gefährliche Konfigurationsmöglichkeit.

**Konkreter Fix:**  
- In der Dokumentation (`README.md`/`SECURITY.md`) klar vorgeben, dass die API ausschließlich hinter TLS/HTTPS (z. B. Reverse Proxy) bereitgestellt werden darf.  
- Optional: Standard-`ADDR` weiterhin nur auf `127.0.0.1:8080` belassen und bei davon abweichender Konfiguration einen deutlichen Warnhinweis loggen.

---

### 4. NIEDRIG — Log-Injection über URL-Pfad möglich

**Betroffene Datei/Stelle:**  
`internal/middleware/middleware.go`  
`log.Printf("%s %s %d %s", r.Method, r.URL.Path, rec.status, time.Since(start))`

**Befund:**  
`r.URL.Path` kann bei percent-encodeten Eingaben Steuerzeichen wie `%0A` (Zeilenumbruch) enthalten. Der Pfad wird mit `%s` unverändert ausgegeben, wodurch ein Angreifer die Log-Ausgabe um zusätzliche Zeilen erweitern und Log-Einträge fälschen kann.

**Risiko:**  
- Gefälschte Log-Einträge können forensische Auswertungen stören.  
- Das Logging erfüllt zwar AC-16/AC-18 (kein Query-String), ist aber nicht gegen Steuerzeichen gehärtet.

**Konkreter Fix:**  
Pfad vor der Ausgabe bereinigen oder escapend formatieren:

```go
cleanPath := strings.Map(func(r rune) rune {
    if unicode.IsControl(r) {
        return -1
    }
    return r
}, r.URL.Path)

log.Printf("%s %s %d %s", r.Method, cleanPath, rec.status, time.Since(start))
```

Alternativ `%q` für den Pfad verwenden, sofern die Log-Ausgabe dadurch für die bestehende Werkzeugkette lesbar bleibt.

---

## Ergebnis

Das Produkt erfüllt viele Security-Vorgaben (Timeouts, MaxBytesReader, Content-Type-Prüfung, constant-time API-Key-Vergleich, kein Query-String im Log, thread-sicherer Store, keine externen Dependencies).  
Der Klartext-Log des vollständigen Bearer-Tokens (`FLAG_API_KEY`-Fallback) ist jedoch ein hohes Risiko und muss vor Auslieferung behoben werden. Daher: **BLOCKED**.