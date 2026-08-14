# Feature Specification: Rutinas de ejercicios

**Feature Branch**: `no creada (sin hook configurado)`

**Created**: 2026-08-13

**Status**: Draft

**Input**: User description: "Añadir creación, consulta, edición y eliminación de ejercicios, sesiones y
rutinas privadas por usuario. Las rutinas seleccionan sesiones existentes y asignan a cada una un
día de la semana; las sesiones seleccionan ejercicios existentes y asignan series y repeticiones.
Cada ejercicio seleccionado también tiene un orden de ejecución dentro de su sesión. Todas las
operaciones requieren autenticación. Rediseñar la interfaz como FitPro con autenticación visual,
navegación autenticada, cards, detalles colapsables y wizards modales para crear o editar."

## Clarifications

### Session 2026-08-13

- Q: ¿Cómo debe validarse el orden de ejercicios dentro de cada sesión? → A: Valores únicos y
  consecutivos desde 1 hasta la cantidad de ejercicios.
- Q: ¿Un mismo ejercicio puede aparecer varias veces dentro de una sesión? → A: No; cada ejercicio
  puede aparecer una sola vez por sesión.
- Q: ¿Cómo define el usuario el orden de ejecución? → A: Ordena la lista de ejercicios y el sistema
  asigna automáticamente valores consecutivos desde 1.
- Q: ¿Una misma sesión puede aparecer varias veces dentro de una rutina? → A: Sí, puede repetirse
  en días diferentes, pero no más de una vez en el mismo día.
- Q: ¿Qué ocurre con el orden cuando se elimina un ejercicio usado en sesiones? → A: Cada sesión
  afectada renumera automáticamente sus ejercicios restantes desde 1.
- Q: ¿Qué ocurre en los contenedores cuando se edita un ejercicio o sesión reutilizada? → A: Todas
  las sesiones y rutinas muestran inmediatamente los datos actualizados.
- Q: Si una actualización contiene varios problemas, ¿qué prioridad tienen los errores? → A:
  Primero se valida la estructura, después la existencia del recurso objetivo y finalmente las
  referencias seleccionadas.
- Q: ¿Qué consistencia tiene una consulta detallada durante una edición concurrente? → A: Cada
  respuesta representa un único estado confirmado, sin mezclar datos anteriores y nuevos.
- Q: ¿Qué ocurre al abandonar una edición con cambios sin guardar? → A: La interfaz solicita
  confirmación antes de descartarlos.
- Q: ¿Qué resultado produce un identificador inválido? → A: Texto no numérico, cero, negativos o
  valores fuera del rango admitido producen una solicitud inválida; únicamente identificadores
  válidos ajenos o inexistentes producen un resultado de recurso no encontrado.
- Q: ¿Puede una persona avanzar en el wizard con campos incompletos? → A: Sí; puede cambiar a
  cualquier paso y la validación completa se aplica solamente al guardar.
- Q: ¿Qué ocurre al hacer click fuera del wizard? → A: Se solicita confirmación con una advertencia
  equivalente a “Los cambios se perderán, ¿seguro deseas salir?”; guardar nunca pide confirmación.
- Q: ¿Qué biblioteca de iconos usa la interfaz? → A: `lucide-react`, con iconos acompañados por
  texto o nombres accesibles y sin usar iconos como único indicador de estado.
- Q: ¿Cómo deben mostrarse las URLs de video? → A: Usar reproductor nativo para videos directos y
  mostrar un enlace externo si el navegador no puede reproducirlos.
- Q: ¿Qué ocurre después de guardar exitosamente desde el wizard? → A: Cerrar el wizard, actualizar
  la card correspondiente con la respuesta y anunciar el éxito.
- Q: ¿Cómo se asigna una sesión a varios días desde el wizard? → A: La misma sesión puede
  seleccionarse repetidamente desde el buscador; cada selección crea una fila de asignación.
- Q: ¿Dónde se muestra la acción Guardar del wizard? → A: Solamente en el paso final “Resumen”.
- Q: ¿Cómo deben comportarse los elementos colapsables? → A: Varios elementos del mismo nivel o
  de niveles anidados pueden permanecer abiertos simultáneamente.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Crear ejercicios propios (Priority: P1)

Una persona autenticada registra ejercicios que luego podrá reutilizar al formar sus sesiones. Cada
ejercicio identifica el movimiento y puede incluir material visual de referencia.

**Why this priority**: Los ejercicios son el catálogo mínimo necesario para construir sesiones y,
posteriormente, rutinas.

**Independent Test**: Una persona autenticada crea un ejercicio con nombre y campos opcionales,
consulta sus detalles y comprueba que otra persona autenticada no puede encontrarlo ni utilizarlo.

**Acceptance Scenarios**:

1. **Given** una persona autenticada, **When** crea un ejercicio con nombre válido, **Then** el
   ejercicio queda asociado exclusivamente a su cuenta y aparece en su catálogo.
2. **Given** datos opcionales válidos, **When** crea un ejercicio con descripción, URL de imagen y
   URL de video, **Then** puede consultar posteriormente todos esos datos.
3. **Given** un ejercicio perteneciente a otra cuenta, **When** una persona intenta consultarlo,
   eliminarlo o incluirlo en una sesión, **Then** el sistema no revela su existencia ni permite la
   operación.
4. **Given** una persona no autenticada, **When** intenta crear, consultar o eliminar ejercicios,
   **Then** el sistema rechaza la operación.

---

### User Story 2 - Crear sesiones reutilizables (Priority: P2)

Una persona autenticada crea una sesión con nombre, descripción opcional y ejercicios elegidos de
su propio catálogo. Para cada ejercicio seleccionado indica series, repeticiones y orden de
ejecución.

