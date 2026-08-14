# Tasks: Rutinas de ejercicios y rediseño FitPro

**Input**: Artefactos de diseño en `/specs/002-exercise-routines/`

**Prerequisites**: `plan.md`, `spec.md`, `research.md`, `data-model.md`, `contracts/openapi.yaml`, `contracts/ui.md`, `quickstart.md`

**Tests**: Obligatorios. Tareas T001-T027 registran baseline implementado de US1-US5. Tareas pendientes corrigen el límite DAO, agregan `exerciseCount` y construyen interfaz FitPro.

**Organization**: Cada historia conserva tareas y criterio independiente. Los IDs completados preservan trazabilidad; los pendientes continúan secuencialmente.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Puede ejecutarse en paralelo en archivos distintos.
- **[Story]**: Vincula tarea con historia de `spec.md`.
- `[X]` registra implementación existente; `[ ]` identifica trabajo pendiente.

## Phase 1: Setup completado

**Purpose**: Persistencia, migraciones, autenticación y composición compartida.

- [X] T001 Aplicar esquema relacional, migraciones reversibles, orden de Compose y aislamiento de base de pruebas en `backend/migrations/002_create_exercise_routines.up.sql`, `backend/migrations/002_create_exercise_routines.down.sql`, `compose.yaml` y `backend/internal/routines/tests/testdb_test.go`
- [X] T002 Reutilizar middleware JWT y request context verificado para proteger rutas training en `backend/internal/identity/controller/controller.go`, `backend/internal/identity/controller/middleware.go`, `backend/internal/platform/requestctx/identity.go` y `backend/cmd/api/main.go`

---

## Phase 2: Foundation completada

**Purpose**: Errores, fronteras y registro de rutas compartidos.

- [X] T003 Definir errores públicos, DTO HTTP y DAO persistentes usados por CRUD en `backend/internal/routines/errors/errors.go`, `backend/internal/routines/dto/dto.go` y `backend/internal/routines/dao/dao.go`
- [X] T004 Registrar rutas autenticadas y mappings HTTP comunes sin acceso controller-repository en `backend/internal/routines/controller/controller.go` y `backend/cmd/api/main.go`

---

## Phase 3: User Story 1 - Crear ejercicios propios (Priority: P1) 🎯 MVP

**Goal**: Crear, listar y consultar ejercicios privados con textos y URLs validados.

**Independent Test**: Dos usuarios crean ejercicios y no pueden listar, consultar ni seleccionar contenido ajeno.

- [X] T005 [P] [US1] Probar validación, persistencia, HTTP autenticado y estados frontend de ejercicios en `backend/internal/routines/service/service_test.go`, `backend/internal/routines/repository/repository_test.go`, `backend/internal/routines/tests/integration_test.go` y `frontend/src/routines/ExercisesView.test.tsx`
- [X] T006 [US1] Implementar consultas owner-scoped y opcionales NULL para ejercicios en `backend/internal/routines/repository/repository.go`
- [X] T007 [US1] Implementar reglas de texto/URL y casos de uso de ejercicios en `backend/internal/routines/service/validation.go` y `backend/internal/routines/service/service.go`
- [X] T008 [US1] Implementar contratos HTTP y API TypeScript de ejercicios en `backend/internal/routines/controller/controller.go`, `frontend/src/routines/types.ts` y `frontend/src/routines/api.ts`
- [X] T009 [US1] Implementar catálogo, creación y detalle de ejercicios en `frontend/src/routines/ExercisesView.tsx` y `frontend/src/App.tsx`

---

## Phase 4: User Story 2 - Crear sesiones reutilizables (Priority: P2)

**Goal**: Crear sesiones privadas con ejercicios propios, cantidades y orden consecutivo.

**Independent Test**: Crear sesión vacía y compuesta, comprobar orden, reutilización, rollback y aislamiento.

