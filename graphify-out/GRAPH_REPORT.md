# Graph Report - golang-git-graph  (2026-07-09)

## Corpus Check
- 17 files · ~8,996 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 145 nodes · 232 edges · 14 communities (9 shown, 5 thin omitted)
- Extraction: 95% EXTRACTED · 5% INFERRED · 0% AMBIGUOUS · INFERRED: 12 edges (avg confidence: 0.8)
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
- [[_COMMUNITY_ParseStatus|ParseStatus]]
- [[_COMMUNITY_BranchBrowser|BranchBrowser]]

## God Nodes (most connected - your core abstractions)
1. `AppModel` - 26 edges
2. `Client` - 19 edges
3. `FileChange` - 12 edges
4. `StatusPane` - 11 edges
5. `Styles` - 11 edges
6. `NewAppModel()` - 10 edges
7. `GraphPane` - 9 edges
8. `Commit` - 8 edges
9. `BranchBrowser` - 8 edges
10. `DiffPane` - 8 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `NewClient()`  [INFERRED]
  cmd/git360/main.go → internal/git/client.go
- `main()` --calls--> `NewAppModel()`  [INFERRED]
  cmd/git360/main.go → internal/tui/app.go
- `NewAppModel()` --calls--> `NewBranchBrowser()`  [INFERRED]
  internal/tui/app.go → internal/tui/pane_branch.go
- `NewAppModel()` --calls--> `NewDiffPane()`  [INFERRED]
  internal/tui/app.go → internal/tui/pane_diff.go
- `NewAppModel()` --calls--> `NewStatusPane()`  [INFERRED]
  internal/tui/app.go → internal/tui/pane_status.go

## Import Cycles
- None detected.

## Communities (14 total, 5 thin omitted)

### Community 0 - "AppModel"
Cohesion: 0.27
Nodes (4): Cmd, Model, Msg, AppModel

### Community 1 - "FileChange"
Cohesion: 0.24
Nodes (5): FileChange, RepoState, Time, NewStatusPane(), StatusPane

### Community 2 - "Git-360: Golang Git Graph & Real-Time TUI Dashboard"
Cohesion: 0.14
Nodes (13): Git-360: Golang Git Graph & Real-Time TUI Dashboard, 🚀 Implementation Phases, 🌟 Key Features, ⌨️ Keybindings Reference, Phase 1: Go Git Engine & Client Setup, Phase 2: Bubble Tea Boilerplate & Split Layout, Phase 3: Interactive Git Graph, Phase 4: Status Pane & File Staging (+5 more)

### Community 3 - "NewAppModel"
Cohesion: 0.14
Nodes (13): main(), NewAppModel(), colorizeGraph(), colorizeRefs(), Style, mapGraphRune(), NewGraphPane(), Time (+5 more)

### Community 4 - "GraphPane"
Cohesion: 0.38
Nodes (3): Commit, gitStateMsg, GraphPane

### Community 6 - "DiffPane"
Cohesion: 0.22
Nodes (5): Cmd, Model, Msg, NewDiffPane(), DiffPane

### Community 7 - "ParseStatus"
Cohesion: 0.18
Nodes (10): 🏗️ Architecture Overview, Building from Source, 🚀 Future Roadmap & Next Steps, Git-360: Real-Time Git Observer & TUI Dashboard, Key Technical Decisions, ⌨️ Keybindings, 📐 Layout & Visual Design, Prerequisites (+2 more)

### Community 11 - "MetaPane"
Cohesion: 0.17
Nodes (11): Time, paneName(), branchesLoadedMsg, checkoutFinishedMsg, diffLoadedMsg, errorMsg, fetchFinishedMsg, Pane (+3 more)

### Community 12 - "ParseStatus"
Cohesion: 0.36
Nodes (6): ParseCommits(), parserScan(), ParseStatus(), TestParseCommits(), TestParseStatus(), T

## Knowledge Gaps
- **27 isolated node(s):** `golang-git-graph`, `errorMsg`, `diffLoadedMsg`, `branchesLoadedMsg`, `tagsLoadedMsg` (+22 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **5 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `AppModel` connect `AppModel` to `FileChange`, `NewAppModel`, `GraphPane`, `Client`, `DiffPane`, `MetaPane`, `BranchBrowser`?**
  _High betweenness centrality (0.398) - this node is a cross-community bridge._
- **Why does `Client` connect `Client` to `AppModel`, `NewAppModel`?**
  _High betweenness centrality (0.149) - this node is a cross-community bridge._
- **Why does `FileChange` connect `FileChange` to `AppModel`, `GraphPane`, `ParseStatus`, `Client`?**
  _High betweenness centrality (0.096) - this node is a cross-community bridge._
- **What connects `golang-git-graph`, `errorMsg`, `diffLoadedMsg` to the rest of the system?**
  _27 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Git-360: Golang Git Graph & Real-Time TUI Dashboard` be split into smaller, more focused modules?**
  _Cohesion score 0.14285714285714285 - nodes in this community are weakly interconnected._
- **Should `NewAppModel` be split into smaller, more focused modules?**
  _Cohesion score 0.1368421052631579 - nodes in this community are weakly interconnected._