**Why this priority**: Una sesión organiza ejercicios y sus cantidades, y constituye el siguiente
nivel necesario para formar una rutina.

**Independent Test**: Con ejercicios propios ya creados, la persona forma una sesión, asigna
cantidades y un orden consecutivo a cada ejercicio, y consulta la composición completa sin acceder
a datos ajenos.

**Acceptance Scenarios**:

1. **Given** ejercicios propios existentes, **When** la persona crea una sesión y asigna series,
   repeticiones no negativas y un orden consecutivo a cada ejercicio elegido, **Then** la sesión
   conserva esa composición y su secuencia de ejecución.
2. **Given** un ejercicio propio, **When** se incluye en más de una sesión, **Then** cada sesión
   conserva independientemente sus propias cantidades.
3. **Given** un ejercicio ya incluido en una sesión, **When** se intenta seleccionarlo nuevamente
   en esa misma sesión, **Then** el sistema rechaza la duplicación y no crea la sesión parcialmente.
4. **Given** un ejercicio ajeno o inexistente, **When** se intenta incluir en una sesión, **Then** la
   sesión no se crea y el sistema no revela información del ejercicio.
5. **Given** una sesión propia, **When** la persona consulta sus detalles, **Then** ve nombre,
   descripción y todos sus ejercicios ordenados para su ejecución, con series y repeticiones.
6. **Given** varios ejercicios seleccionados, **When** la persona cambia sus posiciones en la lista,
   **Then** el sistema actualiza automáticamente sus órdenes consecutivos antes de crear la sesión.

---

### User Story 3 - Crear y consultar rutinas (Priority: P3)

Una persona autenticada crea una rutina con nombre, descripción opcional y sesiones elegidas de su
propio catálogo. Para cada sesión seleccionada asigna un día de la semana entre 1 y 7. Después puede
consultar sus rutinas y recorrer toda la información hasta los ejercicios.

**Why this priority**: La rutina entrega el resultado principal de la feature, pero depende de que
existan sesiones y ejercicios reutilizables.

**Independent Test**: Con sesiones propias preparadas, la persona crea una rutina, asigna días,
abre el apartado de rutinas y verifica la estructura completa de sesiones y ejercicios.

**Acceptance Scenarios**:

1. **Given** sesiones propias existentes, **When** la persona crea una rutina con días válidos,
   **Then** la rutina conserva cada sesión junto con su día asignado.
2. **Given** una sesión propia, **When** se selecciona en más de una rutina, **Then** cada rutina
   puede asignarle su propio día sin modificar las demás.
3. **Given** una sesión propia, **When** se asigna a varios días distintos de una misma rutina,
   **Then** la rutina conserva una asignación separada para cada día.
4. **Given** una sesión ya asignada a un día, **When** se intenta asignarla nuevamente al mismo día
   de la misma rutina, **Then** el sistema rechaza la duplicación sin crear la rutina parcialmente.
5. **Given** una sesión ajena o inexistente, **When** se intenta incluir en una rutina, **Then** la
   rutina no se crea y el sistema no revela información de la sesión.
6. **Given** varias rutinas propias, **When** la persona abre el apartado de rutinas, **Then** ve
   solamente sus rutinas y puede consultar todos sus datos, sesiones y ejercicios asociados.

---

### User Story 4 - Eliminar contenido propio (Priority: P4)

Una persona autenticada elimina rutinas, sesiones y ejercicios que ya no desea conservar, sin poder
afectar contenido perteneciente a otra cuenta.

**Why this priority**: Completa el ciclo de gestión solicitado y permite retirar información, pero
no es necesaria para demostrar la creación y consulta principales.

**Independent Test**: La persona elimina una entidad propia y verifica el resultado, luego intenta
eliminar una entidad ajena y confirma que no obtiene información ni produce cambios.

**Acceptance Scenarios**:

1. **Given** una rutina propia, **When** la persona confirma su eliminación, **Then** la rutina y
   sus asignaciones desaparecen, mientras las sesiones y ejercicios reutilizables permanecen.
2. **Given** una entidad ajena o inexistente, **When** se intenta eliminar, **Then** el sistema
   responde de forma indistinguible y no modifica datos.
3. **Given** una sesión o ejercicio utilizado por contenido propio, **When** la persona intenta
   eliminarlo, **Then** el sistema elimina la entidad y retira sus asociaciones de las rutinas o
   sesiones que la utilizaban, sin eliminar esos contenedores.
4. **Given** un ejercicio utilizado en una o más sesiones, **When** la persona lo elimina, **Then**
   cada sesión afectada conserva sus ejercicios restantes renumerados consecutivamente desde 1.

---

### User Story 5 - Editar contenido propio (Priority: P5)

Una persona autenticada modifica ejercicios, sesiones y rutinas ya creados sin cambiar su identidad
ni propietario. La edición de una sesión o rutina reemplaza su composición completa.

**Why this priority**: Permite corregir y mantener contenido reutilizable después de crearlo, sin
agregar historial, versionado ni edición parcial compleja.

**Independent Test**: La persona abre cada entidad propia, entra en edición, modifica sus campos y
asociaciones, guarda y comprueba el detalle completo actualizado; cancelar no persiste cambios.

**Acceptance Scenarios**:

1. **Given** un ejercicio propio, **When** la persona edita sus campos y elimina valores opcionales,
   **Then** conserva el mismo identificador y muestra únicamente los nuevos valores.
2. **Given** una sesión propia, **When** cambia nombre, descripción, cantidades, ejercicios u orden
   y guarda, **Then** la composición completa anterior se reemplaza atómicamente por la nueva.
3. **Given** una rutina propia, **When** cambia nombre, descripción, sesiones o días y guarda,
   **Then** las asignaciones anteriores se reemplazan atómicamente por las nuevas.
