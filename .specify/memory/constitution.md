<!--
Sync Impact Report
- Version change: 1.0.0 → 1.1.0
- Modified principles:
  - VIII. Frontend pequeño y explícito → VIII. Frontend pequeño, explícito y consistente
  - IX. Dependencias justificadas → IX. Dependencias justificadas
- Added sections: ninguna
- Removed sections: ninguna
- Templates requiring updates:
  - ✅ updated: .specify/templates/plan-template.md
  - ✅ no change required: .specify/templates/spec-template.md
  - ✅ updated: .specify/templates/tasks-template.md
  - ✅ no change required: .specify/templates/checklist-template.md
  - ✅ no command templates present: .specify/templates/commands/*.md
  - ✅ updated: specs/001-user-auth/plan.md
  - ✅ updated: specs/001-user-auth/research.md
  - ✅ updated: specs/001-user-auth/quickstart.md
- Follow-up TODOs: ninguno
-->
# Inge Soft 3 Academic Application Constitution

## Core Principles

### I. Objetivo académico y simplicidad

La aplicación DEBE mantenerse deliberadamente pequeña, clara, modificable y explicable durante
una defensa académica. Cada cambio DEBE responder a un requisito actual y demostrable. El diseño
NO DEBE anticipar funcionalidades futuras ni incorporar infraestructura, capas o patrones cuyo
único fundamento sea un posible uso posterior. Una solución explícita y sencilla TIENE prioridad
sobre una alternativa genérica más compleja.

Rationale: el proyecto existe para practicar ingeniería de software durante el semestre, no para
simular la escala o complejidad de un sistema de producción grande.

### II. Monolito modular

La aplicación DEBE ser un único monolito modular con frontend React, backend Go y PostgreSQL como
única base de datos. Los módulos DEBEN representar dominios o áreas funcionales coherentes y PUEDEN
comunicarse dentro del mismo proceso. El proyecto NO DEBE usar microservicios ni comunicación
distribuida. Redis, Kafka, RabbitMQ, service meshes, bases adicionales u otros componentes externos
solo PUEDEN incorporarse por un requisito posterior explícito de la materia.

Rationale: un único despliegue conserva integración real sin sumar complejidad operativa ajena al
objetivo académico.

### III. Backend en tres capas

Cada módulo backend DEBE respetar el flujo `controller -> service -> repository`:

- Controller: DEBE recibir HTTP, parsear y validar estructura, transformar solicitudes y respuestas,
  invocar services y traducir resultados o errores a HTTP. NO DEBE contener reglas de negocio, SQL
  ni acceso directo a PostgreSQL.
- Service: DEBE implementar casos de uso y reglas de negocio, coordinar repositories y controlar
  transacciones cuando corresponda. NO DEBE depender de HTTP, generar respuestas HTTP ni contener
  SQL o detalles innecesarios de persistencia.
- Repository: DEBE encapsular consultas y persistencia PostgreSQL y devolver los datos requeridos
  por services. NO DEBE implementar casos de uso, decidir reglas de negocio ni depender de
  controllers.

Las dependencias `controller -> repository`, `repository -> service` y `repository -> controller`
están prohibidas.

Rationale: límites pequeños y explícitos permiten ubicar responsabilidades y probar reglas sin
arrancar el servidor HTTP.

### IV. Organización por módulo funcional

El backend DEBE organizarse por módulo funcional, con una estructura conceptual
`internal/<module>/{controller,service,repository}`. Cada módulo DEBE contener solamente elementos
necesarios para sus capacidades actuales. Interfaces, factories, adapters, managers, providers,
handlers genéricos y buses internos solo PUEDEN crearse cuando resuelvan un problema concreto ya
presente y documentado. Clean Architecture, arquitectura hexagonal, CQRS, Event Sourcing y DDD
táctico completo están prohibidos salvo requisito futuro explícito.

Rationale: agrupar por capacidad mantiene próximos cambios relacionados y evita carpetas globales
que diluyen propiedad funcional.

### V. Modelo relacional mínimo

PostgreSQL DEBE ser la única persistencia. El modelo DEBE usar relaciones, claves primarias, claves
foráneas y constraints para invariantes que PostgreSQL pueda garantizar razonablemente. DEBE evitar
duplicación innecesaria y NO DEBE usar JSON cuando exista una estructura relacional clara. Tablas
genéricas, sistemas de metadata y estructuras para requisitos hipotéticos están prohibidos. Cada
feature DEBE definir durante su planificación únicamente las decisiones de datos que necesita.

Rationale: un esquema pequeño, normalizado y restringido hace visibles las reglas persistentes y
facilita verificar su cumplimiento.

### VI. Configuración externa y reproducible

Toda configuración dependiente del entorno DEBE poder cambiar sin modificar código fuente. La
conexión PostgreSQL, puertos, hosts, credenciales y valores de despliegue DEBEN provenir de variables
de entorno u otro mecanismo externo explícitamente requerido. Los secretos NO DEBEN almacenarse en
el repositorio. La ejecución local DEBE ser reproducible y los comandos de compilación, prueba y
ejecución DEBEN ser claros.

Rationale: separar configuración y código habilita prácticas de integración y despliegue sin
introducir infraestructura adicional.

### VII. Testabilidad arquitectónica

La facilidad de testing DEBE influir en el diseño. Las reglas de negocio DEBEN residir en services o
unidades equivalentes que puedan probarse sin servidor HTTP. El proyecto DEBE ofrecer comportamiento
significativo suficiente para al menos 8 tests útiles de backend y 4 de frontend. Los tests DEBEN
cubrir conducta real: validaciones, reglas, constraints, autorización, estados, transiciones, casos
borde o interfaz. Verificar solamente un código HTTP NO constituye por sí solo un test útil. Cada
nueva feature DEBE definir criterios de aceptación observables y comportamiento verificable. Está
prohibido inventar reglas únicamente para aumentar el conteo.

Rationale: los tests son parte del aprendizaje y deben demostrar decisiones reales del sistema.

### VIII. Frontend pequeño, explícito y consistente

React DEBE implementar una interfaz pequeña mediante componentes simples, formularios validables y
estados explícitos. Estado local DEBE preferirse mientras sea suficiente. Abstracciones prematuras,
arquitecturas frontend complejas y dependencias grandes para problemas simples están prohibidas.
daisyUI DEBE ser el sistema de componentes y estilos de interfaz; DEBE integrarse sobre una versión
compatible de Tailwind CSS mediante el mecanismo mínimo requerido por el build existente. Los
componentes estándar de daisyUI DEBEN preferirse antes de crear componentes visuales o CSS propios.
Un wrapper, tema personalizado o abstracción adicional solo PUEDE agregarse por una necesidad actual
documentada. El frontend PUEDE validar entradas y manejar comportamiento visual, pero el backend
DEBE conservar la autoridad sobre reglas de negocio importantes.

Rationale: una interfaz directa y un vocabulario visual común reducen estados ocultos, CSS disperso
y decisiones visuales repetidas sin trasladar reglas de negocio al cliente.

### IX. Dependencias justificadas

El número de dependencias DEBE mantenerse reducido. Cada dependencia nueva DEBE documentar una
ventaja concreta frente a las herramientas existentes. Dependencias experimentales, poco
mantenidas, innecesariamente complejas, ligadas a servicios externos o perjudiciales para
compilación, testing, contenerización o despliegue NO DEBEN incorporarse. daisyUI, Tailwind CSS y la
integración mínima de Tailwind con el build quedan autorizadas como stack visual obligatorio; esta
autorización NO se extiende a otras bibliotecas de componentes, iconos, formularios o temas.

Rationale: cada dependencia amplía superficie de aprendizaje, fallos, seguridad y mantenimiento.

### X. Prioridad de decisiones

Toda elección entre soluciones técnicamente válidas DEBE aplicar este orden:

1. cumplir correctamente el requisito actual;
2. facilitar comprensión y explicación;
3. facilitar testing;
4. facilitar ejecución, compilación y despliegue;
5. minimizar código y dependencias;
6. mantener consistencia arquitectónica;
7. considerar extensibilidad futura.

La extensibilidad futura NO DEBE justificar por sí sola abstracción o infraestructura adicional.

Rationale: un orden estable vuelve revisables las decisiones y protege la finalidad académica.

### XI. Simplicidad ante la duda

Ante una solución simple y otra más sofisticada, el equipo DEBE elegir la simple cuando cumpla los
requisitos actuales y permita pruebas adecuadas. Toda excepción DEBE registrar el requisito concreto
que la exige, la alternativa simple evaluada y por qué resulta insuficiente.

Rationale: esta regla resuelve ambigüedades a favor del propósito central del proyecto.

## Restricciones tecnológicas y arquitectónicas

- Stack obligatorio: React con daisyUI y Tailwind CSS, Go y PostgreSQL.
- Unidad de arquitectura y despliegue: monolito modular; microservicios prohibidos.
- Flujo backend obligatorio: `controller -> service -> repository` dentro de cada módulo funcional.
- Persistencia: un esquema relacional PostgreSQL mínimo, con invariantes expresadas mediante
  constraints cuando sea razonable.
- Configuración: valores ambientales externos al código y secretos fuera del repositorio.
- Infraestructura adicional: prohibida sin requisito explícito de la materia.
- Patrones o dependencias nuevos: requieren problema actual y justificación verificable.

## Flujo de desarrollo y controles de calidad

- Cada especificación DEBE delimitar alcance, fuera de alcance, comportamiento observable y casos
  borde; NO DEBE incluir capacidades especulativas.
- Cada plan DEBE superar un Constitution Check antes de investigar o diseñar y repetirlo después del
  diseño. Toda desviación DEBE registrarse en Complexity Tracking.
- Cada plan DEBE preservar organización modular y límites entre controller, service y repository.
- Cada plan con interfaz DEBE usar daisyUI sobre Tailwind CSS, reutilizar componentes existentes y
  justificar cualquier CSS, tema o wrapper propio.
- Cada cambio con comportamiento DEBE incluir pruebas proporcionales a sus reglas y riesgos. El
  conjunto del proyecto DEBE mantener al menos 8 tests útiles de backend y 4 de frontend.
- Cada revisión DEBE verificar secretos ausentes, configuración externa, dependencias justificadas,
  esquema relacional mínimo y ausencia de infraestructura no autorizada.
- Cada entrega DEBE demostrar comandos reproducibles de compilación, testing y ejecución aplicables
  al cambio.

## Governance

Esta constitución prevalece sobre plantillas, planes y prácticas que la contradigan. Una enmienda
DEBE documentar motivo, impacto, artefactos afectados y estrategia de migración cuando cambie código
o estructura existente. Su aprobación requiere revisión explícita del equipo académico responsable.

Las versiones siguen Semantic Versioning: MAJOR para eliminar o redefinir reglas de forma
incompatible; MINOR para agregar principios o ampliar obligaciones materialmente; PATCH para
aclaraciones sin cambio normativo. Toda enmienda DEBE actualizar versión, fecha y Sync Impact Report.

Toda especificación, plan, lista de tareas y revisión de código DEBE comprobar cumplimiento. Las
excepciones solo son válidas cuando un requisito actual las exige y quedan justificadas en Complexity
Tracking. El cumplimiento DEBE revisarse antes de integrar o entregar cada feature.

**Version**: 1.1.0 | **Ratified**: 2026-08-13 | **Last Amended**: 2026-08-13
