# Sicherheitshinweise (SECURITY)

Dieses Dokument beschreibt die sicherheitsrelevanten Betriebsvorgaben und
-Empfehlungen für den Feature-Flag-Service. Es ergänzt die Betriebshinweise in
`README.md`.

## Transportverschlüsselung

Die API darf ausschließlich hinter einem TLS-terminierenden Reverse Proxy oder
API-Gateway betrieben werden. Der Dienst stellt selbst kein TLS bereit und
überträgt den Bearer-Token im `Authorization`-Header im Klartext.

- Setzen Sie `ADDR` nicht ohne TLS auf eine öffentliche Schnittstelle (z. B.
  `ADDR=0.0.0.0:8080`). Öffentlich erreichbar ist der Dienst nur über ein
  vorgelagertes Gateway, das HTTPS abschließt.
- Wird die API ohne TLS-Terminierung auf eine nicht-lokale Schnittstelle
  gebunden, wird der Bearer-Token im Klartext übertragen und kann im Netzwerk
  mitgelesen werden.
- Für die lokale Entwicklung ist die Standardbindung `127.0.0.1:8080` unkritisch.

## Rate-Limiting & Brute-Force-Schutz

Die API-Key-Authentifizierung besitzt kein eingebautes Fehlversuchs- oder
Rate-Limit. Für den Betrieb ist ein vorgelagertes API-Gateway mit Rate-Limiting
(z. B. pro IP oder pro Token) zu empfehlen, um Brute-Force-Versuche und
Überlastung zu begrenzen.

- Der API-Key-Vergleich erfolgt zeitkonstant (`subtle.ConstantTimeCompare`) und
  verhindert Timing-Angriffe, begrenzt aber keine wiederholten Fehlversuche.
- Ein vorgelagertes Gateway sollte zusätzlich eine Begrenzung der Anfragen pro
  Zeitfenster (Rate-Limit) und der fehlgeschlagenen Authentifizierungsversuche
  durchsetzen.

## API-Key-Handling

`FLAG_API_KEY` ist verpflichtend und wird nie geloggt und nie im Repository
abgelegt.

- Der Server startet nicht ohne gesetzten `FLAG_API_KEY`; der Schlüssel wird
  ausschließlich über die Umgebungsvariable übergeben.
- Der Schlüssel wird nicht in Logs ausgegeben und darf nicht in Quellcode,
  Konfigurationsdateien oder Dokumentationen im Repository abgelegt werden.
- Für die lokale Entwicklung kann ein Platzhalter wie
  `FLAG_API_KEY=dev-only-key` genutzt werden; produktive Schlüssel sind geheim
  zu halten.
