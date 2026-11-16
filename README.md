# AREA

AREA: an IFTTT/Zapier-style automation platform (Go server, React web client, React Native Android client).

## Milestone 1 deliverables

- Technology selection: [back-end](docs/tech-selection/backend-selection.md), [front-end](docs/tech-selection/frontend-selection.md), [persistence](docs/tech-selection/persistence-selection.md)
- Spikes: [`spikes/`](spikes/) and the comparison in [`spikes/DISTANCE.md`](spikes/DISTANCE.md)
- [Recommendation](decision/RECOMMENDATION.md)
- [Features and additional requirements](docs/FEATURES.md)
- [Technical questions](docs/TECHNICAL_QUESTIONS.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Project plan and WBS](planning/project-wbs.md) (management method in section 2) and the [board](planning/board-project.md)
- Architecture decision records: [`docs/adr/`](docs/adr/)
- Exported versions (Word): [architecture](planning/AREA_System_Architecture_and_Options_Evaluation.docx), [features](planning/AREA_Feature_Prioritisation_and_Catalogue_Selection.docx), [plan and WBS](planning/AREA_Milestone_1_Project_Plan_and_WBS.docx)

## Repository layout

| Folder | What | Lint / type-check |
|---|---|---|
| [`server/`](server/) | Go application server, one binary with three roles (`ROLE=api\|scheduler\|worker`), entry point `cmd/area` | golangci-lint (config in `server/.golangci.yml`), gofmt + goimports |
| [`web/`](web/) | React + Vite + TypeScript web client | ESLint, `tsc`, Prettier |
| [`mobile/`](mobile/) | React Native + TypeScript Android client | ESLint, `tsc`, Prettier |

Prettier settings are shared from the root [`.prettierrc.json`](.prettierrc.json).

## Development

Requirements: Go 1.25+, [golangci-lint](https://golangci-lint.run/welcome/install/) v2, Node.js 22.11+.

```bash
make install   # download Go modules and npm packages
make lint      # every linter and type-check; must exit 0 before opening a PR
make format    # auto-fix formatting in all three modules
```

Each module can also be checked on its own: `make lint-server`, `make lint-web`, `make lint-mobile`.

## Installation Process

```bash
# Add installation steps here
```

## Usage

```bash
# Add usage examples here
```

## Testing

```bash
# Add the command to run the tests here
```

## License

Add the license here.