- [X] T010 [P] [US2] Probar validaciones, transacción, contrato HTTP y composición frontend de sesiones en `backend/internal/routines/service/service_test.go`, `backend/internal/routines/repository/repository_test.go`, `backend/internal/routines/tests/integration_test.go` y `frontend/src/routines/SessionsView.test.tsx`
- [X] T011 [US2] Implementar selección bloqueada, inserción atómica y detalle ordenado de sesiones en `backend/internal/routines/repository/repository.go`
- [X] T012 [US2] Implementar validación y transacción service-owned de sesiones en `backend/internal/routines/service/validation.go` y `backend/internal/routines/service/service.go`
- [X] T013 [US2] Implementar handlers y contratos frontend de sesiones en `backend/internal/routines/controller/controller.go`, `frontend/src/routines/types.ts` y `frontend/src/routines/api.ts`
- [X] T014 [US2] Implementar selección, cantidades, reordenamiento y detalle de sesiones en `frontend/src/routines/SessionsView.tsx` y `frontend/src/App.tsx`

---

## Phase 5: User Story 3 - Crear y consultar rutinas (Priority: P3)

**Goal**: Crear rutinas privadas con sesiones reutilizables y días ISO.

**Independent Test**: Reutilizar sesiones en días distintos, rechazar pares duplicados y consultar detalle completo.

- [X] T015 [P] [US3] Probar reglas, atomicidad, contratos y detalle anidado de rutinas en `backend/internal/routines/service/service_test.go`, `backend/internal/routines/repository/repository_test.go`, `backend/internal/routines/tests/integration_test.go` y `frontend/src/routines/RoutinesView.test.tsx`
- [X] T016 [US3] Implementar persistencia y transacción de rutinas con asociaciones tenant-safe en `backend/internal/routines/repository/repository.go` y `backend/internal/routines/service/service.go`
- [X] T017 [US3] Implementar handlers y contratos frontend de rutinas en `backend/internal/routines/controller/controller.go`, `frontend/src/routines/types.ts` y `frontend/src/routines/api.ts`
- [X] T018 [US3] Implementar listado y detalle completo de rutinas en `frontend/src/routines/RoutinesView.tsx` y `frontend/src/App.tsx`

---

## Phase 6: User Story 4 - Eliminar contenido propio (Priority: P4)

**Goal**: Eliminar contenido propio conservando contenedores, reutilización y orden.

**Independent Test**: Cancelar y confirmar deletes, verificar cascadas, renumeración y aislamiento.

- [X] T019 [P] [US4] Probar deletes, cascadas, compactación, confirmación y errores en `backend/internal/routines/repository/repository_test.go`, `backend/internal/routines/tests/integration_test.go`, `frontend/src/routines/ExercisesView.test.tsx`, `frontend/src/routines/SessionsView.test.tsx` y `frontend/src/routines/RoutinesView.test.tsx`
- [X] T020 [US4] Implementar deletes owner-scoped, locks y compactación transaccional en `backend/internal/routines/repository/repository.go` y `backend/internal/routines/service/service.go`
- [X] T021 [US4] Implementar handlers y API frontend de eliminación en `backend/internal/routines/controller/controller.go` y `frontend/src/routines/api.ts`
- [X] T022 [US4] Implementar confirmación y actualización visual tras eliminación en `frontend/src/routines/ExercisesView.tsx`, `frontend/src/routines/SessionsView.tsx` y `frontend/src/routines/RoutinesView.tsx`

---

## Phase 7: User Story 5 - Editar contenido propio (Priority: P5)

**Goal**: Reemplazar entidades propias completamente, con atomicidad, propagación y descarte protegido.

**Independent Test**: Editar las tres entidades, vaciar asociaciones, verificar rollback, concurrencia, propagación y cancelación.

- [X] T023 [P] [US5] Probar validación, reemplazo, prioridad de errores, snapshot y concurrencia backend en `backend/internal/routines/service/service_test.go`, `backend/internal/routines/repository/repository_test.go` y `backend/internal/routines/tests/integration_test.go`
- [X] T024 [P] [US5] Probar prefill, PUT completo, dirty state, descarte y autenticación perdida en `frontend/src/routines/ExercisesView.test.tsx`, `frontend/src/routines/SessionsView.test.tsx`, `frontend/src/routines/RoutinesView.test.tsx` y `frontend/src/App.test.tsx`
- [X] T025 [US5] Implementar detalles de snapshot único y reemplazos owner-scoped atómicos en `backend/internal/routines/repository/repository.go`
- [X] T026 [US5] Implementar casos de uso y handlers PUT completos en `backend/internal/routines/service/service.go`, `backend/internal/routines/service/validation.go` y `backend/internal/routines/controller/controller.go`
- [X] T027 [US5] Implementar edición, prefill, dirty guard y PUT frontend en `frontend/src/routines/api.ts`, `frontend/src/routines/ExercisesView.tsx`, `frontend/src/routines/SessionsView.tsx`, `frontend/src/routines/RoutinesView.tsx` y `frontend/src/App.tsx`

