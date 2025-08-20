# KSight

KSight is a user-centered Kubernetes GUI tool, focus on real-world operations rather than Kubernetes resource types.

# Overall Layout

Top: Tabs of Active/Pinned Cluster Connections, Add button to trigger connect cluster from existing KubeConfig or Load From Plaintext KubeConfig / New KubeConfig File (note all saved KubeConfig files should be watched, and ~/.kube Folder is default watched), Think Chrome Tabs, the right side is settings gear to open settings page, design it like VSCode, also extensible for different plugins, each has one namespace

Left Sidebar(icon, narrow, like VSCode icon button at left)
- Vertical Icon Menus: Applications, Operations, Nodes, Resources, Templates, Workflows and more extensions (default to open workloads, remember last active menu and menu order in local storage, order could be drag-n-drop adjustable)
- Main Area, varies for each menu
- Right: AI Chatbox and chat histories, it can call K SDK and workflows like k.workflow('a').run(dynamic params)

On Most Pages' Main Area:
- Heading Line1: All opened views, if no saved active view, show default view with last filter/group/orders, each View could be closed, and plus button after all opened views, when click it, show dropdown of all saved views with filter, open from Saved Views or create new default view. (Views are are stored global scope, each View includes Filter and GroupBy and OrderBy, has some metadata, include which menu it belongs to, isPinned, isShared[gen shortlink for shared filter to send to any body] when switch between filters, criteria are replaced with current page);
- Body: DataTable:
- Table Heading: 
  - left: Filters
  - right: GroupBy Buttons, View/Hide Columns, Expand All/Collapse button
- Table Body: item list with pagination, could be Groups of item list when it has 2 levels, each item has a checkbox to select, and a checkbox to select all items in current page or all pages, and right side of each item is icon and grouped action buttons, icon actions need hover to show, each item clicked, open a width-adjustable drawer to show item details, (on top right, checkbox to hide managedFields, hidden by default)
Note: add pagination for groups, and inside each group, use virtual scroll list when count too large in one sub-group, each group is collapsable

Footer: Ad-hoc Command Tabs, collapsable, user can click '+' button at end of Tab to Add resource | Run k. command scripts | Pod shell | Pod logs | Pod files | Resource Diff
- When it's Add resource, open the MonacoEditor, yaml , with Apply | Cancel buttons
- When it's k.command scripts, open the typescript MonacoEditor left, Xterm right, with Run | Cancel buttons, right side for showing ad-hoc output, if it's K8S resource, show formatted output like kubectl, resolve the printed columns defined in k8s API
- When it's Pod shell | Pod logs, open the Xterm, similar to VSCode terminal, with split view, add switch container, command shortcut features on top of this special Xterm.
- When it's Pod files, left side is Dir/File tree of this Pod, with top input to switch base dir or exact file or glob file pattern, when clicked, right side is monaco editor to show file content for less than 10MB(adjustable setting) file. user can modify it, backed by Upload/Download using tar like kubectl cp, but impl with golang and wrapped in 'K' ts sdk, and when no file is selected, show two buttons upload/download here, to upload/download files to this pod and current active base dir
- When it's Resource Diff, select left and right, each could be existing resource, or template, or ad-hoc pasted, then open Monaco diff viewer, and when one side is existing resource, allow user to one click to merge or replace changes (ignore metadata name/ns, KGV) and Save to cluster

For Every Yaml Editor View, it has 3 parts, metadata, spec|data etc., status:
- If it's modify resource Editor: hide last-applied-configuration and managedFields, collapse status. if the spec is known type like Deployment, show some quick inputs above, like replicas, image, command, args etc. to avoid find inside the monaco editor
- If it's modify status Editor: hide spec|data, show status, and emphasize edit status can result in unexpected behavior

Built-in Topology Dialog:
In the detail view of any resource type, show root level controller object, and show additional topology button on the root controller object, build kubectl tree like structure, each resource can view its detail or edit yaml, if leaf nodes number > 10 items, show as N more items, click to expand 10 or expand all.

# Built in Plugins

##  Applications

Show Pods grouped by app label by default, workloads include k8s native workloads and user-configured custom resource types

Registered filters:
- App names: string includes ignore case
- Namespaces (multi select)
- Resource Types (multi select)
- Resource Labels (multi select)
- Resource Annotations (multi select)
- Image: string includes ignore case
- Node

