# Copilot Configuration Files Guide

This directory contains comprehensive configuration files for GitHub Copilot and Cursor AI to provide optimal code assistance for the BookScraping project.

## Files Overview

### 1. `.github/copilot-instructions.md` (298 lines, 8.2KB)
**Purpose**: Primary Copilot instructions file with comprehensive project information.

**Contains**:
- Project overview and key features
- Complete tech stack details
- Architecture and component descriptions
- Detailed coding conventions for Go and TypeScript/Svelte
- Build and development commands
- Common tasks and workflows
- Security considerations
- License information

**Best for**: General code assistance, understanding project structure, learning conventions.

### 2. `.github/copilot-project-context.md` (412 lines, 11KB)
**Purpose**: Deep-dive technical documentation for complex development tasks.

**Contains**:
- Detailed architecture breakdown
- System component interactions
- Data flow patterns (Booklore sync, Goodreads scraping, SSE)
- External API integrations
- Build and deployment process
- Testing strategies
- Performance considerations
- Security model
- Common gotchas and extension points
- Useful SQL queries

**Best for**: Complex features, architecture decisions, debugging, performance optimization.

### 3. `.github/copilot-workspace-rules.yml` (181 lines, 5.6KB)
**Purpose**: Structured YAML configuration for workspace-level rules.

**Contains**:
- Code style rules for Go and TypeScript/Svelte
- Naming conventions by language
- Architecture patterns
- File organization structure
- Testing guidelines
- Build commands reference
- Common pitfalls
- Environment variables
- Security practices
- Version control guidelines

**Best for**: Quick reference, configuration-based tools, structured rule enforcement.

### 4. `.github/copilot-quick-reference.md` (356 lines, 7.9KB)
**Purpose**: Instant access to common patterns, snippets, and commands.

**Contains**:
- Quick start commands
- File location reference table
- Ready-to-use code snippets for all common patterns
- Command cheat sheet (dev, build, test, database)
- API endpoints table
- Database schema overview
- Troubleshooting guides
- Import path examples
- Git ignore patterns
- Security checklist

**Best for**: Day-to-day development, copy-paste snippets, command lookup, troubleshooting.

### 5. `.cursorrules` (217 lines, 6.8KB)
**Purpose**: Unified configuration file for Cursor editor and Copilot.

**Contains**:
- Project context summary
- Concise code style rules
- Key patterns with examples
- Architecture overview
- Command reference
- Best practices
- Common tasks
- Important notes and warnings
- Things to avoid

**Best for**: Cursor editor integration, consolidated reference, editor-specific assistance.

## Which File to Reference?

### For New Contributors
Start with: `.github/copilot-instructions.md`
- Learn project overview
- Understand tech stack
- Get familiar with conventions

### For Daily Development
Use: `.github/copilot-quick-reference.md`
- Copy code snippets
- Look up commands
- Find file locations
- Quick troubleshooting

### For Complex Features
Reference: `.github/copilot-project-context.md`
- Understand data flows
- Learn architecture patterns
- Plan major changes
- Debug complex issues

### For Editor Integration
Use: `.cursorrules`
- Cursor/Copilot inline assistance
- Real-time code suggestions
- Automated code generation

### For Tool Configuration
Use: `.github/copilot-workspace-rules.yml`
- Structured rule enforcement
- Automated linting
- Configuration-based tools

## How These Files Work Together

```
┌─────────────────────────────────────────────────────────┐
│         New to Project? Start Here                      │
│         .github/copilot-instructions.md                 │
│         (Overview, conventions, basics)                 │
└─────────────────┬───────────────────────────────────────┘
                  │
     ┌────────────┴────────────┐
     │                          │
     ▼                          ▼
┌─────────────────┐    ┌──────────────────┐
│ Daily Work?     │    │ Complex Feature? │
│ Quick Ref       │    │ Project Context  │
│ (Snippets)      │    │ (Architecture)   │
└─────────────────┘    └──────────────────┘
     │                          │
     └────────────┬─────────────┘
                  │
                  ▼
         ┌────────────────┐
         │ Editor Help?   │
         │ .cursorrules   │
         │ (Inline hints) │
         └────────────────┘
```

## Usage Tips

### For GitHub Copilot
GitHub Copilot automatically reads files in `.github/` directory, especially:
- `copilot-instructions.md` - Primary instructions
- `copilot-project-context.md` - Additional context
- `copilot-workspace-rules.yml` - Structured rules

### For Cursor Editor
Cursor reads `.cursorrules` file from the repository root for inline assistance and chat context.

### For Manual Reference
All files are markdown/YAML and human-readable. Use them as documentation:
```bash
# Search across all Copilot files
grep -r "sqlc" .github/copilot*.md .cursorrules

# View specific file
cat .github/copilot-quick-reference.md | less
```

### For Team Onboarding
1. Read `copilot-instructions.md` for project overview
2. Run commands from `copilot-quick-reference.md` to set up
3. Reference `copilot-project-context.md` when implementing features
4. Keep `copilot-quick-reference.md` open for daily reference

## Maintenance

### Updating These Files
When project changes occur, update relevant sections in:
- **New dependency**: Update tech stack in `copilot-instructions.md` and dependencies in YAML
- **New command**: Add to `copilot-quick-reference.md` and build section in `copilot-instructions.md`
- **Architecture change**: Update `copilot-project-context.md` and architecture section in all files
- **New pattern**: Add code snippet to `copilot-quick-reference.md` and pattern to `copilot-instructions.md`
- **Convention change**: Update all files to maintain consistency

### Consistency Checks
Before committing updates:
1. Ensure command examples are accurate (test them)
2. Verify code snippets compile/run
3. Check cross-references between files
4. Update line counts and sizes in this README
5. Maintain consistent formatting and style

## File Statistics Summary

| File | Lines | Size | Purpose |
|------|-------|------|---------|
| `copilot-instructions.md` | 298 | 8.2KB | Primary instructions & conventions |
| `copilot-project-context.md` | 412 | 11KB | Deep technical documentation |
| `copilot-workspace-rules.yml` | 181 | 5.6KB | Structured rules configuration |
| `copilot-quick-reference.md` | 356 | 7.9KB | Daily reference & snippets |
| `.cursorrules` | 217 | 6.8KB | Cursor editor configuration |
| **Total** | **1,464** | **39.5KB** | **Complete coverage** |

## Benefits

With these files, GitHub Copilot and Cursor will:
- ✅ Suggest code that follows project conventions
- ✅ Generate correct import statements
- ✅ Use proper error handling patterns
- ✅ Follow naming conventions
- ✅ Implement architecture patterns correctly
- ✅ Suggest appropriate logging and comments
- ✅ Use correct commands and tools
- ✅ Avoid common pitfalls
- ✅ Maintain consistency across codebase
- ✅ Provide relevant code snippets

## Questions?

If you're unsure which file to reference:
1. Check this README's "Which File to Reference?" section
2. Use `copilot-quick-reference.md` as a starting point
3. Search across all files for specific topics: `grep -r "search-term" .github/copilot* .cursorrules`

---

*Last updated: 2026-01-20*
*Total documentation: ~1,500 lines covering all aspects of the BookScraping project*