4. **Given** una sesión o rutina propia, **When** guarda una lista vacía, **Then** la entidad se
   conserva sin asociaciones.
5. **Given** una entidad ajena o inexistente, **When** intenta editarla, **Then** el sistema no revela
   su existencia ni modifica datos.
6. **Given** una referencia seleccionada que es ajena, inexistente o dejó de estar disponible,
   **When** confirma la edición, **Then** falla toda la operación y permanece el estado anterior.
7. **Given** un formulario de edición con cambios sin guardar, **When** la persona cancela, **Then**
   no se envía una actualización y la entidad persistida permanece intacta.
8. **Given** un ejercicio o sesión reutilizada, **When** la persona guarda su edición, **Then** todas
   las sesiones o rutinas que la referencian muestran inmediatamente sus datos actualizados.
9. **Given** una actualización estructuralmente válida cuyo objetivo no existe o es ajeno y además
   contiene referencias no disponibles, **When** se procesa, **Then** responde como recurso no
   encontrado sin informar primero errores sobre las referencias internas.
10. **Given** una edición con cambios sin guardar, **When** la persona cancela o intenta cambiar de
    apartado, **Then** la interfaz solicita confirmación y conserva la edición si no acepta el
    descarte.

---

### User Story 6 - Gestionar contenido desde la interfaz FitPro (Priority: P6)

Una persona usa una interfaz visual oscura y consistente para autenticarse, navegar entre rutinas,
sesiones y ejercicios, consultar detalles progresivos y completar creaciones o ediciones mediante
un wizard modal con pasos visibles.

**Why this priority**: Mejora la comprensión y operación de capacidades ya existentes sin cambiar
sus reglas, persistencia ni contratos HTTP.

**Independent Test**: Una persona entra desde login, alterna a registro, inicia sesión, recorre las
tres vistas, abre detalles y completa o abandona los tres wizards usando mouse y teclado.

**Acceptance Scenarios**:

1. **Given** una persona no autenticada, **When** abre la aplicación, **Then** ve la imagen de
   gimnasio indicada como fondo, un degradado oscuro izquierdo y una card de login sin navbar.
2. **Given** la card de login, **When** activa “¿No tienes cuenta? Créate una aquí”, **Then** la
   misma superficie cambia a registro con todos los campos existentes, confirmación de contraseña
   y controles para ver ambas contraseñas.
3. **Given** una persona autenticada, **When** entra al workspace, **Then** ve la marca FitPro con
   icono de mancuerna a la izquierda, navegación centrada y un hero correspondiente al apartado.
4. **Given** una creación o edición, **When** abre el wizard, **Then** puede recorrer sus pasos por
   las tabs con borde inferior coloreado aunque existan campos incompletos.
5. **Given** un wizard abierto, **When** hace click en el backdrop, **Then** la interfaz solicita
   confirmación antes de cerrar y nunca guarda automáticamente.
6. **Given** el último paso del wizard, **When** guarda, **Then** no aparece una confirmación extra;
   el formulario se envía una sola vez y el resultado actualiza la vista.
7. **Given** una rutina, **When** abre su detalle, **Then** ve sesiones colapsables por día y dentro
   ejercicios colapsables con miniatura, cantidades, descripción y video cuando estén disponibles.
8. **Given** sesiones o ejercicios propios, **When** recorre sus cards, **Then** puede expandir sus
   detalles y dispone de acciones accesibles para editar y eliminar.

### Edge Cases

- Una rutina o sesión se crea sin elementos seleccionados.
- Una misma sesión se reutiliza en varias rutinas con días distintos.
- Una misma sesión se asigna a varios días distintos dentro de una rutina.
- Una misma sesión se intenta asignar dos veces al mismo día dentro de una rutina.
- Un mismo ejercicio se reutiliza en varias sesiones con cantidades distintas.
- Un mismo ejercicio se envía más de una vez dentro de una única sesión.
- Los órdenes de una sesión contienen duplicados, comienzan en 0 o dejan huecos entre valores.
- La persona cambia varias veces la posición de un ejercicio antes de crear la sesión.
- Una rutina asigna el mismo día a más de una sesión; esto se admite mientras no se establezca una
  regla contraria.
- El día vale exactamente 1 o 7, o queda fuera de ese intervalo.
- La correspondencia semanal es 1 lunes, 2 martes, 3 miércoles, 4 jueves, 5 viernes, 6 sábado y 7
  domingo.
- Series o repeticiones valen 0, son negativas, contienen decimales o exceden el rango entero
  admitido.
- Un texto tiene espacios exteriores, dos espacios consecutivos, tabulaciones o saltos de línea.
- Una descripción opcional se omite o se envía vacía.
- Una URL opcional se omite, usa una dirección relativa o emplea un esquema distinto de HTTP/HTTPS.
- Una entidad seleccionada se elimina o deja de pertenecer al usuario antes de confirmar la
  creación o edición del elemento que la referencia.
- Una edición elimina todos los ejercicios de una sesión o todas las sesiones de una rutina.
- Una edición elimina una descripción o URL opcional previamente guardada.
- Dos ediciones concurrentes del mismo recurso terminan en uno de los estados completos enviados,
  sin mezclar parcialmente sus asociaciones.
- Una consulta detallada coincide con un único estado confirmado aunque una edición concurrente
  modifique campos o asociaciones durante la lectura.
- La persona cancela después de reordenar o quitar elementos en el formulario de edición.
- La persona intenta cambiar de apartado con una edición modificada y rechaza el descarte.
- La persona cambia entre pasos con campos obligatorios vacíos y regresa sin perder el draft.
- La búsqueda de un multiselect no obtiene coincidencias o oculta elementos ya seleccionados.
- Una sesión elegida para una rutina se asigna a varios días distintos desde el wizard.
- La persona hace click dentro del wizard y el evento no se interpreta como click en el backdrop.
- La persona hace click fuera mientras una operación de guardado está pendiente.
- Una imagen o video externo no carga, rechaza embedding o usa un formato no reproducible por el
  navegador.
