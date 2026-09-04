VERDICT: CHANGES_REQUESTED

## Strukturierter Compliance-Bericht

**Projekttyp:** `go-backend` – reine Backend-API ohne Endnutzer-UI.  
**Prüfumfang:** Sichtbarer Code und Sprint-Spezifikation. Nicht eingesehene Dateien (z. B. `README.md`, `COMPLIANCE.md`, `SECURITY.md`) wurden nur als vorhanden registriert, konnten inhaltlich aber nicht bewertet werden.

---

### 1. DSGVO / GDPR

**Kernbefund:**
- Es entsteht nur sporadisch ein personenbezogenes Datum: der `user`-Parameter in `GET /flags/{key}/evaluate`.  
- Dieser Parameter wird ausschließlich transient in der Hash-Berechnung verwendet, nicht gespeichert und durch die Logging-Middleware (`internal/middleware/middleware.go`) nicht in Logs geschrieben, weil ausschließlich `r.URL.Path` ohne Query-String protokolliert wird. Das erfüllt AC-16/AC-18 und ist datenschutzfreundlich.

**Findings:**

| Schweregrad | Befund | Konkrete Abhilfe |
|---|---|---|
| **Mittel** | Rechtsgrundlage / Datenschutzhinweis für die Verarbeitung des `user`-Parameters ist im sichtbaren Code nicht nachgewiesen. Für eine API kann das über die Betreiberdokumentation erfolgen, aber diese liegt nicht im geprüften Stand vor. | In `README.md` oder `COMPLIANCE.md` einen Abschnitt „Datenverarbeitung“ ergänzen: Benannte Rechtsgrundlage (z. B. Art. 6 Abs. 1 lit. b DSGVO für Vertragserfüllung oder lit. f für berechtigtes Interesse), Hinweis auf fehlende persistente Speicherung von `user`-Werten und Kontakt für Betroffenenrechte. |
| **Mittel** | Transportverschlüsselung (TLS) wird im Code nicht erzwungen. Der Server startet per Default auf `127.0.0.1:8080` (sicher), aber `ADDR` kann auf eine öffentliche Adresse gesetzt werden. Dann würden API-Key und `user`-Parameter unverschlüsselt übertragen. | Entweder in `main.go` auf `ListenAndServeTLS()` umstellen, wenn öffentlich exponiert, oder in `SECURITY.md`/`README.md` verbindlich dokumentieren, dass der Dienst ausschließlich hinter einem TLS-terminierenden Reverse-Proxy bereitgestellt werden darf. |

**Positiv:**  
- Keine persistente Speicherung personenbezogener Daten.  
- Keine Protokollierung des `user`-Parameters.  
- `Cache-Control: no-store` auf allen JSON-Antworten verhindert unerwünschtes Caching.  
- Datenminimierung durch `http.MaxBytesReader` und gezielte Validierung.

---

### 2. EU Cyber Resilience Act (CRA)

**Kernbefund:**  
Der Code enthält wichtige Security-by-Design-Maßnahmen: Timeouts, maximale Body-Größe, Content-Type-Prüfung, API-Key-Authentifizierung, kein Logging sensibler Daten. Sichtbare Lücken betreffen vor allem Dokumentation und Umfeld.

| Schweregrad | Befund | Konkrete Abhilfe |
|---|---|---|
| **Mittel** | Kein Schutz gegen Brute-Force-Angriffe auf den API-Key. `internal/middleware/auth.go` prüft nur den Token, limitiert aber keine Fehlversuche. | In `main.go` oder der Middleware ein einfaches Rate-Limiting (z. B. pro Client-IP, 10 Fehlversuche / Minute) ergänzen oder dokumentierte Empfehlung für einen starken, rotierbaren API-Key in `SECURITY.md`. |
| **Niedrig** | SBOM / Abhängigkeitsdokumentation ist im sichtbaren Stand nicht vorhanden. Da `go.mod` auf die Standardbibliothek und minimale Module zu beschränken scheint, ist das Risiko gering, aber für CRA-Konformität sollte eine SBOM vorgehalten werden. | `go.mod`-Inhalt prüfen und z. B. mit `syft` oder `govulncheck` eine Software Bill of Materials (CycloneDX/SPDX) erzeugen und im Repo ablegen. |
| **Niedrig** | Update-/Patch-Prozess ist im sichtbaren Code nicht beschrieben. CRA verlangt eine klar beschriebene Update-Fähigkeit. | In `SECURITY.md` oder `README.md` den Prozess zur Aktualisierung des Dienstes (Binary-Release, Rollback) ergänzen. |

**Positiv:**  
- `ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `MaxHeaderBytes` korrekt gesetzt.  
- `MaxBytesReader` verhindert übermäßige Request-Bodys.  
- Kein PII-Leak im Logging.  
- `subtle.ConstantTimeCompare` für Token-Vergleich.

---

### 3. EU AI Act

**Nicht anwendbar.**  
Im sichtbaren Produkt befindet sich keine KI-Funktion. Der Feature-Flag-Service nutzt lediglich deterministische Hash-Berechnung (`FNV-64a`); dies fällt nicht unter den KI-Begriff.

---

### 4. Pflichttexte und UI

**Nicht anwendbar.**  
Es handelt sich um eine reine REST-API ohne Web-UI, daher keine Impressums-, Cookie- oder Rechtstextpflichten im Web-Interface.

---

### 5. Barrierefreiheit (WCAG / BITV / EAA)

**Nicht anwendbar.**  
Keine öffentliche Web-UI, daher keine spezifischen Barrierefreiheitspflichten.

---

## Gesamteinschätzung

Die funktionalen und sicherheitstechnischen Anforderungen der Sprint-Spezifikation sind überwiegend erfüllt. Es bestehen **keine fundamentalen rechtlichen Verstöße** (daher nicht `BLOCKED`), jedoch mehrere **behebbare Lücken** insbesondere im Bereich Datenschutznachweis, Transportverschlüsselung und CRA-Dokumentation. Nach Behebung dieser Punkte kann das Produkt als marktreif für den vorgesehenen Backend-Einsatz eingestuft werden.