---

## Phase 8: Architectural Boundary Remediation

**Purpose**: Cumplir la constitución impidiendo que estructuras DAO alcancen service o controller.

**⚠️ CRITICAL**: Completar esta fase antes del conteo backend o cualquier cambio posterior.

- [X] T028 [P] Definir estructuras mínimas del módulo para ejercicios, resúmenes, detalles y asociaciones, sin tags HTTP/DB ni reglas de negocio, en `backend/internal/routines/types.go`
- [X] T029 Convertir filas y parámetros DAO dentro de repository y cambiar sus métodos públicos para aceptar o devolver únicamente tipos del módulo, manteniendo DAO privados a persistencia, en `backend/internal/routines/repository/repository.go` y `backend/internal/routines/dao/dao.go`
- [X] T030 Actualizar casos de uso para consumir y devolver tipos del módulo, eliminar imports de `internal/routines/dao` desde service y conservar validaciones/transacciones actuales en `backend/internal/routines/service/service.go` y `backend/internal/routines/service/validation.go`
- [X] T031 Actualizar mappings controller y pruebas afectadas, verificar que controller/service no importen DAO y ejecutar suites enfocadas de límites, contratos y persistencia en `backend/internal/routines/controller/controller.go`, `backend/internal/routines/controller/controller_test.go`, `backend/internal/routines/service/service_test.go`, `backend/internal/routines/repository/repository_test.go` y `backend/internal/routines/tests/integration_test.go`

**Checkpoint**: Flujo efectivo queda HTTP → DTO → controller → service/module types → repository → DAO → PostgreSQL.

---

## Phase 9: Setup (Shared Infrastructure)

**Purpose**: Incorporar los únicos recursos y configuraciones compartidos requeridos por el
rediseño, sin cambiar persistencia ni agregar infraestructura.

- [X] T032 Instalar exactamente `lucide-react@1.31.0`, conservar dependencias existentes y actualizar el lockfile reproducible en `frontend/package.json` y `frontend/package-lock.json`
- [X] T033 Descargar una sola vez la imagen suministrada, redimensionarla y comprimirla como WebP razonable para web, verificar que no dependa del hotlink en ejecución y guardarla en `frontend/src/assets/fitpro-gym.webp`
- [X] T034 [P] Agregar un shim de jsdom para `HTMLDialogElement.showModal()` y `close()` que solo refleje el atributo `open`, sin simular un focus trap inexistente, en `frontend/src/test/setup.ts`
- [X] T035 [P] Ampliar CSP únicamente con `http:` y `https:` en `img-src` y con `media-src 'self' http: https:`, preservando todas las demás restricciones y `Referrer-Policy: no-referrer`, en `frontend/nginx.conf`

**Checkpoint**: Dependencia, asset, entorno de tests y política de medios quedan listos.

---

## Phase 10: Foundational (Blocking Shared UI)

**Purpose**: Construir únicamente los cuatro comportamientos compartidos que usan varios flujos.

**⚠️ CRITICAL**: Completar esta fase antes de implementar US6.

### Tests for shared UI *(write first and confirm meaningful failure)*

- [X] T036 [P] Probar apertura nativa, cierre por botón/backdrop/Escape, click interior, título accesible, bloqueo durante guardado y retorno de foco al disparador en `frontend/src/routines/components/ModalDialog.test.tsx`
- [X] T037 [P] Probar tabs con borde/semántica, roving focus con flechas/Home/End, salto libre entre pasos incompletos, ausencia de Guardar fuera de Resumen, Enter sin escritura, confirmación de backdrop, guard dirty en Close/Cancel/Escape y bloqueo de cierre/doble submit pendiente en `frontend/src/routines/components/WizardDialog.test.tsx`
- [X] T038 [P] Probar búsqueda local case-insensitive, resultados anunciados, estado sin coincidencias, conservación de selección oculta y acciones de resultado configurables en `frontend/src/routines/components/SearchableCatalog.test.tsx`
- [X] T039 [P] Probar imagen lazy con fallback, video nativo diferido con `controls`, `playsInline` y `preload="metadata"`, enlace seguro permanente, error textual y ausencia de `iframe` o probes en `frontend/src/routines/components/MediaPreview.test.tsx`

