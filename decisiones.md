# Decisiones técnicas

## TP2 — Contenedores

### Aplicación elegida

Elegí **FitPro**, una aplicación web académica para gestionar ejercicios, sesiones y rutinas de
entrenamiento. Incluye registro e inicio de sesión, un backend HTTP, una SPA y persistencia
relacional.

La aplicación cumple los criterios propuestos por la cátedra:

- **Construcción y ejecución local:** backend y frontend poseen comandos de compilación definidos y
  el sistema completo puede iniciarse con Docker Compose.
- **Posibilidad de testing:** existen pruebas automatizadas para reglas y comportamiento del backend
  Go y para componentes del frontend React.
- **Comprensión y modificación:** el backend es un monolito modular con una organización explícita
  por capas. El frontend utiliza componentes React y estado local, sin una arquitectura distribuida.
- **Tamaño acotado:** el alcance se concentra en autenticación y en el CRUD de ejercicios, sesiones
  y rutinas. Esto ofrece comportamiento suficiente para los próximos trabajos prácticos sin agregar
  complejidad innecesaria.

El proyecto es propio e individual. Lo utilizaré como base para los siguientes trabajos prácticos y
para el integrador.

### Contenerización del backend

El backend usa un Dockerfile multi-stage:

1. La etapa de construcción parte de `golang:1.26.5-alpine`.
2. Primero copia `go.mod` y `go.sum` y ejecuta `go mod download`. Esto permite reutilizar la caché de
   dependencias cuando solo cambia el código fuente.
3. Después copia el código y genera un binario Linux estático con `CGO_ENABLED=0`.
4. La etapa final parte de `scratch` y contiene únicamente el binario `/api`.

Elegí `scratch` porque el backend no necesita shell, compilador ni herramientas del sistema durante
la ejecución. Esto reduce el tamaño y la superficie de ataque. El proceso se ejecuta con el usuario
no privilegiado `65532:65532`, expone el puerto interno `8080` y usa `/api` como `ENTRYPOINT`.

Las migraciones SQL se incluyen dentro del binario y se aplican antes de iniciar el servidor HTTP.
Se ejecutan ordenadas, dentro de una transacción, y se registran en `schema_migrations`. Así no hace
falta levantar un contenedor exclusivo para migraciones y un error de esquema impide iniciar un
backend inconsistente.

### Contenerización del frontend

El frontend también usa un Dockerfile multi-stage:

1. La etapa de construcción parte de `node:26.7-alpine`.
2. Copia primero `package.json` y `package-lock.json` y ejecuta `npm ci` para obtener una instalación
   reproducible y aprovechar la caché.
3. Compila TypeScript y React con Vite.
4. La etapa final parte de `nginx:1.28-alpine` y recibe solamente los archivos estáticos de `dist`.

Node y las dependencias de desarrollo no viajan en la imagen final. Nginx sirve la SPA y resuelve
rutas del frontend mediante `try_files ... /index.html`.

La SPA consume rutas relativas `/api`. Nginx las reenvía a `http://backend:8080`, donde `backend` es
el nombre DNS del servicio dentro de la red de Compose. Elegí este enfoque porque evita fijar una URL
del backend dentro del bundle y mantiene frontend y API bajo el mismo origen, sin agregar CORS.

Cada contexto de construcción tiene su propio `.dockerignore`. Se excluyen archivos de Git,
dependencias locales, builds previos, logs, coberturas y archivos de entorno.

### Orquestación con Docker Compose

`docker-compose.yml` declara tres servicios:

- `db`: PostgreSQL 18.4 Alpine.
- `backend`: API Go construida desde `backend/Dockerfile`.
- `frontend`: SPA construida desde `frontend/Dockerfile` y servida por Nginx.

El backend se conecta a PostgreSQL usando `db:5432`. No usa `localhost`, porque dentro de un
contenedor `localhost` identifica al propio contenedor. Compose proporciona resolución DNS mediante
el nombre del servicio.

`depends_on` espera que el healthcheck de PostgreSQL informe que la base acepta conexiones. Esto es
necesario porque el orden de creación de contenedores no garantiza que la base ya esté preparada.

El backend solo usa `expose`, por lo que queda disponible dentro de la red de Compose. El único punto
de entrada de la aplicación es el frontend, publicado en `127.0.0.1:3000` por defecto. Nginx realiza
el proxy interno hacia la API.

### Persistencia y configuración

El volumen nombrado `postgres_data` conserva el contenido de PostgreSQL fuera de la capa efímera del
contenedor. Por eso `docker compose down` elimina contenedores y red, pero mantiene usuarios,
ejercicios, sesiones y rutinas. `docker compose down -v` elimina también el volumen y reinicia la
base desde cero en el siguiente arranque.

Backend y frontend son descartables. No guardan estado persistente dentro de sus contenedores.

Credenciales y secreto JWT entran mediante variables de entorno. El archivo `.env` real está
ignorado por Git y `.env.example` conserva solamente nombres y valores no secretos necesarios para
configurar una instalación nueva.

### Registry y arquitectura

Las imágenes publicadas son:

- `ghcr.io/germorini/backend:v1.0.0`
- `ghcr.io/germorini/frontend:v1.0.0`

El tag `v1.0.0` sigue versionado semántico. `docker-compose.registry.yml` mantiene la misma base,
red, volumen, healthcheck y configuración, pero usa esas imágenes mediante `image:` en lugar de
construirlas desde el código.

Las imágenes se construyeron en una máquina Linux `amd64`/`x86_64`. En este TP se publica esa única
arquitectura. Una publicación multi-arquitectura se podrá resolver posteriormente con
`docker buildx`.

### Problemas encontrados

#### Comunicación entre SPA y backend

Una SPA ejecuta sus solicitudes dentro del navegador. El nombre `backend` pertenece a la red interna
de Compose y no puede resolverse directamente desde el navegador. Se resolvió usando rutas relativas
`/api` y configurando Nginx como proxy inverso hacia `backend:8080`.

#### Disponibilidad de PostgreSQL

Iniciar primero el contenedor de la base no asegura que PostgreSQL ya acepte conexiones. Se agregó
un healthcheck con `pg_isready`, y el backend depende de su estado saludable.

#### Migraciones sin contenedor adicional

Inicialmente las migraciones se ejecutaban mediante un servicio temporal. Se reemplazó por un
ejecutor embebido en el backend. El backend abre la conexión, bloquea ejecuciones concurrentes,
aplica solamente migraciones pendientes y después inicia HTTP.

#### Persistencia durante recreaciones

Se comprobó que recrear contenedores sin `-v` conserva los ejercicios existentes. También se
comprobó que `down -v` elimina el volumen: las credenciales anteriores dejan de existir después del
nuevo arranque.

### Uso de inteligencia artificial

Utilicé inteligencia artificial para crear e iterar el código de la aplicación y para asistir la
redacción de `decisiones.md` y `evidencias.md`.

Revisé el resultado mediante la compilación, las pruebas automatizadas disponibles, la ejecución
completa con Docker Compose y las pruebas manuales documentadas en las capturas. Los commits, ramas,
tags, comandos de Docker, publicación de imágenes y capturas fueron realizados manualmente por mí.
También revisé los archivos generados y puedo explicar las decisiones técnicas usadas.