- La interfaz se usa a 320 px, con zoom al 200% o con textos y URLs largos.
- Se elimina el primer ejercicio, uno intermedio o el último de una sesión con varios ejercicios.
- Se solicita una entidad con un identificador inexistente o perteneciente a otra cuenta.
- Se usa como identificador texto no numérico, cero, un negativo o un entero fuera del
  rango admitido.
- Dos usuarios crean entidades con el mismo nombre; no existe una regla de unicidad global.

## Scope Boundaries *(mandatory)*

- **In scope**: Crear, listar, consultar en detalle, editar mediante reemplazo completo y eliminar
  ejercicios, sesiones y rutinas propios; seleccionar ejercicios existentes al crear o editar
  sesiones; seleccionar sesiones existentes y asignar días al crear o editar rutinas; validar
  textos, URLs, enteros y pertenencia; mostrar solamente contenido de la persona autenticada;
  rediseñar autenticación y workspace; wizards modales; búsqueda local dentro de catálogos ya
  cargados; previews progresivos de medios externos; iconos Lucide.
- **Out of scope**: Edición parcial por campos, historial o resolución interactiva de conflictos,
  duplicar entidades, compartirlas, publicarlas, usar
  plantillas, registrar ejecución o progreso, controlar pesos o descansos, ordenar elementos
  fuera del flujo de creación/edición, búsqueda o filtrado en backend, paginar, adjuntar archivos,
  alojar/procesar medios y administrar permisos adicionales.
- **Simplicity rationale**: La feature incorpora únicamente tres catálogos privados y sus
  asociaciones necesarias. La edición reemplaza formularios completos y reutiliza las mismas
  validaciones, evitando contratos parciales, historial o capacidades anticipadas.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Toda creación, listado, consulta detallada, edición y eliminación de rutinas, sesiones o
  ejercicios DEBE requerir una persona autenticada.
- **FR-002**: Cada rutina, sesión y ejercicio DEBE pertenecer exactamente a la cuenta que lo creó.
- **FR-003**: El sistema DEBE limitar cada listado a entidades pertenecientes a la persona
  autenticada.
- **FR-004**: El sistema NO DEBE permitir consultar, editar, eliminar ni seleccionar una entidad de otra
  cuenta mediante su identificador.
- **FR-005**: Una entidad ajena y una inexistente DEBEN producir resultados públicamente
  indistinguibles para evitar revelar su existencia.
- **FR-006**: Un ejercicio DEBE contener nombre obligatorio y PUEDE contener descripción, URL de
  imagen y URL de video.
- **FR-007**: Una sesión DEBE contener nombre obligatorio, PUEDE contener descripción y DEBE
  conservar el conjunto de ejercicios propios seleccionado al crearla.
- **FR-008**: Cada ejercicio seleccionado en una sesión DEBE tener cantidades enteras de series y
  repeticiones, y ambas DEBEN admitir 0 pero NO valores negativos ni decimales.
- **FR-009**: Cada ejercicio seleccionado en una sesión DEBE tener un orden de ejecución entero,
  único dentro de esa sesión y consecutivo desde 1 hasta la cantidad total de ejercicios
  seleccionados.
- **FR-010**: Un ejercicio NO DEBE aparecer más de una vez dentro de la misma sesión, aunque PUEDE
  reutilizarse en otras sesiones del mismo usuario.
- **FR-011**: La interfaz DEBE permitir ordenar la lista de ejercicios seleccionados y DEBE asignar
  automáticamente los valores de orden `1..N` según sus posiciones; la persona NO DEBE ingresar
  esos números manualmente.
- **FR-012**: Una rutina DEBE contener nombre obligatorio, PUEDE contener descripción y DEBE
  conservar el conjunto de sesiones propias seleccionado al crearla.
- **FR-013**: Cada sesión seleccionada en una rutina DEBE tener exactamente un día asignado mediante
  un entero entre 1 y 7 inclusive.
- **FR-014**: La interpretación de los días DEBE ser 1 lunes, 2 martes, 3 miércoles, 4 jueves, 5
  viernes, 6 sábado y 7 domingo.
- **FR-015**: Sesiones y ejercicios DEBEN poder reutilizarse en más de una entidad contenedora del
  mismo usuario sin compartir los valores propios de cada asociación.
- **FR-016**: La creación de una sesión o rutina DEBE ser completa: si cualquier entidad elegida es
  ajena, inexistente o inválida, no se debe crear parcialmente el nuevo elemento.
- **FR-017**: Una rutina o sesión PUEDE crearse con un conjunto vacío, porque no se indicó una
  cantidad mínima de elementos.
- **FR-018**: Todo nombre y toda descripción presente DEBEN carecer de espacios al inicio o al final
  y DEBEN usar exactamente un espacio simple entre palabras, sin tabulaciones ni saltos de línea.
- **FR-019**: Los textos que incumplan la regla de espacios DEBEN rechazarse con indicaciones que
  permitan corregirlos; el sistema NO DEBE modificarlos silenciosamente.
- **FR-020**: Las URLs de imagen y video, cuando se informen, DEBEN ser URLs absolutas válidas con
  esquema HTTP o HTTPS y NO DEBEN contener espacios.
- **FR-021**: El sistema DEBE informar todos los campos y asociaciones inválidos detectables en un
  intento de creación o edición.
- **FR-022**: El apartado de rutinas DEBE mostrar las rutinas propias y permitir abrir el detalle
  completo de cada una, incluyendo descripción, días, sesiones, ejercicios, series, repeticiones y
  orden de ejecución, además de las URLs opcionales.