### Implementation for shared UI

- [X] T040 [P] Implementar envoltura mínima de `<dialog>` con daisyUI, nombre/descripción accesibles, foco inicial, backdrop distinguible, Escape interceptado, cierre explícito y restauración del disparador en `frontend/src/routines/components/ModalDialog.tsx`
- [X] T041 Implementar wizard sobre `ModalDialog` con tablist accesible, tabpanels, navegación libre, Previous/Next, Guardar solo en Resumen, cierre centralizado que reciba `dirty` y proteja Close/Cancel/Escape, y estado pending sin conocer entidades ni HTTP en `frontend/src/routines/components/WizardDialog.tsx`
- [X] T042 [P] Implementar catálogo buscable presentacional con `input type="search"`, filtrado local, conteos/estado anunciados y render prop o children concretos para selección única o agregado repetible en `frontend/src/routines/components/SearchableCatalog.tsx`
- [X] T043 [P] Implementar preview compartido de imagen/video externo con carga diferida, fallback oscuro/textual, enlace descriptivo seguro siempre visible y sin inferir formato ni consultar backend en `frontend/src/routines/components/MediaPreview.tsx`

**Checkpoint**: Diálogo, wizard, búsqueda y medios pueden validarse sin una entidad concreta.

---

## Phase 11: User Story 2 - Resumir sesiones reutilizables (Priority: P2)

**Goal**: Entregar `exerciseCount` desde el listado autenticado de sesiones mediante una consulta
agregada del backend, con cero para sesiones vacías y sin solicitudes de detalle por card.

**Independent Test**: Crear dos sesiones propias, una vacía y otra con varios ejercicios, solicitar
`GET /api/sessions` y verificar conteos `0` y `N`, orden por ID, aislamiento entre usuarios y una
sola fila por sesión.

### Tests for User Story 2 *(write first and confirm meaningful failure)*

- [X] T044 [P] [US2] Extender pruebas PostgreSQL para conteo cero/poblado, aislamiento por usuario, ausencia de multiplicación de filas y orden estable al listar sesiones en `backend/internal/routines/repository/repository_test.go`
- [X] T045 [P] [US2] Extender pruebas HTTP para exigir `exerciseCount` no negativo en cada resumen, comprobar cero/poblado y validar que no se expongan asociaciones ni datos ajenos en `backend/internal/routines/tests/integration_test.go`

### Implementation for User Story 2

- [X] T046 [P] [US2] Agregar `ExerciseCount int64` únicamente al tipo de resumen del módulo, sin tags HTTP ni campo almacenado, en `backend/internal/routines/types.go`
- [X] T047 [P] [US2] Agregar `exerciseCount` requerido con tipo `int64` al DTO de `SessionSummary`, sin incorporarlo a `SessionDetail`, en `backend/internal/routines/dto/dto.go`
- [X] T048 [US2] Reemplazar el listado de sesiones por una consulta owner-scoped con `LEFT JOIN session_exercises`, `COUNT(session_exercises.exercise_id)`, agrupación exacta y orden por ID, escaneando el conteo como `int64`, en `backend/internal/routines/repository/repository.go`
- [X] T049 [US2] Mapear explícitamente `ExerciseCount` desde el resultado del service hacia cada respuesta `SessionSummary`, conservando el contrato autenticado y errores existentes, en `backend/internal/routines/controller/controller.go`
- [X] T050 [P] [US2] Agregar `exerciseCount` a `SessionSummary` y redefinir `SessionDetail` como tipo independiente con `id`, `name`, `description` y `exercises`, evitando heredar el conteo; conservar `exercises.length` como proyección tras POST/PUT sin cambiar endpoints en `frontend/src/routines/types.ts`

**Checkpoint**: Cada card puede mostrar el conteo inicial mediante una única llamada de colección.

---

## Phase 12: User Story 6 - Gestionar contenido desde interfaz FitPro (Priority: P6)