Registered groupBy buttons:
- App labels (kubernetes.io/app, could be customized in settings, join together with defined delimiter for grouping): groupBy title side info: show pod image versions distribution with sub-groupBy, show pod readiness with sub-groupBy
- Env labels (kubernetes.io/env, could be customized in settings, join together with defined delimiter for grouping)
- Workload type
- Node

Default to AppLabels + WorkloadType groupBy, show sub-group count

Columns:
- Pod
- Workload [show root workload (settings to exclude some types)]
- Node
- Image
- Status
- Age
- Ready
- Containers 
- ...
- Action: Detail | Edit Yaml | Edit Status | Delete | Owned Workload Detail | Edit ConfigMaps | Edit Secrets | Exec | Logs | Upload/Download Files

Pod Detail:
- Heading: actions are the same, show pod metadata, show owner(controller-ref) workload tree in list with underline, click to show detail of each, show edit icon upon hover, click to edit yaml
- Body:
 - Computing (show container images/cmds/args, quotas, status, metrics chart of containers, also link non standard resources and devices claims here)
 - Configuration (show configmaps, secrets, one click to edit)
 - Networking (show related k8s service, one click to port-forward)
 - Storage (show related volumes)
- Footer: k8s events, last exit reason, last stop time etc.

workload view，show pod，linked resource and node view，group to owner，kind， log，exec，attach，events，metrics （versions and img versions，one click update version to owner workload） reverse find service and one click forward，ephemeral debug container（showhide daemonsets）

## Operations

Heading: Opened Operation View, Add/Duplicate view.

Fully customizable dashboard, user can build any View with Add Panel, Add Group, each Panel can be a Resource List View, Resource Detail View, Node/Pod Metrics View, or a Web URL, or an Ad-hoc Terminal, or a Topology Tree View, or a Pod Shell with Pre-defined Commands(prompt user to click run or run with params or not run when open this view), or a Pod Logs, or a Github/Ksight Workflow Trigger, or a Resource Status View, or a Events View.

## Nodes

Registered Filters:
- Node name | IP
- Node labels
- Resource/Device Type (multi select)

Registered groupBy buttons:
- Region & Zone
- MachineType Combo (customize in settings, like instanceType)
- Architecture (customize in settings, default to cpu arch)
- Tenant (user configured custom labels, such as service-group/team)

Additional Group side info:
- Total cost per hour, per month of this group
- Avg allocation and usage of each type of resource

