VERDICT: CHANGES_REQUESTED

## Prüfbericht: Feature-Flag-Service (Go-REST-API)

Projekttyp: `go-backend` — reine Backend-API ohne Endnutzer-UI. Keine Pflicht zu Impressum/Cookie-Banner/DSGVO-Webseite/Barrierefreiheit. Relevante Prüfbereiche: **DSGVO** (transiente Verarbeitung des `user`-Parameters, Logging, Datenminimierung) und **EU Cyber Resilience Act** (Security by Design/Default, Transportverschlüsselung, Zugriffsschutz, Update-Fähigkeit). Der **EU AI Act** ist nicht einschlägig, da keine KI-Funktion enthalten ist.

---

## 1. DSGVO

### 1.1 Protokollierung personenbezogener Daten
**Befund:**  
Die Logging-Middleware (`internal/middleware/middleware.go`) verwendet `r.URL.Path` ohne Query-String. Der `user`-Parameter aus `GET /flags/{key}/evaluate?user=…` erscheint daher nicht im Anwendungslog.  
**Schweregrad:** niedrig  
**Bewertung:** Datenschutzfreundlich. Keine PII im Log.  
**Remedy:** Keine Änderung erforderlich. Optional: Log-Format dokumentieren, damit Betreiber keine Query-Strings über vorgelagerte Proxys erfassen.

### 1.2 Transiente Verarbeitung des `user`-Parameters
**Befund:**  
Der `user`-Parameter wird in `internal/handlers/evaluate.go` gelesen und an `rollout.Evaluate` übergeben. Er wird **nicht gespeichert, nicht geloggt und nicht in der Antwort wiederholt**. Die Hash-Berechnung ist deterministisch, aber der Hash selbst wird nicht gespeichert.  
**Schweregrad:** niedrig  
**Bewertung:** Die Datenminimierung ist grundsätzlich eingehalten. Es fehlt jedoch eine transparente Dokumentation der Rechtsgrundlage und des Verarbeitungszwecks für Betreiber.  
**Remedy:** In `README.md` einen Abschnitt „Datenschutz / Verarbeitete Daten“ ergänzen:  
- Verarbeitete Daten: Feature-Flag-Key, optionale Description, Rollout-Prozentsatz, transient übergebene Nutzerkennung (`user`).  
- Rechtsgrundlage: berechtigtes Interesse (Art. 6 Abs. 1 lit. f DSGVO) bzw. Auftragsverarbeitung, wenn die API im Kundenauftrag betrieben wird.  
- Speicherdauer: alle Flags und der `user`-Parameter **nur im flüchtigen Arbeitsspeicher**; keine Persistenz, keine Protokollierung der Nutzerkennung.  
- Hinweis: **keine personenbezogenen Daten** in `description` oder `key` ablegen.

### 1.3 Cache-Verhalten von Evaluationsergebnissen
**Befund:**  
`writeJSON` und `writeError` setzen nur `Content-Type`, nicht jedoch `Cache-Control`. Browser oder Zwischenproxys könnten die Antwort von `GET /flags/{key}/evaluate` zwischenspeichern. Das Ergebnis kann Rückschlüsse auf eine bestimmte Nutzerkennung zulassen (z. B. ob ein Feature für eine Nutzergruppe aktiv ist).  
**Schweregrad:** mittel  
**Bewertung:** Verstoß gegen den Grundsatz der Vertraulichkeit (Art. 5 Abs. 1 lit. f DSGVO) und unsicherer Default.  
**Remedy:** In `internal/handlers/handlers.go` die Funktion `writeJSON` um folgende Zeile ergänzen:  
`w.Header().Set("Cache-Control", "no-store")`  
Analog auch für `writeError` (falls nicht über `writeJSON` abgedeckt). Dadurch werden Evaluationen und Flag-Antworten grundsätzlich nicht zwischengespeichert. Die Funktion des Produkts wird nicht beeinträchtigt.