- **FR-023**: El sistema DEBE ofrecer también listados y detalles propios de sesiones y ejercicios
  para que puedan seleccionarse, consultarse y eliminarse.
- **FR-024**: Eliminar una rutina DEBE eliminar sus asignaciones de sesiones, pero NO las sesiones
  ni ejercicios reutilizables.
- **FR-025**: Al eliminar una sesión actualmente utilizada, el sistema DEBE retirar sus
  asociaciones de todas las rutinas propias y conservar esas rutinas sin la sesión eliminada.
- **FR-026**: Al eliminar un ejercicio actualmente utilizado, el sistema DEBE retirar sus
  asociaciones de todas las sesiones propias, conservar esas sesiones sin el ejercicio eliminado y
  renumerar sus ejercicios restantes consecutivamente desde 1, respetando su orden relativo previo.
- **FR-027**: La eliminación de una sesión o ejercicio y el retiro de todas sus asociaciones DEBEN
  completarse como una única operación; ante cualquier fallo, no se debe aplicar un resultado
  parcial.
- **FR-028**: La interfaz DEBE solicitar confirmación antes de eliminar una entidad propia.
- **FR-029**: Los nombres NO DEBEN ser únicos; una misma cuenta y cuentas distintas PUEDEN tener
  entidades con nombres iguales.
- **FR-030**: Una entidad seleccionada DEBE seguir perteneciendo al usuario al momento de confirmar
  la creación o edición, aunque el catálogo mostrado se haya cargado anteriormente.
- **FR-031**: Una misma sesión PUEDE aparecer varias veces en una rutina solamente cuando cada
  aparición tenga un día diferente; la combinación de rutina, sesión y día NO DEBE repetirse.
- **FR-032**: La persona autenticada DEBE poder editar únicamente ejercicios, sesiones y rutinas
  propios, conservando el identificador y propietario originales.
- **FR-033**: Editar un ejercicio DEBE reemplazar su nombre, descripción y URLs opcionales usando
  las mismas validaciones aplicadas durante la creación.
- **FR-034**: Editar una sesión DEBE reemplazar su nombre, descripción y conjunto completo de
  ejercicios, incluyendo series, repeticiones y orden, como una única operación atómica.
- **FR-035**: Editar una rutina DEBE reemplazar su nombre, descripción y conjunto completo de
  asignaciones sesión/día como una única operación atómica.
- **FR-036**: Una lista vacía durante la edición DEBE retirar todas las asociaciones y conservar la
  sesión o rutina; los campos opcionales omitidos o vacíos DEBEN eliminar su valor anterior.
- **FR-037**: La interfaz DEBE precargar la representación actual al iniciar una edición, permitir
  guardar o cancelar y conservar los valores ingresados cuando la actualización sea rechazada.
- **FR-038**: Cancelar o abandonar una edición con cambios sin guardar DEBE solicitar confirmación;
  al confirmar, NO DEBE enviar una actualización ni modificar datos persistidos, y al rechazar DEBE
  conservar el formulario y permanecer en la edición.
- **FR-039**: Una actualización concurrente DEBE producir una representación completa confirmada,
  nunca una mezcla parcial; no se requiere detectar ni resolver conflictos entre escritores.
- **FR-040**: Las asociaciones DEBEN conservar referencias reutilizables: editar un ejercicio o
  sesión DEBE reflejar sus datos actualizados en todos los detalles de sesiones o rutinas que lo
  incluyan, sin crear copias históricas ni bloquear la edición por estar en uso.
- **FR-041**: Una actualización DEBE aplicar esta prioridad de errores: estructura y reglas
  detectables sin persistencia, existencia y pertenencia del recurso objetivo, y disponibilidad de
  referencias seleccionadas. Un objetivo ajeno o inexistente DEBE informarse como recurso no
  encontrado antes de informar referencias internas no disponibles.
- **FR-042**: Cada consulta detallada de sesión o rutina DEBE representar un único estado confirmado;
  NO DEBE combinar campos o asociaciones pertenecientes a estados anteriores y posteriores de una
  edición concurrente.
- **FR-043**: Un identificador no numérico, no positivo o fuera del rango entero admitido DEBE
  producir un resultado de solicitud inválida; un identificador positivo válido ajeno o inexistente
  DEBE producir el mismo resultado de recurso no encontrado sin revelar cuál caso ocurrió.
- **FR-044**: La vista no autenticada DEBE carecer de navbar y usar la imagen de gimnasio indicada
  por el usuario como fondo visual, con fallback oscuro y un degradado oscuro al lado izquierdo que
  mantenga legible la card de autenticación.
- **FR-045**: Login DEBE mostrarse inicialmente dentro de una card e incluir el enlace “¿No tienes
  cuenta? Créate una aquí”; activarlo DEBE sustituir la card por registro sin navegar a otra página.
- **FR-046**: Registro DEBE conservar todos los campos, confirmación de contraseña y controles para
  mostrar u ocultar cada contraseña; también DEBE permitir volver al login.
- **FR-047**: El workspace autenticado DEBE mostrar una navbar con “FitPro” e icono de mancuerna a
  la izquierda y controles centrados para “Rutinas”, “Sesiones” y “Ejercicios”.
- **FR-048**: Cada apartado autenticado DEBE mostrar un hero con la imagen indicada, título del
  apartado y una descripción breve específica de su función.
- **FR-049**: Crear y editar rutinas, sesiones y ejercicios DEBE realizarse dentro de un mismo
  componente wizard reutilizable, presentado como diálogo modal con título y acciones accesibles.