**Goal**: Reemplazar la interfaz existente por autenticación visual, navbar y héroes FitPro, cards
progresivas, detalles modales y wizards accesibles para crear/editar las tres entidades.

**Independent Test**: Ingresar desde login, alternar a registro, autenticar, navegar las tres vistas,
abrir varios detalles simultáneos y crear/editar cada entidad mediante sus pasos, comprobando
descarte, medios, foco, respuesta del servidor y ausencia de requests redundantes.

### Tests for User Story 6 *(write first and confirm meaningful failure)*

- [X] T051 [P] [US6] Probar shell sin navbar, login inicial, enlaces exactos de intercambio login/registro, formulario completo, dos toggles de contraseña, imagen decorativa y operación con fallback en `frontend/src/auth/AuthShell.test.tsx`, `frontend/src/auth/LoginForm.test.tsx` y `frontend/src/auth/RegisterForm.test.tsx`
- [X] T052 [P] [US6] Probar marca FitPro con `Dumbbell`, navegación centrada por click/teclado, `aria-current`, héroes por apartado, logout y guard de draft al cambiar de vista en `frontend/src/App.test.tsx`
- [X] T053 [P] [US6] Reescribir pruebas de ejercicios para wizard de dos pasos, navegación incompleta, validación final, create/edit prefill, payload exacto, descarte, respuesta sin GET, cards independientes, preview/fallback y acciones accesibles en `frontend/src/routines/ExercisesView.test.tsx`
- [X] T054 [P] [US6] Reescribir pruebas de sesiones para wizard de tres pasos, búsqueda única, selección oculta, cantidades, reordenamiento `1..N`, resumen, create/edit, baseline normalizada, Cancelar/Close/Escape dirty aceptado o rechazado, `exerciseCount` del backend y proyectado tras save, cards independientes y listado compacto en `frontend/src/routines/SessionsView.test.tsx`
- [X] T055 [P] [US6] Reescribir pruebas de rutinas para búsqueda repetible, filas con keys/días independientes, misma sesión en días distintos, duplicado bloqueado, resumen, create/edit, baseline canónica, Cancelar/Close/Escape dirty aceptado o rechazado, respuesta sin GET, card/modal, jerarquía anidada y disclosures simultáneos en `frontend/src/routines/RoutinesView.test.tsx`
- [X] T056 [P] [US6] Ampliar axe y teclado para auth, navbar, héroes, dialogs, wizards, errores, controles dinámicos, iconos, medios, retorno de foco y estados abiertos simultáneos en `frontend/src/App.accessibility.test.tsx`

### Implementation for User Story 6

- [X] T057 [P] [US6] Crear `AuthShell` con asset local decorativo, fallback oscuro, degradado izquierdo, card responsiva y estado local inicial de login que intercambie formularios sin navbar ni navegación externa en `frontend/src/auth/AuthShell.tsx`
- [X] T058 [US6] Adaptar login y registro al shell visual, incorporar enlaces exactos de intercambio y usar `Eye`/`EyeOff` con texto o nombres accesibles para ambos campos de contraseña sin cambiar contratos auth en `frontend/src/auth/LoginForm.tsx` y `frontend/src/auth/RegisterForm.tsx`
- [X] T059 [US6] Reorganizar shell autenticado con navbar FitPro, `Dumbbell`, controles centrados con iconos nombrados, SessionStatus/logout, navegación local accesible y coordinación del dirty guard; renderizar `AuthShell` sin navbar cuando no hay token en `frontend/src/App.tsx` y `frontend/src/auth/SessionStatus.tsx`
- [X] T060 [P] [US6] Implementar `ExerciseWizard` con modo create/edit, baseline/draft local, pasos `Datos básicos` y `Resumen`, validación completa solo al guardar, URLs opcionales, errores por paso, pending, descarte y payload POST/PUT completo en `frontend/src/routines/ExerciseWizard.tsx`
- [X] T061 [US6] Rehacer vista de ejercicios con héroe, CTA, grid de `<details>` independientes, imagen amplia, descripción/video progresivo, acciones Lucide separadas, delete existente y orquestación del wizard sin GET posterior al save en `frontend/src/routines/ExercisesView.tsx`
- [X] T062 [P] [US6] Implementar `SessionWizard` con modo create/edit, baseline inmutable, comparación dirty de campos y composición ordenada, descarte protegido, pasos `Datos básicos`, `Ejercicios` y `Resumen`, catálogo buscable único, checkboxes, filas ordenadas, spinboxes, `ArrowUp`/`ArrowDown`, orden derivado y payload completo en `frontend/src/routines/SessionWizard.tsx`
- [X] T063 [US6] Rehacer vista de sesiones con héroe, CTA, cards `<details>` independientes, `exerciseCount` del listado, carga de detalle al expandir/editar, ejercicios compactos, acciones separadas y proyección `saved.exercises.length` después de POST/PUT sin GET adicional en `frontend/src/routines/SessionsView.tsx`
- [X] T064 [P] [US6] Implementar `RoutineWizard` con modo create/edit, baseline canónica por pares `(sessionId, day)`, comparación dirty y descarte protegido, pasos `Datos básicos`, `Sesiones` y `Resumen`, búsqueda con acción `Agregar` persistente, row key local, día inicialmente vacío, múltiples días y validación de par duplicado en `frontend/src/routines/RoutineWizard.tsx`
- [X] T065 [US6] Rehacer vista de rutinas con héroe, CTA, cards que abren detalle mediante `ModalDialog`, acciones separadas, sesiones y ejercicios como `<details>` independientes anidados, días, cantidades y `MediaPreview`, aplicando POST/PUT a la card sin GET ni apertura automática en `frontend/src/routines/RoutinesView.tsx`

