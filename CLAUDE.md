# KSight Design Memory

## Core Concept
User-centered Kubernetes GUI focused on real-world operations, not resource types.

## Layout Structure
- **Top**: Chrome-style tabs for cluster connections + settings gear
- **Left**: VSCode-style icon sidebar (Applications, Operations, Nodes, Resources, Templates, Boards)  
- **Main**: DataTable with filters, groupBy, views system
- **Right**: AI chatbox with k.operation() calls
- **Footer**: Ad-hoc command tabs (Add resource, K8S scripts, Pod shell/logs/files, Resource diff)

## Key Features
### Applications Plugin
- Pods grouped by app label
- Filters: app names, namespaces, resource types, labels, annotations, image, node
- GroupBy: app labels, env labels, workload type, node
- Actions: detail, edit yaml/status, delete, exec, logs, file upload/download

### Nodes Plugin  
- Filters: name/IP, labels, resource/device types
- GroupBy: region/zone, machine type, architecture, tenant
- Special: resource usage thumbnails, schedule simulator
- Cost tracking per node/group

### Resources Plugin
- Dynamic resource tree + table
- Configurable filters/groupBy/actions/columns per resource type

### Templates Plugin
- File tree structure with versioned git storage
- Template params with `~{}` syntax
- Share/save/duplicate functionality

### Operations Plugin
- TypeScript operations running in frontend
- Built-in troubleshooting tools (bpftrace, strace, pprof, arthas, etc.)
- Run history saved locally

### Boards Plugin
- Fully customizable dashboard panels
- Resource lists, metrics, terminals, topology trees, web URLs

## Tech Stack
- **Backend**: Golang dynamic informers + Wails framework
- **Frontend**: Shadcn-Vue + Pinia + TailwindCSS
- **SDK**: window.k with strong typing (k.pods.list().first().exec())
- **Cache**: Local storage with resourceVersion for consistency
- **Extensions**: Dynamic component loading, VSCode-style extension system

## Key Behaviors
- Views system: saved filters/groupBy/orderBy per menu
- Topology dialog: kubectl tree-like structure for any resource
- YAML editor: 3-part view (metadata/spec/status) with quick inputs
- Keyboard shortcuts: Cmd+T palette, S for shell, L for logs, etc.
- Resource diff: Monaco diff viewer with one-click merge

## Performance
- Virtual scroll for large lists
- Pagination for groups
- Local cache with file storage
- Avoid full sync on restart