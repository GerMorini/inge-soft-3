# Feature Specification: Registro e inicio de sesión

**Feature Branch**: `feature/login-register`

**Created**: 2026-08-13

**Status**: Draft

**Input**: User description: "Añadir registro con datos personales e inicio de sesión mediante
nombre de usuario y contraseña. El login entrega un JWT con 30 minutos de vigencia, almacenado por
el frontend y validado por middleware del backend, sin persistir sesiones en base de datos. No
incluir recuperación, validación de email, MFA, roles ni permisos. Durante el registro, solicitar
la contraseña dos veces y ofrecer un control para verla u ocultarla en ambos campos."

## Clarifications

### Session 2026-08-13

- Q: ¿Qué caracteres y longitud admite el username? → A: Entre 3 y 30 caracteres: letras ASCII,
  números, punto, guion y guion bajo.
- Q: ¿Cómo deben limitarse los intentos fallidos de login? → A: Sin límites; queda fuera del
  alcance.
- Q: ¿Cómo se ingresan ciudad y provincia? → A: Ambos como texto libre obligatorio.
- Q: ¿Cómo se manejan espacios exteriores? → A: Recortar nombre, apellido, email y dirección;
  validar literalmente username, teléfono y contraseña.
- Q: ¿Qué formato admite el número domiciliario? → A: Texto obligatorio de hasta 20 caracteres.
- Q: ¿Cómo se mantiene la sesión autenticada? → A: Mediante un JWT con 30 minutos de vigencia,
  almacenado en el frontend y validado por middleware, sin persistencia de sesión en base de datos.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Registrar una cuenta (Priority: P1)

Una persona sin cuenta ingresa sus datos personales, dirección, nombre de usuario, email,
contraseña y confirmación de contraseña para crear su cuenta y poder identificarse posteriormente.

**Why this priority**: El registro crea la identidad necesaria para cualquier inicio de sesión.

**Independent Test**: Se completa el formulario con dos contraseñas coincidentes, se comprueba que
cada campo permite mostrar u ocultar su valor y se verifica que la cuenta queda disponible para un
inicio de sesión posterior.

**Acceptance Scenarios**:

1. **Given** que no existen cuentas con el username ni el email ingresados, **When** la persona
   completa todos los datos obligatorios con formatos válidos, **Then** el sistema crea una única
   cuenta y confirma el registro.
2. **Given** una dirección sin departamento, **When** los demás datos obligatorios son válidos,
   **Then** el sistema crea la cuenta sin exigir ese dato opcional.
3. **Given** los campos de contraseña y confirmación completos, **When** la persona usa el botón
   `Ver contraseña` de cualquiera de ellos, **Then** puede ver el valor de ese campo y volver a
   ocultarlo sin modificarlo.
4. **Given** contraseñas distintas en los dos campos, **When** la persona intenta registrarse,
   **Then** el sistema no envía ni crea la cuenta e indica que ambas contraseñas deben coincidir.

---

### User Story 2 - Iniciar sesión (Priority: P2)

Una persona registrada ingresa su username y contraseña para demostrar que conoce las credenciales
de su cuenta.

**Why this priority**: Permite comprobar la identidad registrada, pero depende de que exista una
cuenta previa.

**Independent Test**: Con una cuenta existente preparada para la prueba, se ingresan sus
credenciales correctas y se verifica que el sistema entrega un token con 30 minutos de vigencia sin
crear registros de sesión en base de datos.

**Acceptance Scenarios**:

1. **Given** una cuenta registrada, **When** la persona ingresa su username y contraseña correctos,
   **Then** el sistema entrega un token de acceso válido durante 30 minutos y el frontend lo
   conserva para solicitudes autenticadas.
2. **Given** una cuenta registrada, **When** la persona ingresa una contraseña incorrecta, **Then**
   el sistema rechaza el intento con un mensaje genérico que no revela cuál credencial falló.
