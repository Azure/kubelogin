# Development Instructions

## Project Context

`kubelogin` is a golang commandline application implementing azure authentication to be used by kubectl via client-go credentials plugin. It provides features that are not available in kubectl such as using `azurecli`, `spn`, `workloadidentity` login, etc.

Most of the "login modes" are the implementation of [`azidentity`](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity) authentication flow.

`kubelogin` provides two main sub-commands `convert-kubeconfig` (pkg/cmd/convert.go) and `get-token` (pkg/cmd/token.go). The `get-token` sub-command gets the Entra ID access token based on the specified login mode (--login) in the kubeconfig. The `convert-kubeconfig` sub-command is a helper command to convert the kubeconfig to the specified login modes. The conversion may be lossy as some information may not be needed by the credential mode. The `convert-kubeconfig` command shares the exact command options as `get-token` command in `pkg/internal/token/options.go`. Besides using the command options, some options may be set via environment variable. The conversion code is in `pkg/internal/converter/convert.go`.

### Architecture Guidelines

For detailed guidance on extending CLI options and kubeconfig conversion functionality, see [CLI Options and Conversion Architecture](instructions/options-and-conversion.instructions.md). This document covers the unified options system, validation patterns, and best practices for adding new CLI arguments.

#### Execution Flow

**Token Execution:**
```
executeToken() → ValidateForTokenExecution() → token.NewAzIdentityCredential() → tokenPlugin.Do()
```

**Conversion:**
```
executeConvert() → buildExecConfig() → ValidateForConversion() → Save kubeconfig
```

#### Validation Architecture

- **`ValidateForTokenExecution()`**: Strict validation before authentication (all fields required)
- **`ValidateForConversion()`**: Lenient validation after field extraction (allows env vars)

## Development Flow

- Build: `make kubelogin`
- Test: `make test`

## Coding Standard

coding standard is defined in `.github/instructions/go.instructions.md`.

## Breadcrumb Protocol

A breadcrumb is a collaborative scratch pad that allow the user and agent to get alignment on context. When working on tasks in this repository, you **MUST** follow this collaborative documentation workflow to create a clear trail of decisions and implementations:

1. At the start of each new task, ask me for a breadcrumb file name if you can't determine a suitable one.

2. Create the breadcrumb file in the `.github/.copilot/breadcrumbs` folder using the format: `yyyy-mm-dd-HHMM-{title}.md` (*year-month-date-current_time_in-24hr_format-{title}.md* using UTC timezone)

3. Structure the breadcrumb file with these required sections:
   - **Requirements**: Clear list of what needs to be implemented.
   - **Additional comments from user**: Any additional input from the user during the conversation.
   - **Plan**: Strategy and technical plan before implementation.
   - **Decisions**: Why specific implementation choices were made.
   - **Implementation Details**: Code snippets with explanations for key files.
   - **Changes Made**: Summary of files modified and how they changed.
   - **Before/After Comparison**: Highlighting the improvements.
   - **References**: List of referred material like domain knowledge files, specification files, URLs and summary of what is was used for. If there is a version in the domain knowledge or in the specifications, record the version in the breadcrumb.

4. Workflow rules:
   - Update the breadcrumb **BEFORE** making any code changes.
   - **Get explicit approval** on the plan before implementation.
   - Update the breadcrumb **AFTER completing each significant change**.
   - Keep the breadcrumb as our single source of truth as it contains the most recent information.

5. Ask me to verify the plan with: "Are you happy with this implementation plan?" before proceeding with code changes.

6. Reference related breadcrumbs when a task builds on previous work.

7. Before concluding, ensure the breadcrumb file properly documents the entire process, including any course corrections or challenges encountered.

This practice creates a trail of decision points that document our thought process while building features in this solution, making pull request review for the current change easier to follow as well.

### Plan Structure Guidelines
- When creating a plan, organize it into numbered phases (e.g., "Phase 1: Setup Dependencies").
- Break down each phase into specific tasks with numeric identifiers (e.g., "Task 1.1: Add Dependencies").
- Include a detailed checklist at the end of the document that maps to all phases and tasks.
- Mark tasks as `- [ ]` for pending tasks and `- [x]` for completed tasks.
- Start all planning tasks as unchecked, and update them to checked as implementation proceeds.
- Each planning task should have clear success criteria.
- End the plan with success criteria that define when the implementation is complete.
- Plans should start with writing Unit Tests first when possible, so we can use those to guide our implementation. Same for UI tests when it makes sense.

### Following Plans
- When coding you need to follow the plan phases and check off the tasks as they are completed.  
- As you complete a task, update the plan and mark that task complete before you being the next task. 
- Tasks that involved tests should not be marked complete until the tests pass. 