### 1.4 Fehlende Zugriffskontrolle auf personenbezogene Verarbeitung
**Befund:**  
Die REST-API besitzt **keine Authentifizierung oder Autorisierung**. Jeder, der den Port erreicht, kann:  
- Feature-Flags auslesen (einschließlich evtl. vertraulicher Descriptions),  
- Flags ändern/löschen,  
- den `user`-Parameter für beliebige Nutzerkennungen an `evaluate` übergeben und so Rollout-Entscheidungen abfragen.  
**Schweregrad:** kritisch  
**Bewertung:** Zwar speichert der Dienst selbst keine Nutzerkennung, aber er verarbeitet sie und gibt eine Auskunft zurück. Ohne Zugriffsschutz können Unbefugte diese Verarbeitung auslösen und personenbezogene Inferenzen über Nutzerkennungen anstellen. Dies ist ein klarer Verstoß gegen den Grundsatz der Vertraulichkeit und gegen Security by Default.  
**Remedy:** Einführung einer Authentifizierungs-/Autorisierungsschicht (z. B. statischer API-Key oder Bearer-Token) als Middleware, z. B. in `internal/middleware/auth.go`. Anforderung:  
- Alle `/flags`-Routen (außer `/healthz`) müssen einen gültigen API-Key verlangen.  
- Konfiguration über Umgebungsvariable, z. B. `FLAG_API_KEY`.  
- Fehlerantwort bei ungültigem Token: `401` mit JSON-Fehlerobjekt über `writeError`.  
- Der Evaluation-Endpoint funktioniert weiterhin: Client muss den API-Key im `Authorization`-Header mitsenden.  
- Dokumentation im README, wie der Key gesetzt wird.  
Diese Änderung bricht die Produktfunktion nicht, solange Tests und Aufrufer den Key setzen. Sie ist zwingend, bevor der Dienst öffentlich erreichbar betrieben wird.

---

## 2. EU Cyber Resilience Act (CRA)

### 2.1 Security by Design / Default — Transportverschlüsselung
**Befund:**  
`main.go` startet `http.Server` mit `ListenAndServe()` (HTTP im Klartext). Es gibt keine TLS-Konfiguration, keine Zertifikatsoptionen und keinen Hinweis auf TLS-Terminierung durch ein vorgelagertes Gateway.  
**Schweregrad:** hoch  
**Bewertung:** Ein Produkt mit digitalen Elementen, das über ein Netzwerk kommuniziert, muss standardmäßig sichere Kommunikation ermöglichen. Klartext-HTTP erlaubt Abhören und Manipulation von Feature-Flags und Evaluationsergebnissen.  
**Remedy:**  
- Entweder `ListenAndServeTLS` mit Zertifikats- und Schlüsselpfaden aus Umgebungsvariablen implementieren,  
- oder in `README.md` verbindlich dokumentieren, dass der Dienst **nur hinter einem TLS-terminierenden Reverse Proxy / API-Gateway** betrieben werden darf, mit Beispiel-Konfiguration.  
Die produktinterne Funktion bleibt erhalten, wenn TLS sauber konfiguriert wird. Keine PII im Klartext transportieren.

### 2.2 Security by Design / Default — Authentifizierung und Autorisierung
**Befund:** Siehe DSGVO-Befund 1.4.  
**Schweregrad:** kritisch  
**Remedy:** Wie unter 1.4 beschrieben.

### 2.3 Schutz gegen Überlastung / Rate-Limiting
**Befund:**  
Es gibt keine Begrenzung der Anfragefrequenz pro Client. Die API kann von einem einzelnen Client mit vielen Anfragen überlastet werden (DoS).  
**Schweregrad:** mittel  
**Bewertung:** Für ein Produkt mit digitalen Elementen ist ein Basis-Schutz vor Ressourcenerschöpfung Teil sicherer Voreinstellungen.  
**Remedy:**  
- Optionaler, einfacher Token-Bucket oder Sliding-Window-Limiter in `internal/middleware/ratelimit.go`.  
- Alternativ im README dokumentieren, dass ein vorgelagertes API-Gateway Rate-Limiting übernehmen muss.  
- Konfiguration über Umgebungsvariablen, z. B. `RATE_LIMIT_RPS`.  
Die Erfüllung der fachlichen Anforderungen wird dadurch nicht beeinträchtigt, solange das Limit großzügig ist und Tests dies berücksichtigen.