**Checkpoint**: FitPro satisface auth, navegación, cards, detalles y los tres wizards de extremo a extremo.

---

## Phase 13: Polish & Cross-Cutting Validation

**Purpose**: Verificar integración, seguridad, accesibilidad, estilos y ejecución reproducible.

- [X] T066 [P] Auditar imports nombrados Lucide, `currentColor`, iconos decorativos ocultos, botones icon-only con nombre contextual, targets táctiles, clases semánticas daisyUI y ausencia de hex/CSS específico global en `frontend/src/App.tsx`, `frontend/src/auth/`, `frontend/src/routines/` y `frontend/src/styles.css`
- [X] T067 [P] Ejecutar `gofmt`, `go test ./...`, pruebas con `-tags=integration` y `go vet ./...` desde `backend/`; resolver únicamente regresiones del conteo y confirmar que no existe migración nueva en `backend/internal/routines/`
- [X] T068 Ejecutar `npm test -- --run` y `npm run build` desde `frontend/`; corregir fallos de tipos, accesibilidad, diálogo, medios, navegación, conteos y comportamiento de wizards en `frontend/src/`
- [ ] T069 Levantar `docker compose up --build` y ejecutar los escenarios 9-12 a 320/768/1280 px y zoom 200%, verificando asset local/fallback, CSP, medios, foco nativo, disclosures independientes, un solo write sin GET y conteos backend conforme `specs/002-exercise-routines/quickstart.md`

---

## Dependencies & Execution Order

### Phase Dependencies

```text
Phases 1-7 completed baseline
    ↓
Phase 8 Boundary repair
    ↓
Phase 9 Setup
    ↓
Phase 10 Shared UI foundation
    ├───────────────┐
    ↓               ↓
US2 count       US6 test preparation
    └───────┬───────┘
            ↓
       US6 implementation
            ↓
    Phase 13 validation
```

- **Phase 8** corrige el límite DAO y bloquea todo cambio backend posterior.
- **Phase 9** depende de Phase 8 para conservar compilación y contratos internos.
- **Phase 10** depende del setup y bloquea componentes de US6.
- **US2** puede desarrollarse después del setup y debe terminar antes de integrar las cards de
  sesiones de US6.
- **US6 tests** pueden escribirse tras Foundation y en paralelo con backend US2.
- **US6 implementation** depende de componentes compartidos; `SessionsView` depende además de US2.
- **Phase 13** depende de todo el alcance seleccionado.

### User Story Dependencies

- **US2 (P2)**: Conserva sesiones existentes. Solo agrega un resumen derivado y es comprobable con
  datos sembrados, sin US6.
- **US6 (P6)**: Reutiliza CRUD completo de US1-US5 y depende del contrato `exerciseCount` de US2
  únicamente para las cards iniciales de sesiones.
