# Graph Report - golang-git-graph  (2026-07-09)

## Corpus Check
- 15 files · ~7,435 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 117 nodes · 182 edges · 12 communities (9 shown, 3 thin omitted)
- Extraction: 95% EXTRACTED · 5% INFERRED · 0% AMBIGUOUS · INFERRED: 9 edges (avg confidence: 0.8)
- Token cost: 0 input · 0 output

## Graph Freshness
- Built from commit: `125168ef`
- Run `git rev-parse HEAD` and compare to check if the graph is stale.
- Run `graphify update .` after code changes (no API cost).

## Community Hubs (Navigation)
- [[_COMMUNITY_AppModel|AppModel]]
- [[_COMMUNITY_FileChange|FileChange]]
- [[_COMMUNITY_Git-360 Golang Git Graph & Real-Time TUI Dashboard|Git-360: Golang Git Graph & Real-Time TUI Dashboard]]
- [[_COMMUNITY_NewAppModel|NewAppModel]]
- [[_COMMUNITY_GraphPane|GraphPane]]
- [[_COMMUNITY_Client|Client]]
- [[_COMMUNITY_DiffPane|DiffPane]]
- [[_COMMUNITY_ParseStatus|ParseStatus]]
- [[_COMMUNITY_graphify|graphify.md]]
- [[_COMMUNITY_graphify|graphify.md]]
- [[_COMMUNITY_golang-git-graph|golang-git-graph]]
- [[_COMMUNITY_MetaPane|MetaPane]]

## God Nodes (most connected - your core abstractions)
1. `AppModel` - 19 edges
2. `Client` - 13 edges
3. `FileChange` - 12 edges
4. `StatusPane` - 11 edges
5. `Styles` - 10 edges
6. `NewAppModel()` - 9 edges
7. `GraphPane` - 9 edges
8. `Commit` - 8 edges
9. `DiffPane` - 8 edges
10. `Git-360: Golang Git Graph & Real-Time TUI Dashboard` - 7 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `NewClient()`  [INFERRED]
  cmd/git360/main.go → internal/git/client.go
- `main()` --calls--> `NewAppModel()`  [INFERRED]
  cmd/git360/main.go → internal/tui/app.go
- `NewAppModel()` --calls--> `NewDiffPane()`  [INFERRED]
  internal/tui/app.go → internal/tui/pane_diff.go
- `NewAppModel()` --calls--> `NewMetaPane()`  [INFERRED]
  internal/tui/app.go → internal/tui/pane_meta.go
- `NewAppModel()` --calls--> `NewStatusPane()`  [INFERRED]
  internal/tui/app.go → internal/tui/pane_status.go

## Import Cycles
- None detected.

## Communities (12 total, 3 thin omitted)

### Community 0 - "AppModel"
Cohesion: 0.17
Nodes (12): Cmd, Model, Msg, Time, paneName(), AppModel, diffLoadedMsg, errorMsg (+4 more)

### Community 1 - "FileChange"
Cohesion: 0.33
Nodes (3): FileChange, NewStatusPane(), StatusPane

### Community 2 - "Git-360: Golang Git Graph & Real-Time TUI Dashboard"
Cohesion: 0.14
Nodes (13): Git-360: Golang Git Graph & Real-Time TUI Dashboard, 🚀 Implementation Phases, 🌟 Key Features, ⌨️ Keybindings Reference, Phase 1: Go Git Engine & Client Setup, Phase 2: Bubble Tea Boilerplate & Split Layout, Phase 3: Interactive Git Graph, Phase 4: Status Pane & File Staging (+5 more)

### Community 3 - "NewAppModel"
Cohesion: 0.20
Nodes (10): main(), NewAppModel(), colorizeGraph(), colorizeRefs(), Style, mapGraphRune(), NewGraphPane(), DefaultStyles() (+2 more)

### Community 4 - "GraphPane"
Cohesion: 0.31
Nodes (4): Commit, RepoState, Time, GraphPane

### Community 5 - "Client"
Cohesion: 0.22
Nodes (5): Client, NewClient(), ParseCommits(), parserScan(), ParseStatus()

### Community 6 - "DiffPane"
Cohesion: 0.22
Nodes (5): Cmd, Model, Msg, NewDiffPane(), DiffPane

### Community 7 - "ParseStatus"
Cohesion: 0.18
Nodes (10): 🏗️ Architecture Overview, Building from Source, 🚀 Future Roadmap & Next Steps, Git-360: Real-Time Git Observer & TUI Dashboard, Key Technical Decisions, ⌨️ Keybindings, 📐 Layout & Visual Design, Prerequisites (+2 more)

### Community 11 - "MetaPane"
Cohesion: 0.40
Nodes (3): Time, NewMetaPane(), MetaPane

## Knowledge Gaps
- **23 isolated node(s):** `golang-git-graph`, `errorMsg`, `diffLoadedMsg`, `graphify`, `Workflow: graphify` (+18 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **3 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `AppModel` connect `AppModel` to `FileChange`, `NewAppModel`, `GraphPane`, `Client`, `DiffPane`, `MetaPane`?**
  _High betweenness centrality (0.331) - this node is a cross-community bridge._
- **Why does `Client` connect `Client` to `AppModel`, `NewAppModel`?**
  _High betweenness centrality (0.107) - this node is a cross-community bridge._
- **Why does `DiffPane` connect `DiffPane` to `AppModel`?**
  _High betweenness centrality (0.089) - this node is a cross-community bridge._
- **What connects `golang-git-graph`, `errorMsg`, `diffLoadedMsg` to the rest of the system?**
  _23 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Git-360: Golang Git Graph & Real-Time TUI Dashboard` be split into smaller, more focused modules?**
  _Cohesion score 0.14285714285714285 - nodes in this community are weakly interconnected._