### 2.4 Abhängigkeiten / SBOM
**Befund:**  
Das Projekt nutzt ausschließlich die Go-Standardbibliothek. Die `go.mod` enthält keine externen Abhängigkeiten.  
**Schweregrad:** niedrig  
**Bewertung:** Exzellent — minimale Angriffsfläche. Eine formale SBOM ist im Quellcode nicht sichtbar, aber bei nur Standardbibliothek vernachlässigbar; für ein echtes Release sollte ein maschinenlesbares SBOM erzeugt werden.  
**Remedy:**  
- Im README eine kurze Dependencies/SBOM-Sektion ergänzen: „Das Produkt verwendet ausschließlich die Go-Standardbibliothek; keine Drittanbieter-Abhängigkeiten.“  
- Optional: SBOM-Generierung in CI aufnehmen (z. B. `go version -m` oder CycloneDX-Tooling), sobald externe Module hinzukommen.

### 2.5 Update- und Patch-Fähigkeit
**Befund:**  
Es gibt keine im Code sichtbaren Mechanismen für Updates oder Patch-Verteilung. Der Server besitzt keine Versionsnummer im HTTP-Header oder Health-Endpoint.  
**Schweregrad:** niedrig  
**Bewertung:** Für ein kompiliertes Go-Binary ist die Update-Fähigkeit primär ein Betreiberthema. Eine Versionsangabe würde Transparenz schaffen und das Patch-Management erleichtern.  
**Remedy:**  
- In `main.go` eine Konstante `const ServiceVersion = "1.0.0"` definieren und im `handleHealth`-Antwort (`/healthz`) als `version`-Feld mitgeben.  
- Im README dokumentieren, wie Updates eingespielt werden (Build, Deployment).  
- Optional: eigene Update-Routine ist bei diesem Backend nicht nötig; Betreiber-Skript genügt.

---

## 3. EU AI Act
**Befund:**  
Keine KI-Funktion erkennbar. Der Dienst trifft deterministische Rollout-Entscheidungen auf Basis einer Hash-Funktion.  
**Schweregrad:** entfällt  
**Bewertung:** Kein Anwendungsbereich.

---

## 4. Mandatory texts & UI
**Befund:**  
Kein Endnutzer-UI, keine Website, kein Cookie-Banner, keine Verkaufsfunktion.  
**Schweregrad:** entfällt  
**Bewertung:** Für ein reines Backend ohne Nutzeroberfläche nicht einschlägig.

---

## 5. Accessibility
**Befund:**  
Keine öffentliche Web-Benutzeroberfläche.  
**Schweregrad:** entfällt  
**Bewertung:** WCAG/BITV/EAA nicht anwendbar.

---

## Zusammenfassung der offenen Punkte

| Nr. | Thema | Schweregrad | Bereich | Datei / Maßnahme |
|-----|-------|-------------|----------|------------------|
| 1   | Fehlende Authentifizierung/Autorisierung | kritisch | DSGVO, CRA | `internal/middleware/auth.go` neu; Konfiguration `FLAG_API_KEY`; alle `/flags`-Routen schützen, `/healthz` offen lassen |
| 2   | Kein TLS im Server | hoch | CRA | `main.go` `ListenAndServeTLS` oder verbindliche Gateway-Doku im `README.md` |
| 3   | Kein `Cache-Control: no-store` | mittel | DSGVO | `internal/handlers/handlers.go` in `writeJSON` ergänzen |
| 4   | Fehlende Datenschutz-/Betriebsdokumentation | mittel | DSGVO | `README.md` Abschnitt „Datenschutz / Verarbeitete Daten“ |
| 5   | Kein Rate-Limiting | mittel | CRA | optionale Middleware `internal/middleware/ratelimit.go` oder Gateway-Doku |
| 6   | Keine Versionsangabe im Health-Endpoint | niedrig | CRA | `main.go` Konstante; in `handleHealth` als `version` zurückgeben |
| 7   | Content-Type-Prüfung zu strikt | niedrig | Interoperabilität | `internal/handlers/create.go` `mime.ParseMediaType` verwenden |

Alle genannten Maßnahmen sind umsetzbar, ohne die Kernfunktion des Dienstes zu brechen. Nach Umsetzung der kritischen und hohen Punkte ist das Produkt für den internen, abgesicherten Betrieb grundsätzlich marktreif.