- **US1, US3, US4 y US5**: No reciben tareas nuevas. Sus implementaciones y pruebas existentes
  permanecen como regresión obligatoria en T067-T068.

### Within Shared UI

1. Escribir T036-T039 y comprobar fallos por componentes ausentes.
2. Implementar ModalDialog antes de WizardDialog.
3. Implementar SearchableCatalog y MediaPreview en paralelo.
4. Ejecutar las cuatro suites enfocadas antes de US6.

### Within US2

1. Escribir T044-T045 y observar la ausencia de `exerciseCount`.
2. Agregar DAO y DTO en paralelo.
3. Implementar consulta agregada owner-scoped.
4. Mapear respuesta HTTP y actualizar contrato TypeScript.
5. Ejecutar tests repository/HTTP antes de integrar SessionView.

### Within US6

1. Escribir T051-T056 y confirmar fallos por la interfaz anterior.
2. Implementar AuthShell y luego integrar formularios.
3. Implementar shell autenticado después de AuthShell.
4. Implementar los tres wizards concretos en paralelo.
5. Integrar cada vista después de su wizard; SessionView espera `exerciseCount`.
6. Ejecutar suites enfocadas y axe antes de validación total.

## Parallel Opportunities

- T034 y T035 pueden ejecutarse en paralelo con T032-T033.
- T036-T039 prueban archivos compartidos distintos y pueden escribirse en paralelo.
- T040, T042 y T043 implementan componentes distintos; T041 espera solamente T040.
- T044 y T045 pueden escribirse en paralelo; T046, T047 y T050 usan paquetes distintos.
- T051-T056 son suites frontend separadas y pueden prepararse en paralelo.
- T057, T060, T062 y T064 crean componentes independientes después de Foundation.
- T061, T063 y T065 integran vistas distintas, aunque cada una espera su wizard.
- T066 y T067 revisan frontend/backend por separado.

## Parallel Examples

### Shared UI

```text
T036: Modal behavior tests in frontend/src/routines/components/ModalDialog.test.tsx
T038: Catalog filtering tests in frontend/src/routines/components/SearchableCatalog.test.tsx
T039: Media fallback tests in frontend/src/routines/components/MediaPreview.test.tsx
```

### User Story 2

```text
T044: PostgreSQL aggregate tests in backend/internal/routines/repository/repository_test.go
T045: HTTP summary contract tests in backend/internal/routines/tests/integration_test.go
T050: TypeScript SessionSummary contract in frontend/src/routines/types.ts
```

### User Story 6

```text
T051: Authentication shell tests in frontend/src/auth/AuthShell.test.tsx
T053: Exercise wizard/view tests in frontend/src/routines/ExercisesView.test.tsx
T054: Session wizard/view tests in frontend/src/routines/SessionsView.test.tsx
T055: Routine wizard/view tests in frontend/src/routines/RoutinesView.test.tsx
T056: Cross-surface accessibility tests in frontend/src/App.accessibility.test.tsx
```

## Implementation Strategy

### MVP First

1. Completar Setup.
2. Completar Shared UI.
3. Completar US2 `exerciseCount`.
4. Validar listado sin N+1.
5. Integrar SessionView como primer corte demostrable.

### Incremental Delivery

1. Setup deja asset, dependencia y CSP reproducibles.
2. Foundation entrega componentes UI probables aisladamente.
3. US2 entrega conteos correctos desde backend.
4. US6 entrega auth y navbar.
5. US6 agrega ExerciseWizard y ExercisesView.
6. US6 agrega SessionWizard y SessionsView.
7. US6 agrega RoutineWizard y RoutinesView.
8. Polish valida Docker, medios y responsive.

## Notes

- `[P]` exige archivos distintos y ausencia de dependencias pendientes.
- Tests preceden implementación y deben fallar por la conducta esperada.
- Backend conserva `controller -> service -> repository`; DTO y DAO no son capas nuevas.
- `exerciseCount` se calcula en PostgreSQL; no se persiste ni requiere migración.
- POST/PUT exitoso actualiza cards desde su respuesta; no dispara GET adicional.
- Los medios externos nunca pasan por backend ni usan iframe.
- CSS específico permanece junto al componente mediante utilidades Tailwind/daisyUI.
- No editar README, agregar router, store global, librería modal/formularios ni segunda paleta.