3. **Given** un username inexistente, **When** la persona intenta iniciar sesión, **Then** el sistema
   responde con el mismo rechazo genérico usado para una contraseña incorrecta.

---

### User Story 3 - Mantener acceso autenticado temporal (Priority: P3)

Una persona que inició sesión puede acceder a su identidad actual mientras su token siga vigente,
sin que el backend mantenga una sesión persistida.

**Why this priority**: Hace utilizable el resultado del login y demuestra la validación stateless
requerida, sin incorporar roles ni permisos.

**Independent Test**: Se obtiene un token mediante login, se consulta la identidad actual con ese
token y luego se comprueba que un token vencido o alterado sea rechazado.

**Acceptance Scenarios**:

1. **Given** un token válido emitido por el sistema, **When** el frontend solicita la identidad
   actual, **Then** el middleware permite la solicitud y el sistema devuelve el identificador y
   username autenticados.
2. **Given** un token vencido, alterado o ausente, **When** se solicita la identidad actual,
   **Then** el middleware rechaza la solicitud con una respuesta genérica de autenticación.
3. **Given** un token almacenado en el frontend, **When** alcanza sus 30 minutos de vigencia o el
   backend lo rechaza, **Then** el frontend elimina el token y vuelve al estado no autenticado.

---

### User Story 4 - Corregir datos inválidos (Priority: P4)

Una persona que intenta registrarse recibe indicaciones específicas sobre los campos inválidos para
poder corregirlos sin perder los demás datos ingresados, excepto la contraseña y su confirmación.

**Why this priority**: La retroalimentación permite completar el flujo principal y hace visibles las
reglas de registro.

**Independent Test**: Se envía cada combinación inválida definida y se verifica que el sistema
rechaza el registro, identifica los campos que deben corregirse y no crea una cuenta parcial.

**Acceptance Scenarios**:

1. **Given** una cuenta con el mismo email o username, **When** otra persona intenta registrarse con
   el dato duplicado, **Then** el sistema rechaza el registro e identifica el conflicto.
2. **Given** un username con espacios, **When** la persona intenta registrarse, **Then** el sistema
   rechaza el registro e indica que el username no admite espacios.
3. **Given** una contraseña que incumple una o más reglas, **When** la persona intenta registrarse,
   **Then** el sistema no crea la cuenta e informa cada regla incumplida.
4. **Given** un teléfono fuera del formato E.164, **When** la persona intenta registrarse, **Then**
   el sistema no crea la cuenta e identifica el formato requerido.

### Edge Cases

- El email o username solo difiere en mayúsculas y minúsculas respecto de una cuenta existente.
- Dos registros concurrentes intentan utilizar el mismo email o username.
- La contraseña tiene exactamente 8 caracteres y cumple los conteos mínimos.
- La contraseña repite el mismo número o símbolo hasta alcanzar los conteos requeridos.
- La confirmación difiere de la contraseña por mayúsculas, espacios o cualquier otro carácter.
- Se alterna varias veces la visibilidad de uno o ambos campos de contraseña antes de enviar.
- El username contiene espacios al inicio, al final o entre caracteres.
- El username tiene menos de 3 o más de 30 caracteres, o contiene caracteres no admitidos.
- El teléfono incluye separadores, espacios, prefijo internacional `00` o más de 15 dígitos.
- Falta uno o más campos obligatorios de datos personales o dirección.
- Nombre, apellido, email o dirección contienen espacios exteriores o consisten solo en espacios.
- El número domiciliario contiene 20 caracteres o utiliza valores como `123 Bis` o `S/N`.
- El inicio de sesión usa la contraseña correcta con un username que cambia mayúsculas o minúsculas.
- Una persona realiza múltiples intentos fallidos y luego ingresa las credenciales correctas.
- Un registro rechazado no deja una cuenta parcial ni reserva su email o username.
- Un token se usa inmediatamente antes y exactamente al alcanzar su instante de expiración.
- Un token presenta firma inválida, formato incorrecto o un algoritmo distinto del admitido.
- Una solicitud autenticada presenta incorrectamente el token o no lo incluye.
- La página se recarga mientras el token almacenado todavía es válido o ya venció.