- **FR-050**: Los pasos del wizard DEBEN mostrarse como tabs con borde inferior coloreado, texto y
  estado accesible; la persona DEBE poder visitar cualquier paso aunque existan campos incompletos.
- **FR-051**: Un click sobre el backdrop del wizard DEBE solicitar confirmación mediante una leyenda
  equivalente a “Los cambios se perderán, ¿seguro deseas salir?”. Rechazar DEBE conservar modal,
  paso y draft; aceptar DEBE cerrar sin POST ni PUT. Un click dentro NO DEBE activar este flujo.
- **FR-052**: Guardar desde el wizard NO DEBE solicitar confirmación, DEBE impedir envíos duplicados
  mientras esté pendiente y, ante rechazo, DEBE conservar el draft y mostrar los errores en el paso
  correspondiente. La acción Guardar DEBE mostrarse solamente en el paso final “Resumen”; los pasos
  anteriores DEBEN ofrecer navegación sin realizar escrituras.
- **FR-053**: El wizard de rutina DEBE contener “Datos básicos”, “Sesiones” y “Resumen”. El segundo
  paso DEBE ofrecer búsqueda y selección múltiple local. Una sesión DEBE permanecer seleccionable
  repetidamente desde el buscador y cada selección DEBE crear una fila independiente con su propio
  selector de día. La combinación sesión/día NO DEBE duplicarse.
- **FR-054**: El wizard de sesión DEBE contener “Datos básicos”, “Ejercicios” y “Resumen”. El segundo
  paso DEBE ofrecer búsqueda y selección múltiple local, una lista ordenable y spinboxes no
  negativos para series y repeticiones; el orden DEBE derivarse de la posición.
- **FR-055**: El wizard de ejercicio DEBE contener “Datos básicos” y “Resumen”. El primer paso DEBE
  incluir nombre, descripción y URLs opcionales de imagen y video.
- **FR-056**: La vista de rutinas DEBE mostrar cards cuyo contenido principal abre el detalle modal,
  con nombre, descripción y acciones de edición/eliminación separadas abajo a la derecha. El modal
  DEBE contener sesiones colapsables por día y ejercicios internos colapsables con
  miniatura, nombre, series, repeticiones, descripción y preview de video cuando exista.
- **FR-057**: La vista de sesiones DEBE mostrar cards colapsables con nombre, descripción, cantidad
  de ejercicios y acciones; al expandirlas DEBE listar ejercicios con nombre, series y repeticiones.
- **FR-058**: La vista de ejercicios DEBE mostrar cards con imagen amplia y nombre; al expandirlas
  DEBE mostrar descripción y preview de video, además de acciones de edición y eliminación.
- **FR-059**: La interfaz DEBE usar iconos de `lucide-react` para marca, navegación y acciones
  principales. Los SVG DEBEN heredar `currentColor`; los decorativos DEBEN ocultarse de tecnologías
  de asistencia y todo botón solo-icono DEBE poseer nombre accesible contextual.
- **FR-060**: La interfaz DEBE mantener operación por teclado, foco visible, retorno de foco al
  cerrar diálogos, nombres accesibles, estados no comunicados solo por color y layout utilizable a
  320 px y zoom de 200%.
- **FR-061**: Una URL de video DEBE intentar reproducirse únicamente mediante el reproductor nativo
  del navegador, sin `iframe` ni integración específica con plataformas. Si el recurso no es un
  video directo reproducible, la interfaz DEBE conservar un enlace externo seguro y descriptivo.
- **FR-062**: Después de un POST o PUT exitoso, la interfaz DEBE aplicar la representación devuelta
  a la card correspondiente, cerrar el wizard, devolver el foco a un punto lógico del apartado y
  anunciar el éxito. NO DEBE abrir automáticamente otro modal ni realizar un GET adicional.
- **FR-063**: Los colapsables de rutinas, sesiones y ejercicios DEBEN funcionar de forma
  independiente. Abrir uno NO DEBE cerrar otros elementos del mismo nivel ni elementos anidados que
  permanezcan montados.

### Key Entities *(include if feature involves data)*

- **Ejercicio**: Movimiento perteneciente a una cuenta. Tiene nombre, descripción opcional y URLs
  opcionales de imagen y video. Puede ser reutilizado en varias sesiones propias.
- **Sesión**: Agrupación perteneciente a una cuenta. Tiene nombre, descripción opcional y ejercicios
  seleccionados del mismo usuario.
- **Ejercicio de sesión**: Asociación entre una sesión y un ejercicio propio. Conserva series y
  repeticiones específicas para esa sesión y su orden de ejecución dentro de ella.
- **Rutina**: Plan perteneciente a una cuenta. Tiene nombre, descripción opcional y sesiones
  seleccionadas del mismo usuario.
- **Sesión de rutina**: Asociación entre una rutina y una sesión propia. Conserva el día semanal
  específico para esa aparición en la rutina. Una sesión puede tener varias asociaciones dentro de
  la misma rutina si sus días son distintos.

### Testable Behaviors *(mandatory)*

- **TB-001 Backend**: Crea y consulta un ejercicio propio con campos opcionales presentes o
  ausentes.
- **TB-002 Backend**: Rechaza textos con espacios exteriores, separaciones múltiples, tabulaciones
  o saltos de línea.
- **TB-003 Backend**: Acepta URLs HTTP/HTTPS válidas y rechaza URLs relativas, con espacios o con
  esquemas no admitidos.
- **TB-004 Backend**: Crea una sesión atómicamente con ejercicios propios, cantidades no negativas
  y órdenes únicos consecutivos desde 1.
- **TB-005 Backend**: Rechaza la repetición de un ejercicio dentro de la misma sesión sin impedir
  que se reutilice en otras sesiones.
