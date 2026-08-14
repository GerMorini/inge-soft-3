# UI Contract: FitPro exercise routines

## Scope

This contract defines observable frontend behavior. It does not change
`contracts/openapi.yaml`, PostgreSQL, ownership rules or authenticated endpoint semantics.

## Shared visual language

- Use the existing `academic-dark` daisyUI theme and constitutional semantic colors.
- Use the optimized local copy of the supplied gym image for authentication and section heroes.
- Use named `lucide-react` imports. Icons inherit `currentColor` and never replace required text,
  accessible names or non-color state indicators.
- Keep component styles in daisyUI/Tailwind classes beside each component. Global CSS remains
  limited to theme and document-wide rules.

## Unauthenticated shell

- No navbar is rendered.
- A decorative full-viewport gym image sits over the dark base fallback.
- A dark left gradient keeps the authentication card readable.
- Login is initial and exposes username, password and submit.
- Link text is `¿No tienes cuenta? Créate una aquí` and swaps the card to registration locally.
- Registration renders every existing personal, address, account and password field, including
  password confirmation and two independent show/hide controls.
- Registration exposes `¿Ya tienes cuenta? Inicia sesión` to return locally.
- Image failure hides only the image; authentication remains complete and operable.

## Authenticated shell

- Navbar left: Lucide `Dumbbell` plus `FitPro`.
- Navbar center: `Rutinas`, `Sesiones`, `Ejercicios` as keyboard-operable local navigation.
- Active destination uses text, `aria-current` or tab state, and semantic secondary styling.
- Each destination shows the local gym image hero:
  - Rutinas: `Visualiza y ajusta tus rutinas`.
  - Sesiones: `Organiza tus sesiones de entrenamiento`.
  - Ejercicios: `Administra tu catálogo de ejercicios`.
- Logout/session status remains available without displacing centered navigation.

## Modal dialog behavior

- Use native `<dialog>` with daisyUI `modal` and `modal-box` classes.
- Opening focuses the dialog heading or first meaningful field.
- Closing returns focus to the trigger when it still exists.
- Read-only detail dialogs close by backdrop, Escape or explicit close without confirmation.
- Wizard dialogs expose an explicit close button and Cancel action in addition to backdrop/Escape.
- Nested dialogs are forbidden. Starting Edit from detail closes detail before opening its wizard.

## Wizard contract

### Step tabs

- Steps render as a `tablist` with one tab and tabpanel per step.
- The selected tab has a semantic colored bottom border, visible text and `aria-selected=true`.
- Arrow Left/Right, Home and End move roving tab focus; Enter/Space activates if focus movement does
  not activate automatically.
- Clicking tabs and using Previous/Next never validates or blocks navigation.
- Draft values survive every step change.

### Validation and save

- Full client validation occurs only when Save is activated.
- Invalid Save keeps the wizard open, focuses the error summary and activates the first step with
  an invalid field. All safe draft values remain.
- Server errors preserve draft and association ordering.
- Save is rendered and reachable only in `Resumen`. Previous steps expose only tab, Previous and
  Next buttons with `type=button`; pressing Enter outside Summary cannot submit or call the API.
- Save sends exactly one POST or PUT, shows pending state and blocks repeat submission.
- Save never asks for confirmation.
- Successful save inserts/replaces the card from the returned representation, closes the wizard,
  announces success and restores logical focus. It opens no detail modal and performs no GET.

### Dismissal

- Clicking the backdrop always asks:
  `Los cambios se perderán, ¿seguro deseas salir?`
- Rejecting preserves open state, active step, draft, errors and focus context.
- Accepting closes without POST or PUT.
- Clicking inside the dialog never invokes backdrop handling.
- Close, Cancel, Escape and workspace changes use the existing dirty-draft guard.
- While saving, backdrop, Escape, Close and Cancel are disabled or ignored.

## Exercise wizard

1. `Datos básicos`: name, optional description, optional image URL, optional video URL.
2. `Resumen`: entered text plus media references or explicit absent states.

Create uses POST. Edit preloads GET detail and uses PUT with the same complete writable shape.

## Session wizard

