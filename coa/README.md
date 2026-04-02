# ROADMAP: coa (Orchestrator) 🗺️

L'evoluzione di **coa** per raggiungere la parità funzionale con `penguins-eggs`.

## ✅ Phase 1: Il Nido (Completata)
- [x] Architettura base in Go (monorepo).
- [x] Integrazione con il motore C (`oa`).
- [x] Discovery delle distribuzioni (Madri e Derivate via YAML).
- [x] Generazione dinamica dei Flight Plans (JSON in memoria).

## 🚧 Phase 2: La Cova (In Corso - v0.5.x)
- [ ] **CLI Avanzata**: Implementazione dei sotto-comandi `produce`, `kill`, `status`.
- [ ] **Validation Layer**: Controllo preventivo dello spazio su disco e dei permessi di root.
- [ ] **Log Streaming**: Visualizzazione pulita dei log provenienti dal motore C.
- [ ] **Custom Excludes**: Gestione delle liste di esclusione dinamiche.

## 🥚 Phase 3: La Schiusa (v0.6.x - v0.8.x)
- [ ] **Wardrobe Integration**: Gestione dei "costumi" (configurazioni) via Git/YAML.
- [ ] **TUI (Terminal User Interface)**: Integrazione di `coa dad` e `coa mom` (configuratori visuali).
- [ ] **Export Tools**: Automazione per l'upload e il checksum delle ISO.

## 🐧 Phase 4: Volo Libero (v1.0.0)
- [ ] **Krill Rebirth**: Integrazione del nuovo installer di sistema.
- [ ] **Parità Funzionale**: Raggiungimento di tutte le feature core di `penguins-eggs`.
- [ ] **Documentazione**: Wiki e man pages generate automaticamente.