- **TB-006 Backend**: Rechaza series o repeticiones negativas, decimales o fuera del rango entero.
- **TB-007 Backend**: Crea una rutina atómicamente con sesiones propias y días entre 1 y 7.
- **TB-008 Backend**: Conserva valores independientes al reutilizar ejercicios y sesiones.
- **TB-009 Backend**: Impide acceso cruzado en listados, detalles, selecciones y eliminaciones sin
  distinguir públicamente entidades ajenas de inexistentes.
- **TB-010 Backend**: Elimina una rutina sin eliminar sesiones ni ejercicios reutilizados.
- **TB-011 Backend**: Elimina una sesión o ejercicio en uso junto con todas sus asociaciones, de
  forma atómica, conserva las rutinas o sesiones contenedoras y renumera desde 1 los ejercicios
  restantes de cada sesión afectada sin alterar su orden relativo.
- **TB-012 Backend**: Permite reutilizar una sesión en días diferentes de una rutina y rechaza una
  combinación repetida de rutina, sesión y día.
- **TB-013 Frontend**: Permite crear ejercicios y muestra errores de texto o URL junto a sus campos.
- **TB-014 Frontend**: Permite crear sesiones seleccionando solamente ejercicios propios y
  capturando series y repeticiones por selección, reordenar la lista y visualizar el orden `1..N`
  asignado automáticamente.
- **TB-015 Frontend**: Permite crear rutinas seleccionando solamente sesiones propias, asignando un
  día válido por aparición y reutilizando una sesión únicamente en días distintos.
- **TB-016 Frontend**: Lista rutinas propias y muestra su detalle completo, incluyendo estados
  vacíos y campos opcionales ausentes.
- **TB-017 Frontend**: Solicita confirmación de eliminación, representa éxito o rechazo y no muestra
  acciones de entidades ajenas.
- **TB-018 Frontend**: Ante autenticación ausente o vencida, no muestra información privada y vuelve
  al flujo no autenticado existente.
- **TB-019 Backend**: Edita un ejercicio propio, permite limpiar opcionales y mantiene el mismo
  identificador; una entidad ajena o inexistente produce el mismo resultado público.
- **TB-020 Backend**: Reemplaza atómicamente la composición completa de una sesión, admite vaciarla
  y conserva el estado anterior cuando una referencia o asociación es inválida.
- **TB-021 Backend**: Reemplaza atómicamente las asignaciones completas de una rutina, admite
  vaciarla y conserva el estado anterior ante cualquier fallo.
- **TB-022 Backend**: Dos reemplazos concurrentes dejan uno de los estados completos y no mezclan
  campos o asociaciones.
- **TB-023 Frontend**: Precarga y guarda la edición de ejercicios, actualizando listado y detalle
  con la respuesta completa del servidor.
- **TB-024 Frontend**: Precarga sesiones y rutinas, permite agregar, quitar y reordenar asociaciones,
  y envía la representación completa.
- **TB-025 Frontend**: Cancelar o abandonar una edición modificada solicita confirmación; aceptarla
  no realiza una actualización y restaura el modo de creación, mientras rechazarla conserva el
  formulario y apartado actuales.
- **TB-026 Frontend**: Un rechazo conserva los valores editados, muestra errores accesibles y evita
  envíos duplicados; un `401` vuelve al flujo no autenticado.
- **TB-027 Backend**: Editar un ejercicio o sesión reutilizada actualiza los detalles anidados de
  todos sus contenedores sin cambiar los valores propios de cada asociación.
- **TB-028 Backend**: Una actualización con objetivo ausente o ajeno y referencias no disponibles
  informa primero que el recurso no fue encontrado y no revela el estado de esas referencias.
- **TB-029 Backend**: Una consulta detallada concurrente con una actualización devuelve íntegramente
  el estado confirmado anterior o el posterior, nunca una combinación de ambos.
- **TB-030 Backend**: Consulta, edición y eliminación distinguen identificadores inválidos como
  solicitudes inválidas, mientras identificadores válidos ajenos o inexistentes comparten el mismo
  resultado de recurso no encontrado.
- **TB-031 Frontend**: Login y registro alternan dentro de la misma card, permanecen operativos si
  la imagen falla y la vista no autenticada no presenta navbar.
- **TB-032 Frontend**: Navbar FitPro y héroes permiten cambiar de apartado por mouse y teclado sin
  perder contenido privado ni romper el guard de drafts.
- **TB-033 Frontend**: El wizard permite salto directo, navegación anterior/siguiente y regreso entre
  pasos incompletos sin borrar valores ni validar como puerta.
- **TB-034 Frontend**: Click en backdrop solicita confirmación; rechazar preserva modal, paso y draft,
  aceptar cierra sin request, y click interno no cierra.
- **TB-035 Frontend**: Guardar no solicita confirmación, evita doble submit y dirige los errores al
  primer paso afectado conservando el draft; la acción no aparece fuera de “Resumen”.
- **TB-036 Frontend**: El multiselect filtra localmente, conserva selecciones ocultas y comunica
  resultados vacíos; el buscador de rutinas permite elegir repetidamente una sesión para días
  distintos sin admitir el mismo par, y el selector de sesiones impide ejercicios repetidos.
- **TB-037 Frontend**: Cards y modal de rutina muestran la jerarquía colapsable completa, incluidos
  estados vacíos y fallbacks de medios.
- **TB-038 Frontend**: Cards de sesiones y ejercicios expanden información progresiva y conservan
  acciones accesibles separadas.
- **TB-039 Frontend**: Los iconos Lucide no alteran nombres accesibles, foco, semántica de acciones
  ni contraste del tema.
- **TB-040 Frontend**: Auth, navbar, héroes, cards, diálogos y controles dinámicos superan axe y
  verificación manual a 320/768/1280 px y zoom de 200%.