## Scope Boundaries *(mandatory)*

- **In scope**: Registro de una cuenta con los datos indicados, confirmación de contraseña,
  controles para mostrar u ocultar ambos campos de contraseña, validación de sus reglas, detección
  de email y username duplicados, inicio de sesión mediante username y contraseña, emisión y
  almacenamiento frontend de un JWT, y validación del token para consultar la identidad actual.
- **Out of scope**: Recuperación o cambio de contraseña, persistencia de sesiones o tokens,
  renovación y revocación anticipada de tokens, validación o confirmación de email, autenticación
  multifactor, roles, permisos, acceso mediante email, edición o eliminación de perfiles, y
  limitación o bloqueo por intentos fallidos.
- **Simplicity rationale**: La feature mantiene una sesión temporal y stateless con un solo token.
  No agrega tablas, refresh tokens, listas de revocación ni capacidades de autorización.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: El sistema DEBE permitir registrar una cuenta con nombre, apellido, teléfono,
  dirección, username, email y contraseña.
- **FR-002**: La dirección DEBE incluir calle, número, ciudad y provincia; ciudad y provincia DEBEN
  ser textos libres obligatorios, el número DEBE ser texto obligatorio de hasta 20 caracteres y el
  departamento DEBE ser opcional.
- **FR-003**: Todos los campos obligatorios DEBEN contener un valor no vacío.
- **FR-004**: El teléfono DEBE cumplir el formato E.164: signo `+`, código de país y número, con un
  máximo total de 15 dígitos y sin espacios ni separadores.
- **FR-005**: El email DEBE tener un formato válido y ser único sin distinguir mayúsculas de
  minúsculas.
- **FR-006**: El username DEBE ser único sin distinguir mayúsculas de minúsculas.
- **FR-007**: El username DEBE contener entre 3 y 30 caracteres y admitir únicamente letras ASCII,
  números, punto, guion y guion bajo; NO DEBE contener espacios.
- **FR-008**: La contraseña DEBE contener al menos 8 caracteres, una letra mayúscula, tres dígitos y
  dos símbolos especiales.
- **FR-009**: Para contar símbolos especiales, el sistema DEBE considerar cualquier carácter que no
  sea una letra, un dígito ni un espacio. Cada aparición DEBE contar, aunque repita carácter.
- **FR-010**: El sistema DEBE rechazar el registro completo cuando cualquier dato obligatorio sea
  inválido o cuando email o username ya existan.
- **FR-011**: Un registro rechazado NO DEBE crear datos parciales ni impedir que sus valores se usen
  en un intento válido posterior.
- **FR-012**: El sistema DEBE informar los campos inválidos y las reglas incumplidas durante el
  registro, sin volver a mostrar la contraseña ni su confirmación después de un intento enviado.
- **FR-013**: El sistema DEBE permitir iniciar sesión exclusivamente con username y contraseña.
- **FR-014**: La comparación del username al iniciar sesión NO DEBE distinguir mayúsculas de
  minúsculas; la comparación de la contraseña DEBE distinguirlas.
- **FR-015**: El sistema DEBE entregar un JWT firmado cuando ambas credenciales correspondan a una
  cuenta registrada.
- **FR-016**: El sistema DEBE rechazar username inexistente y contraseña incorrecta con el mismo
  mensaje genérico, sin revelar si una cuenta existe.
- **FR-017**: El JWT DEBE expirar 30 minutos después de su emisión y DEBE considerarse inválido al
  alcanzar o superar ese instante.
- **FR-018**: El sistema NO DEBE exponer ni devolver la contraseña o su confirmación después de
  recibirlas.
- **FR-019**: La creación de cuentas DEBE preservar la unicidad de email y username incluso ante
  intentos concurrentes.
