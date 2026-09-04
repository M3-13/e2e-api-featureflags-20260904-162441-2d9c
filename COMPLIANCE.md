# COMPLIANCE.md

Betriebsdokumentation für den Feature-Flag-Dienst. Dieses Dokument beschreibt,
wie der Dienst mit Daten umgeht, welche datenschutzrechtlichen Rollen und
Rechtsgrundlagen für den Betrieb gelten und wie die Software-Abhängigkeiten
(Software Bill of Materials, SBOM) aufgestellt werden.

---

## 1. Verarbeitete Daten & Datenminimierung

Der Dienst verarbeitet ausschließlich die Daten, die über die HTTP-API
übermittelt werden. Es findet keine Persistenz statt: alle Daten liegen nur im
Arbeitsspeicher des laufenden Prozesses und gehen bei dessen Beendigung
verloren.

- **`user`-Parameter (`GET /flags/{key}/evaluate?user=...`):** Der Wert wird
  ausschließlich **transient im Arbeitsspeicher** für die deterministische
  Hash-Berechnung der Rollout-Entscheidung verwendet (FNV-64a über
  `key + "\x00" + user`). Er wird **weder gespeichert noch geloggt** und auch
  nicht in der Antwort wiederholt. Es erfolgt keinerlei Persistenz und keine
  Weitergabe an Dritte.
- **`key` und `description`:** Die Flag-Bezeichnung und Beschreibung sind
  anwendungsbezogene Metadaten. Es dürfen darin **keine personenbezogenen
  Daten** abgelegt werden. Die `description` ist optional und dient
  ausschließlich der internen Organisation der Flags.
- **Logging:** Das Zugriffs-Logging (Middleware) protokolliert Methode,
  URL-Pfad (ohne Query-String), Status und Dauer. Der `user`-Wert aus dem
  Query-String wird dabei bewusst nicht erfasst.

Datenminimierung: Es werden nur diejenigen Daten verarbeitet, die für die
Erbringung der jeweiligen API-Funktion zwingend erforderlich sind. Es gibt
keine Langzeitspeicherung, keine Protokollierung personenbezogener oder
pseudonymer Merkmale und keine Profilbildung.

---

## 2. Rollen & Rechtsgrundlage

- **Verantwortlicher / Auftragsverarbeiter:** Der Betreiber des Dienstes ist
  datenschutzrechtlich Verantwortlicher im Sinne des Art. 4 Nr. 7 DSGVO. Wird
  der Dienst als technische Komponente im Auftrag eines anderen Unternehmens
  betrieben, handelt der Betreiber als Auftragsverarbeiter im Sinne des
  Art. 4 Nr. 8 DSGVO.
- **Rechtsgrundlage:** Die Verarbeitung stützt sich auf das **berechtigte
  Interesse nach Art. 6 Abs. 1 lit. f DSGVO** (technischer und
  sicherheitstechnischer Betrieb des Feature-Flag-Dienstes). Soweit der Dienst
  im Auftrag betrieben wird, erfolgt die Verarbeitung im Rahmen einer
  **Auftragsverarbeitung nach Art. 28 DSGVO**.
- **Auftragsverarbeitungsvertrag (AVV):** Wird die API von anderen Unternehmen
  genutzt oder im Auftrag betrieben, ist zwischen dem Verantwortlichen und dem
  Auftragsverarbeiter ein **Auftragsverarbeitungsvertrag (AVV)** nach
  Art. 28 DSGVO erforderlich. Der Betreiber stellt sicher, dass ein solcher
  Vertrag vor Aufnahme der Verarbeitung geschlossen wird.

---

## 3. Dependencies / SBOM

- **Abhängigkeiten:** Das Produkt verwendet **ausschließlich die
  Go-Standardbibliothek** (`net/http` und zugehörige Pakete). In `go.mod` sind
  **keine externen Module** als Abhängigkeiten deklariert.
- **SBOM-Erzeugung:** Für Releases kann eine maschinenlesbare Software Bill of
  Materials (SBOM) erzeugt und im Repository abgelegt werden, z. B.:
  - `cyclonedx-gomod` → Ausgabe als `sbom.json` (CycloneDX) oder
    `sbom.spdx` (SPDX)
  - `syft` → Ausgabe als `sbom.json` (CycloneDX) oder `sbom.spdx` (SPDX)

  Beispiel:

  ```sh
  # CycloneDX (JSON)
  cyclonedx-gomod app -json -output sbom.json

  # Syft (SPDX)
  syft dir:. -o spdx-json=sbom.spdx.json
  ```

  Da keine externen Abhängigkeiten bestehen, enthält die SBOM keine
  Drittanbieter-Komponenten; sie dokumentiert die eingesetzte
  Go-Laufzeitumgebung und die Produktversion.