- **TB-041 Frontend**: Una URL directa reproducible usa controles nativos sin autoplay; una URL no
  reproducible muestra un enlace externo y nunca crea un `iframe`.
- **TB-042 Frontend**: Un guardado exitoso cierra el wizard, actualiza la card usando la respuesta,
  anuncia el resultado y no abre el detalle ni realiza una consulta adicional.
- **TB-043 Frontend**: Dos o más colapsables del mismo nivel y de niveles anidados pueden permanecer
  abiertos simultáneamente y operar por teclado.

Esta feature define 20 comportamientos útiles de backend y 23 de frontend, además de reutilizar las
pruebas de autenticación existentes.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Una persona puede crear un ejercicio válido en menos de 1 minuto sin asistencia
  técnica.
- **SC-002**: Una persona con ejercicios existentes puede formar una sesión válida en menos de 2
  minutos y una rutina válida en menos de 3 minutos.
- **SC-003**: El 100% de los intentos de acceso cruzado probados son rechazados sin revelar datos ni
  modificar contenido de otra cuenta.
- **SC-004**: El 100% de los textos, URLs, días y cantidades inválidos definidos por esta
  especificación son rechazados sin crear entidades parciales.
- **SC-005**: El 100% de las rutinas creadas pueden consultarse mostrando correctamente todas sus
  sesiones, ejercicios y valores asociados.
- **SC-006**: Al menos 90% de las personas de una prueba guiada puede localizar una rutina propia y
  comprender su planificación completa sin asistencia.
- **SC-007**: El 100% de las eliminaciones confirmadas respetan pertenencia, retiran asociaciones
  dependientes y renumeran las sesiones afectadas sin alterar contenido ajeno.
- **SC-008**: Una sesión o ejercicio reutilizado conserva correctamente sus valores particulares en
  el 100% de las rutinas o sesiones donde aparece.
- **SC-009**: Una persona puede editar un ejercicio en menos de 1 minuto, una sesión en menos de 2
  minutos y una rutina en menos de 3 minutos, partiendo de sus valores actuales.
- **SC-010**: El 100% de las actualizaciones exitosas conserva identificador y propietario y muestra
  el detalle completo actualizado.
- **SC-011**: El 100% de las actualizaciones rechazadas conserva íntegramente el estado persistido
  anterior, sin campos ni asociaciones parciales.
- **SC-012**: El 100% de las cancelaciones probadas termina sin solicitudes de actualización ni
  cambios persistidos.
- **SC-013**: El 100% de las consultas detalladas probadas bajo edición concurrente representa un
  único estado confirmado sin mezclar campos o asociaciones de estados distintos.
- **SC-014**: El 100% de los wizards probados permite recorrer libremente todos sus pasos y conserva
  el draft al regresar, aun cuando existan campos incompletos.
- **SC-015**: El 100% de los clicks de backdrop probados solicita confirmación y ningún descarte
  aceptado ni guardado cancelado genera una escritura accidental.
- **SC-016**: El 100% de los detalles probados muestra la jerarquía solicitada y mantiene una
  alternativa visible cuando una imagen o video no puede reproducirse.
- **SC-017**: Login, registro, navegación, cards y wizards permanecen utilizables sin scroll
  horizontal a 320 px y con zoom de 200% en los escenarios guiados.

## Assumptions

- El token y la identidad autenticada existentes se reutilizan; esta feature no cambia login ni
  gestión de sesión.
- Una sesión o rutina puede comenzar vacía porque no se indicó una cantidad mínima de selecciones.
- Una sesión puede repetirse dentro de una rutina en días distintos, pero una misma combinación de
  sesión y día no puede duplicarse.
- Los nombres pueden repetirse porque no se solicitó unicidad.
- Las descripciones son textos opcionales de una sola línea; una cadena vacía equivale a ausencia.
- Las URLs opcionales siguen siendo referencias externas. El navegador puede cargar imágenes y
  videos directos de forma progresiva mediante elementos nativos, pero backend no descarga, valida
  contenido, aloja, transforma ni actúa como proxy. YouTube, Vimeo, páginas HTML y formatos no
  reproducibles no se integran: conservan un enlace externo seguro.
- La numeración semanal sigue la convención ISO: lunes es 1 y domingo es 7.
- Eliminar una sesión o ejercicio en uso retira automáticamente sus asociaciones y conserva los
  contenedores, que pueden quedar vacíos.
- Después de eliminar un ejercicio, cada sesión afectada conserva el orden relativo de los
  ejercicios restantes y vuelve a numerarlos consecutivamente desde 1.
- Los límites técnicos de longitud y rango se definirán durante la planificación sin agregar reglas
  de negocio innecesarias.
- El orden de ejercicios es explícito dentro de cada sesión; el orden de sesiones dentro de una
  rutina se determina por su día asignado.
- La edición reemplaza la representación editable completa; no admite actualización parcial ni
  edición anidada de ejercicios dentro de sesiones o de sesiones dentro de rutinas.
- Las listas `exercises` y `sessions` son obligatorias en una actualización y pueden enviarse
  vacías. Los campos opcionales omitidos o vacíos equivalen a ausencia.
- Las ediciones concurrentes usan una política simple de última transacción confirmada; no se
  incorporan versiones, ETags, historial ni resolución de conflictos.
- La imagen de gimnasio suministrada se incorporará como asset local optimizado para evitar una
  dependencia de hotlink en ejecución; seguirá existiendo un fallback oscuro si el asset falla.
- La búsqueda de los multiselect opera en memoria sobre catálogos ya cargados y no agrega búsqueda,
  filtrado ni paginación al backend.
- El backdrop del wizard siempre solicita confirmación; otros abandonos conservan el guard de
  cambios existente. Mientras una escritura está pendiente, el wizard no puede cerrarse.