1. `Datos básicos`: name and optional description.
2. `Ejercicios`:
   - local search over owned exercises;
   - native checkbox multi-selection;
   - selected ordered list;
   - non-negative integer spinboxes for series and repetitions;
   - Lucide ArrowUp/ArrowDown and remove controls;
   - displayed order derived as `index + 1`.
3. `Resumen`: fields and ordered composition.

Filtering never deselects hidden items. An exercise cannot be selected twice. Empty composition is
valid. POST and PUT send the complete `exercises` array.

## Routine wizard

1. `Datos básicos`: name and optional description.
2. `Sesiones`:
   - local search over owned sessions;
   - an accessible `Agregar {session name}` action on every result;
   - results remain available after being added;
   - every activation appends a separate draft row with its own local key;
   - every new row starts with an unset required day select;
   - duplicate `(sessionId, day)` pairs prevented with textual feedback.
3. `Resumen`: fields and assignments ordered by day then session ID.

Empty assignments are valid. POST and PUT send the complete `sessions` array.

## Routine view and detail

- Grid cards show routine name and optional description. Their main labelled content button opens a
  read-only detail dialog; Edit/Delete remain separate lower-right footer actions.
- Each assigned session is a native collapsible showing day, name and description.
- Each session contains ordered exercise collapsibles showing thumbnail, name, series and
  repetitions before expansion.
- Exercise expansion shows description and the largest practical native video preview area.
- Empty routine/session and missing optional values have explicit text states.
- Every session/exercise disclosure is an independent `<details>` without a shared `name` or
  accordion state. Multiple siblings and nested disclosures may remain open simultaneously.

## Session view

- Each session is a native collapsible card showing name, optional description and `exerciseCount`
  from the session collection response; newly created/updated details derive it from array length.
- Expansion lists exercise name, series and repetitions compactly in execution order.
- Edit and Delete remain separate controls at the lower right and are not disclosure triggers.
- Session cards are independent disclosures; opening one never closes another.

## Exercise view

- Each exercise is a native collapsible card with large image area or placeholder and name.
- Expansion shows optional description and native video preview or fallback link.
- Edit and Delete remain separate controls.
- Exercise cards are independent disclosures; opening one never closes another.

## Searchable catalog selection

- A labelled `input type=search` filters names case-insensitively in memory.
- Session composition uses unique exercise checkboxes; hidden results retain selection.
- Routine assignment uses repeatable Add actions; selected count counts assignment rows rather than
  unique sessions.
- Result and no-match states are textual and announced.
- No API request occurs while typing.

## External media

- Images use lazy loading, async decoding, no-referrer policy, useful alt text and a dark fallback.
- Direct videos mount only after expansion and use native controls, `playsInline`, no autoplay and
  metadata preload. The browser determines format support; no URL-extension inference or probe runs.
- A descriptive external link with `target=_blank` and `rel="noopener noreferrer"` remains visible
  while loading, after success and after failure. Error replaces the player with textual fallback.
- YouTube, Vimeo and HTML pages remain links. No iframe, backend proxy, content validation, upload,
  platform integration or media transformation is added.

## Lucide usage

- Brand/navigation: `Dumbbell`, `CalendarDays`, `ListChecks`.
- Actions: `Plus`, `Eye`, `EyeOff`, `Pencil`, `Trash2`, `X`, `Save`, `ChevronLeft`, `ChevronRight`,
  `Search`, `ArrowUp`, `ArrowDown`, `Image`, `Video`, `ExternalLink`, `CircleAlert`.
- Icons beside visible text use `aria-hidden=true` and `focusable=false`.
- Icon-only close controls use contextual names such as `Cerrar detalle de Rutina A`.
- Tests locate actions by accessible name, never by SVG markup or test IDs.

## Responsive and accessibility acceptance

- At 320 px, all content uses one column, dialogs fit viewport and dynamic rows stack.
- At 768 px and 1280 px, cards and auth shell use available columns without horizontal overflow.
- At 200% zoom, navigation, dialog actions and form fields remain reachable.
- Interactive targets are at least 44 CSS pixels where practical.
- Focus is visible; errors/statuses are announced; state never depends on color alone.
- Axe covers rendered structure. Browser validation covers native dialog focus containment, image
  fallback, contrast and responsive layout.
