VERDICT: CHANGES_REQUESTED

## Prüfbericht

### 1. DSGVO

**F1 — Schweregrad: niedrig**  
Sachverhalt: Der `user`-Parameter (`GET /flags/{key}/evaluate?user=...`) ist potenziell ein personenbezogenes oder pseudonymes Merkmal. Im Code wird er ausschließlich für die Hash-Berechnung verwendet, nicht gespeichert, nicht persistent verarbeitet und nicht geloggt.  
Bewertung: Datenminimierung und Vermeidung von Log-Leaks sind gut umgesetzt; es besteht kein akuter Verstoß.  
Remedie: Keine Codeänderung erforderlich. In der Betriebsdokumentation (`README.md` oder `COMPLIANCE.md`) sollte festgehalten werden, dass der `user`-Wert nur flüchtig im Arbeitsspeicher verarbeitet und nicht gespeichert wird.

**F2 — Schweregrad: mittel**  
Sachverhalt: Für den Betrieb durch Dritte fehlt im Repository eine klare datenschutzrechtliche Einordnung des Dienstes (Verantwortlicher/Auftragsverarbeiter, Rechtsgrundlage, Auftragsverarbeitungsvertrag).  
Bewertung: Bei einem reinen Backend ohne Endnutzer-UI ist eine Datenschutzerklärung im Code nicht zwingend; die Verantwortlichkeit trifft den Betreiber.  
Remedie: In `COMPLIANCE.md` oder `README.md` einen Abschnitt ergänzen, der die Rollen und Pflichten für Betreiber beschreibt und auf die Notwendigkeit eines AV-Vertrags hinweist, falls die API durch andere Unternehmen genutzt wird.

### 2. EU Cyber Resilience Act (CRA) / Security by Design

**F3 — Schweregrad: kritisch**  
Sachverhalt: In `main.go` wird bei fehlender Umgebungsvariable `FLAG_API_KEY` ein zufälliger API-Schlüssel generiert und **im Klartext in das Log geschrieben**:  
`log.Printf("WARNING: FLAG_API_KEY not set; generated ephemeral key %s", apiKey)`  
Bewertung: Ein Logeintrag mit Zugangsdaten ermöglicht unbefugten Zugriff auf alle administrativen API-Endpunkte. Das verletzt Security by design/default und den Schutz vor unbefugtem Zugriff nach CRA.  
Remedie: In `main.go` die Zeile ändern zu z. B.:  
`log.Printf("WARNING: FLAG_API_KEY not set; generated ephemeral API key (not logged)")`  
oder den Key gar nicht generieren, sondern den Dienst mit klarer Fehlermeldung beenden, wenn `FLAG_API_KEY` fehlt.

**F4 — Schweregrad: hoch**  
Sachverhalt: Der Einsatz eines ephemeren API-Keys bei jedem Start (wenn `FLAG_API_KEY` nicht gesetzt ist) führt zu nicht reproduzierbaren Zugangsdaten. Clients können den Key nur aus dem Log erfahren, was unsicher und betrieblich nicht tragfähig ist.  
Bewertung: Sicherheitsrelevante Konfiguration muss deterministisch und geheim bleiben. Der Fallback ist ein Sicherheitsrisiko.  
Remedie: In `main.go` eine Pflichtkonfiguration einführen:  
`apiKey := os.Getenv("FLAG_API_KEY")`  
`if apiKey == "" { log.Fatal("FLAG_API_KEY must be set") }`  
Für lokale Entwicklung kann in der Doku ein Platzhalter wie `FLAG_API_KEY=dev-only-key` vorgesehen werden; das Produkt bleibt damit lauffähig.

**F5 — Schweregrad: mittel**  
Sachverhalt: Im Repository ist keine SBOM-Datei (Software Bill of Materials) sichtbar; `go.mod` ist vorhanden, aber eine maschinenlesbare SBOM fehlt.  
Bewertung: Für ein Produkt mit digitalen Elementen verlangt der CRA die Bereitstellung einer SBOM bzw. dokumentierter Abhängigkeiten. Da das Projekt nur die Standardbibliothek nutzt, ist das Risiko gering, aber die Dokumentationspflicht bleibt.  
Remedie: In der CI-Pipeline oder manuell ein SBOM-Tool einbinden, z. B. `cyclonedx-gomod` oder `syft`, und die erzeugte Datei `sbom.json` oder `sbom.spdx` im Repository ablegen. Alternativ eine maschinenlesbare Aufstellung der (leeren) externen Abhängigkeiten in `SECURITY.md` oder `COMPLIANCE.md`.

**F6 — Schweregrad: niedrig**  
Sachverhalt: Der Auth-Endpunkt (`Authorization: Bearer`) besitzt keinen Brute-Force-Schutz oder Rate-Limiting.  
Bewertung: `subtle.ConstantTimeCompare` verhindert Timing-Angriffe, aber wiederholte Fehlversuche werden nicht begrenzt.  
Remedie: In `internal/middleware/auth.go` oder einer zusätzlichen Middleware einen einfachen Token-Bucket/limiter pro IP oder Token ergänzen. Falls dies für die Zielumgebung nicht nötig ist, mindestens als Betriebshinweis in `SECURITY.md` dokumentieren.

### 3. EU AI Act

Nicht anwendbar: Die Anwendung enthält keine KI-Funktion im Sinne des AI Act.

### 4. Pflichttexte & UI

Nicht anwendbar: Reines Backend ohne öffentliche Web-Oberfläche; keine Impressum-, Cookie- oder Zustimmungspflichten im Produkt selbst.

### 5. Barrierefreiheit

Nicht anwendbar: Keine öffentliche Web-UI vorhanden.

---

**Hinweis zur Widerspruchsfreiheit:** Die vorgeschlagenen Maßnahmen (z. B. Pflichtkonfiguration `FLAG_API_KEY`, SBOM-Erzeugung, Rate-Limiting) verändern nicht die legitimen Datenflüsse oder Funktionen der API. Der Dienst bleibt unter den eigenen Sicherheitsanforderungen voll nutzbar; es sind lediglich Konfigurations- und Dokumentationsanpassungen erforderlich.