- **FR-020**: Los intentos fallidos NO DEBEN bloquear la cuenta ni introducir esperas; la limitación
  de intentos queda fuera del alcance de esta feature.
- **FR-021**: El sistema DEBE quitar espacios exteriores de nombre, apellido, email y todos los
  campos de dirección antes de validarlos y guardarlos. Username, teléfono y contraseña DEBEN
  validarse exactamente como fueron ingresados.
- **FR-022**: El frontend DEBE almacenar el JWT y enviarlo al solicitar recursos protegidos.
- **FR-023**: Un middleware del backend DEBE validar la firma, el algoritmo admitido y la expiración
  del JWT antes de permitir acceso a un recurso protegido.
- **FR-024**: Un token ausente, malformado, alterado o vencido DEBE recibir el mismo rechazo genérico
  y NO DEBE ejecutar la operación protegida.
- **FR-025**: El sistema DEBE permitir que una persona autenticada consulte su identificador y
  username actuales mediante un recurso protegido.
- **FR-026**: El frontend DEBE eliminar el token almacenado cuando esté vencido o cuando el backend
  lo rechace, y DEBE representar nuevamente el estado no autenticado.
- **FR-027**: El backend NO DEBE persistir sesiones ni tokens, y NO DEBE implementar refresh tokens,
  revocación anticipada o listas de bloqueo.
- **FR-028**: El sistema NO DEBE exponer el JWT en logs ni incluir credenciales o datos personales
  innecesarios dentro de sus claims.
- **FR-029**: El formulario de registro DEBE solicitar la contraseña dos veces mediante campos
  obligatorios de contraseña y confirmación de contraseña.
- **FR-030**: El formulario de registro DEBE impedir el envío cuando contraseña y confirmación no
  coincidan exactamente, y DEBE indicar el desacuerdo junto al campo de confirmación.
- **FR-031**: Cada campo de contraseña del registro DEBE tener su propio botón `Ver contraseña`.
  Al activarlo, DEBE mostrar el valor de ese campo, cambiar a una acción `Ocultar contraseña` y
  permitir ocultarlo nuevamente sin cambiar el valor ni afectar la visibilidad del otro campo.

### Key Entities *(include if feature involves data)*

- **Cuenta de usuario**: Representa la identidad registrada. Incluye nombre, apellido, teléfono,
  username único, email único, credencial secreta y una dirección de residencia.
- **Dirección**: Representa el domicilio de la cuenta. Incluye calle, número, departamento opcional,
  ciudad y provincia ingresadas como texto libre. El número admite hasta 20 caracteres.
- **Token de acceso**: Credencial firmada y temporal que representa la identidad autenticada.
  Contiene solamente identificador de usuario, username y tiempos necesarios; no se persiste.

### Testable Behaviors *(mandatory)*

- **TB-001 Backend**: Normaliza los campos definidos y crea una cuenta cuando todos cumplen las
  reglas.
- **TB-002 Backend**: Acepta una dirección sin departamento.
- **TB-003 Backend**: Rechaza teléfonos que no cumplen E.164.
- **TB-004 Backend**: Rechaza email duplicado sin distinguir mayúsculas.
- **TB-005 Backend**: Rechaza username duplicado sin distinguir mayúsculas.
- **TB-006 Backend**: Rechaza username fuera de longitud o con caracteres no admitidos.
- **TB-007 Backend**: Evalúa longitud, mayúscula, tres dígitos y dos símbolos de la contraseña.
- **TB-008 Backend**: Evita cuentas parciales ante cualquier rechazo.
- **TB-009 Backend**: Acepta credenciales correctas incluso después de fallos previos y rechaza las
  incorrectas sin revelar cuál falló.
- **TB-010 Backend**: Conserva unicidad ante dos registros concurrentes equivalentes.
- **TB-011 Backend**: Emite un JWT cuya expiración queda exactamente 30 minutos después de su
  emisión y cuyos claims no contienen contraseña, email, teléfono ni dirección.