Registered special button:
- Toggle cpu|mem|storage|device: show/hide resource usage thumbnail
- Simulate Schedule: show schedule dialog, dynamic add array of pod yaml + count, show schedule result at bottom, pod yaml input can be full screen or add from Template (backend fork a separate process to handle this, it won't watch resource, only call SchedulePod scheduling cycle API, and show result in frontend), show/hide fake pods consumption in resource view, the scheduler config can be customized and switched in simulate schedule dialog

Columns:
- common node fields
- device resourceSlices count (click to expand ResourceSlices of this node)
- pod count(progress bar, calculate NodeLevel pod and DaemonSet pods separately)
- resource allocation thumbnail (each type one line, show progress bar, click to expand, see each Pod's usage)
- resource usage thumbnail
- provisioner (root controller ref)
- hostPorts assigned(click to expand detail)
- Cost: show cost of this node (pricing can be customized and override in settings, include overall discount and certain type override)
- Action: Detail, cordon/evict, Drain, Node Shell, Resource Usage, Events, Metrics, Pod List(one click to hide DaemonSet-like pods, group by namespace and order by name by default)

## Resources

Left Sidebar:
- top: pinned resource kind
- resource type tree, similar to other k8s dashboards

Right TableList:
- Filters: name / namespace / dynamic filters from settings (type-> field-> )
- GroupBy: dynamic groupBy actions from settings (type-> field)
- Action: dynamic actions from settings (type-> ActionName, ActionIcon, ActionFn)
- Columns: dynamic columns from print columns from CRD API or built-in map of known resources, Age ...

## Templates

Left Sidebar: File tree structure, manage any templates, pattern search input on top, versioned using local git. 3 parts: pinned templates, my templates, shared templates (from links)
Right: when click and template, show overview info on top, name, description etc., include template params(name, type, default val), actions like share|save|delete|duplicate, and monaco editor for main area, use `~{}` for template params

## Workflows

Layout similar to Templates, but each workflow click is to show overview and run history, when click edit, show monaco editor of typescript, and run button.
Actions: Run(with params, default to current active cluster), Run in multiple clusters, need double confirm

Saved Workflows (Belongs to Operation Menu, each Workflow is a typescript running on frontend);

Workflow ship with built-in troubleshooting ways:
- General Linux tools: bpftrace strace ，dump mem/stacktrace
- Language specific tools: go pprof, java arthas etc., c++ gdb
- Http/grpc/websocket tracing and debug workflow etc.
- K8S operation flows: batch delete res, remove owner ref

Run history and output are saved in local file system

## Security

(not built in for OpenSource, Paid feature impl in other repo)
K8S RBAC insight, like kubectl auth can-i，who-can
Filter outdated images with CVE
...

## Networking

(not built in for OpenSource, Paid feature impl in other repo)

debug network policy，net topology, network tracing flows and commands etc.


# Keyboard Shortcuts

(similar to K9S, Chrome)

- Cmd + T, Command Palette (Run any command here, especially connect/disconnect cluster, switch namespace filter, toggle resource type filter, save/load favorite filters)
- Cmd + W, Close Current Cluster Connection
- Cmd + D, Pin Current Cluster Connection
- Cmd + F, Whole page search, use WebView's API to register, don't write component code
- Cmd + R, Force reload all cache and watchers
- ↑ ↓, Move active resource item in list, Enter to show details
- S, Open shell (first container)
- L,  Open logs (all containers), and press P to switch to previous terminated container
- A, Toggle All namespaces
- D, Delete dialog (Force Deletion / Background / Remove Finalizer and Delete checkboxes in side the dialog)
- ......

# Tech Requirements

- Use Golang Dynamic Informers to watch needed resources, and use Wails Framework for EventsOn/EventsOff to sync to JS side
- Keep Shadcn-Vue / Pinia / TailwindCSS stack, when you need new common component, Ask me first to offer a draft, if i say continue, you define the component directly
- All TS functions wrapped in window.k as SDK, considering type inference and strong typing. all data from frontend memory maintained by events, and it can also be used in workflow, like await k.pods.list({ labelSelector: 'app=nginx' }).first().exec({ command: ['sh', '-c', 'echo hello'] }); await k.resource("apiVersionAndGroup", "SomeCR").get("namespace", "name").util(k.conditionReady or custom predicate); k.update(() => {k.deployment.get("namespace", "name");return updatedDeployment}, maxRetryOnConflictOption); k.patch(jsonPatch); k.createOrUpdate((obj) => { obj.isEmpty() return new Obj; else assigned fields return updatedObj}, maxRetryOnConflictOption); wrap errors, allow err.isNotFound(), err.isConflict(), err.isRateLimited() etc., and err.statusCode, err.message for raw error.
- Be careful of K8S cache mechanism, think utilize stored resourceVersion and local file storage to speed up first init and avoid loading too much data and make pressure on etcd, think best and concise way of modify informer lastVersion, don't trigger full sync when this process restarts, load from local files as much as possible, and keep cache strong consistency
- Use log file to store all sdk mutations and get/list/watch request metadata for verbose log and debug, each log has unique id(workflow execution ID, or default UUID when each time this process starts), and store all workflow execution history, can link to all SDK run log, and link to workflow info
- Each page can mock and run separately, and use dynamic component to register and mount it, this is key to allow Extension mechanism, any one can use shadcn-vue and wrapped "K" sdk to develop KSight extension, make sure the whole project is extensible, also the command+T palette should be extensible and publish to npm registry, think how VSCode/Raycast extension works. For example Security could be a extension, detect all outdated images and dangerous k8s permissions, configs etc. All data layer functions are run in frontend using 'with + Proxy' sandbox, UI components are limited to main project existing components and auto imported to dynamic injected components
- Pages be condense, like IDE style, don't write too many repeated tailwind classes, think abstraction and componentization

# Other Requirements

- Don't write too many codes at once, do one small thing at a time, ask me to review and continue
- Be concise and clear, don't reinvent wheels, no redundant codes and comments, think logic abstraction, don't repeat, like pro.