- **TB-012 Backend**: El middleware permite un token válido y expone al request solamente la
  identidad verificada.
- **TB-013 Backend**: El middleware rechaza de forma uniforme tokens ausentes, malformados,
  alterados, vencidos o con algoritmo no admitido.
- **TB-014 Backend**: Consultar la identidad actual requiere un token válido y no crea datos de
  sesión en PostgreSQL.
- **TB-015 Frontend**: Presenta todos los campos y marca claramente los obligatorios.
- **TB-016 Frontend**: Muestra las reglas incumplidas sin borrar datos corregibles ni reponer la
  contraseña o su confirmación.
- **TB-017 Frontend**: Permite enviar username y contraseña, almacena el JWT recibido y representa
  estados de espera, éxito y rechazo explícitos.
- **TB-018 Frontend**: Restaura un token todavía vigente y elimina uno vencido o rechazado sin
  ofrecer recuperación, MFA, roles ni permisos.
- **TB-019 Frontend**: Exige una confirmación idéntica antes de enviar el registro y permite mostrar
  u ocultar independientemente ambos valores sin modificarlos.

Esta feature aporta 14 comportamientos útiles de backend y 5 de frontend al mínimo de pruebas del
proyecto.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Una persona puede completar un registro válido en menos de 3 minutos durante una
  prueba guiada sin asistencia técnica.
- **SC-002**: El 100% de los registros que violan las reglas de email, username, teléfono o
  contraseña son rechazados sin crear cuentas parciales.
- **SC-003**: El 100% de los intentos con credenciales correctas confirman la autenticación y el 100%
  de los intentos con credenciales incorrectas son rechazados.
- **SC-004**: Los rechazos por username inexistente y contraseña incorrecta son indistinguibles para
  la persona que intenta ingresar.
- **SC-005**: Al menos 90% de las personas de una prueba de uso pueden corregir un formulario
  inválido en el siguiente intento usando únicamente los mensajes mostrados.
- **SC-006**: Una persona registrada puede completar el inicio de sesión en menos de 30 segundos sin
  asistencia.
- **SC-007**: El 100% de las solicitudes protegidas con un token válido emitido hace menos de 30
  minutos acceden a la identidad actual, y el 100% de los tokens vencidos o alterados son rechazados.
- **SC-008**: Ningún inicio de sesión ni solicitud autenticada crea filas de sesión o token en
  PostgreSQL.
- **SC-009**: El 100% de los intentos de registro con contraseñas distintas se detienen antes de
  crear una cuenta, y ambos campos permiten comprobar visualmente su contenido y volver a ocultarlo
  sin perder lo escrito.

## Assumptions

- El departamento puede no existir y por eso es el único dato opcional de la dirección.
- Ciudad y provincia son textos libres; no dependen de catálogos ni se limitan a Argentina.
- El número domiciliario admite valores no numéricos como `123 Bis` y `S/N`.
- Los espacios interiores de nombres y dirección se conservan; solamente se quitan los exteriores.
- Email y username se comparan sin distinguir mayúsculas para evitar identidades visualmente
  duplicadas; el sistema conserva una representación consistente para mostrarlos.
- Repetir un dígito o símbolo satisface el conteo de la contraseña porque la regla exige cantidad de
  caracteres, no variedad.
- La confirmación de contraseña existe solamente para detectar errores de escritura en el
  formulario; no se persiste ni forma parte de la solicitud de registro enviada al backend.
- La validación de email comprueba formato, pero no confirma propiedad ni capacidad de recepción.
- El frontend conserva el JWT mientras permanezca abierta la pestaña; esto permite sobrevivir una
  recarga, pero no compartir la sesión entre pestañas ni conservarla al cerrar la pestaña.
- Al no existir persistencia ni lista de revocación, un JWT válido no puede invalidarse antes de sus
  30 minutos de vigencia. Esta limitación es aceptada para el alcance académico.
- La conservación, edición y eliminación de datos personales fuera de estos flujos quedan fuera del
  alcance actual.
