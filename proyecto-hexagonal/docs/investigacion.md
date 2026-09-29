# Arquitectura Hexagonal (Ports & Adapters)

**Stack:** Vite + React + TypeScript · Go (Golang) · Redis · WebSocket

Documento técnico — Presentación 1: Exploración de Estilos Arquitectónicos y Stack Tecnológico
*Entrega final: investigación, análisis arquitectónico, diseño, implementación y lecciones aprendidas*

---

## 1. Investigación del Estilo Arquitectónico: Arquitectura Hexagonal

### 1.1 Definición

**Qué es:** un estilo arquitectónico propuesto por Alistair Cockburn (2005) que organiza el sistema en tres zonas concéntricas: un núcleo de dominio (lógica de negocio pura), una capa de puertos (interfaces que declaran cómo el núcleo se comunica hacia afuera) y una capa de adaptadores (implementaciones concretas que conectan esos puertos con tecnologías reales como HTTP, WebSocket, bases de datos o colas de mensajes). El objetivo central es que el dominio no dependa de ningún detalle de infraestructura.

**Qué no es:** no es un framework, ni una librería, ni exige una tecnología específica; tampoco es sinónimo de microservicios (se puede aplicar dentro de un monolito) ni de Clean/Onion Architecture, aunque comparte con ellas el principio de inversión de dependencias hacia el dominio.

### 1.2 Clasificación del estilo

Pertenece a la familia de arquitecturas de dominio o "círculos concéntricos", junto con Onion Architecture y Clean Architecture. Es un estilo estructural (organiza la lógica interna del sistema), no distribuido ni de datos; puede combinarse con cualquier estilo distribuido (monolito, microservicios, serverless) porque solo regula cómo se organiza el código dentro de un componente.

### 1.3 Características principales

- **Puertos (Ports):** interfaces definidas por el dominio que describen qué necesita o qué expone, sin saber cómo se implementa (ej. una interfaz Repository, o una interfaz Notifier).
- **Adaptadores (Adapters):** implementaciones concretas de un puerto. Se dividen en adaptadores de entrada/driving (quien usa el sistema: un controlador REST, un handler de WebSocket) y de salida/driven (lo que el sistema usa: un repositorio en Redis, un cliente HTTP externo).
- **Independencia del dominio:** el núcleo de negocio no importa librerías de frameworks, drivers de base de datos ni SDKs externos; solo conoce sus propios puertos.
- **Simetría:** a diferencia de una arquitectura en capas clásica, no hay una dirección única "arriba-abajo"; el dominio queda en el centro y todo apunta hacia él (Regla de Dependencia).
- **Alta testeabilidad:** el dominio se puede probar con pruebas unitarias puras, sustituyendo los adaptadores por dobles de prueba (mocks/stubs) de los puertos.

### 1.4 Historia y evolución

Propuesta por Alistair Cockburn en 2005 bajo el nombre "Ports and Adapters", como respuesta al problema de que la arquitectura en capas tradicional terminaba acoplando la lógica de negocio a la interfaz de usuario y a la base de datos. La idea del dominio aislado influyó directamente en Robert C. Martin para formalizar Clean Architecture (2012) y en Jeffrey Palermo para Onion Architecture (2008); las tres comparten la misma regla de dependencia hacia el centro, con distinta nomenclatura de capas. Con la llegada de microservicios y contenedores (2015 en adelante), Hexagonal se popularizó como el estilo por defecto para diseñar cada servicio individual, ya que facilita cambiar de protocolo de entrada (REST, gRPC, WebSocket) o de motor de persistencia sin tocar el dominio.

### 1.5 Ventajas y desventajas

| Ventajas | Desventajas |
|---|---|
| Alta testeabilidad del dominio sin infraestructura real | Curva de aprendizaje más alta para equipos junior |
| Facilita cambiar tecnología (ej. cambiar WebSocket por REST, o Redis por Postgres) sin tocar la lógica de negocio | Más archivos/carpetas e indirecciones que un CRUD simple; puede sentirse como sobre-ingeniería en proyectos muy pequeños |
| Bajo acoplamiento y alta cohesión; cumple bien SRP y DIP | Requiere disciplina del equipo para no "filtrar" detalles de infraestructura al dominio |
| Permite convivir varios adaptadores de entrada a la vez (ej. WebSocket y REST) reutilizando el mismo dominio | El mapeo entre modelos de dominio y modelos de infraestructura (DTOs) añade código repetitivo |

### 1.6 Problemas comunes y patrones asociados

- **Fuga de infraestructura al dominio:** ocurre cuando structs de la base de datos o tipos de un framework se usan directamente en el dominio; se evita con DTOs y mapeadores en los adaptadores.
- **Puertos mal diseñados (demasiado amplios):** un puerto que expone detalles técnicos (ej. paginación con cursores propios de Redis) rompe la abstracción; se corrige aplicando Interface Segregation.
- **Patrones que suele necesitar este estilo:** Dependency Injection (para inyectar adaptadores en tiempo de arranque), Repository (puerto de salida hacia persistencia), Adapter (patrón GoF, base conceptual del estilo), Facade (para exponer casos de uso del dominio a los adaptadores de entrada) y, en este proyecto, Observer/Publish-Subscribe para el canal de WebSocket.

**Patrones aplicables y cuándo usarlos**

| Patrón | Cuándo usarlo | Dónde se usa en este proyecto |
|---|---|---|
| Ports & Adapters (Adapter) | Cuando el dominio debe comunicarse con tecnologías intercambiables | Todo el backend: `ports/` define los contratos y `adapters/` los implementa |
| Dependency Injection | Cuando un componente necesita colaboradores que pueden cambiar (o simularse en pruebas) | `NewEventService(rooms, events, notify, newID)` y el ensamblado en `main.go` |
| Repository | Cuando se quiere ocultar cómo y dónde se persisten los datos | `RoomRepository` y `EventRepository`, implementados con Redis |
| Facade / Application Service | Cuando varios adaptadores necesitan un único punto de entrada a los casos de uso | `EventService` (puerto de entrada) |
| Publish–Subscribe | Cuando un evento debe llegar a varios receptores sin acoplarlos | `EventBroadcaster` y `EventSubscriber` sobre Redis Pub/Sub |
| DTO / Mapper | Cuando el formato externo (JSON) no debe condicionar el modelo interno | `protocol.go` en el adaptador WebSocket |
| Factory (constructores con validación) | Cuando un objeto solo debe existir en estado válido | `NewRoom`, `NewParticipant`, `NewEvent` |

### 1.7 Casos de uso

**Cuándo usarlo:** cuando la lógica de negocio es valiosa y debe sobrevivir a cambios tecnológicos; cuando se prevén múltiples canales de entrada (API REST + WebSocket + CLI) o múltiples proveedores de persistencia; cuando se requiere probar el dominio de forma aislada y rápida.

**Cuándo no usarlo:** en prototipos desechables, scripts pequeños o CRUDs triviales donde el costo de las capas adicionales no se justifica frente al beneficio de flexibilidad futura.

### 1.8 Casos de aplicación reales

- **Netflix (Studio Engineering):** el equipo de Studio Workflows documentó en el Netflix Tech Blog ("Ready for changes with Hexagonal Architecture") cómo usó Arquitectura Hexagonal para construir una aplicación que debía integrarse con datos repartidos en servicios con distintos protocolos (gRPC, JSON API, GraphQL) y que cambiaban de lugar mientras se descomponía un monolito. Los adaptadores absorben esos cambios sin tocar el dominio.
- **Migraciones de monolito a servicios:** el estilo se usa para aislar la lógica de negocio de los detalles de infraestructura y así poder mover una fuente de datos (de una base local a un servicio) cambiando solo un adaptador.
- **Servicios en Go:** es habitual estructurar los proyectos con `cmd/` e `internal/`, separando el dominio de los adaptadores; es exactamente la organización de este proyecto.

---

## 2. Investigación del Stack Tecnológico

### 2.1 Frontend — Vite + React + TypeScript

**Qué es / qué no es:** Vite es una herramienta de build y servidor de desarrollo (no un framework de UI); React es una librería para construir interfaces basadas en componentes y Virtual DOM (no un framework integral como Angular); TypeScript es un superconjunto tipado de JavaScript que se transpila a JS (no un lenguaje distinto en tiempo de ejecución).

**Características principales:** Vite usa ESBuild/Rollup y servidores de módulos nativos (ESM) para arranque casi instantáneo y Hot Module Replacement muy rápido; React aporta un modelo declarativo de componentes y hooks para manejar estado y efectos; TypeScript aporta tipado estático, autocompletado y detección de errores en tiempo de compilación.

**Historia y evolución:** React nace en Facebook (2013); Vite es creado por Evan You (autor de Vue) en 2020 como alternativa a Webpack, hoy es el bundler recomendado por el propio equipo de React para nuevos proyectos SPA; TypeScript es de Microsoft (2012) y hoy es el estándar de facto para proyectos React de tamaño mediano/grande.

**Ventajas y desventajas:** Ventajas: arranque y recarga extremadamente rápidos, tipado que reduce errores en tiempo de ejecución, enorme ecosistema y talento disponible. Desventajas: Vite/React puro no incluye SSR (a diferencia de Next.js), por lo que el enrutamiento y el SEO deben resolverse aparte; TypeScript añade una etapa de compilación y curva de aprendizaje de tipos.

**Casos de uso:** ideal para SPAs y dashboards interactivos como el cliente de una aplicación en tiempo real sobre WebSocket, donde no se requiere SSR/SEO. No es la mejor opción si el proyecto necesita renderizado en servidor o generación estática (ahí conviene un meta-framework como Next.js o Astro).

**Casos de aplicación:** usado como stack de frontend en herramientas internas, paneles de administración y clientes de aplicaciones colaborativas en tiempo real (chats, tableros compartidos, dashboards de monitoreo).

### 2.2 Backend — Go (Golang)

**Qué es / qué no es:** Go es un lenguaje de programación compilado, con tipado estático y concurrencia nativa (goroutines y channels), creado por Google. No es un framework; para este proyecto se usa con su librería estándar (net/http, encoding/json) más una librería ligera de WebSocket (ej. gorilla/websocket o nhooyr.io/websocket).

**Características principales:** compilación a binario único sin dependencias externas, concurrencia ligera mediante goroutines (miles de conexiones concurrentes con bajo consumo de memoria), tipado estático con inferencia, recolector de basura, y una librería estándar muy completa para HTTP y JSON.

**Historia y evolución:** creado en Google por Robert Griesemer, Rob Pike y Ken Thompson, publicado en 2009 y estable desde 2012; su modelo de concurrencia lo hizo el lenguaje por defecto para infraestructura cloud-native (Docker, Kubernetes y Terraform están escritos en Go).

**Ventajas y desventajas:** Ventajas: rendimiento cercano a C, manejo nativo de miles de conexiones concurrentes (ideal para servidores WebSocket), despliegue simple (un solo binario, contenedores muy livianos). Desventajas: sintaxis más verbosa para manejo de errores (`if err != nil` repetido), ecosistema de frameworks web más minimalista que en Java/Node, sin genéricos completos hasta versiones recientes.

**Casos de uso:** excelente para backends con alta concurrencia y comunicación en tiempo real como servidores WebSocket, proxies, APIs de baja latencia. Menos indicado cuando se requiere un ecosistema ORM/ADMIN muy maduro tipo Django o Rails para CRUDs administrativos complejos.

**Casos de aplicación:** Docker y Kubernetes están escritos en Go, lo que lo consolidó como lenguaje de la infraestructura cloud-native; por su modelo de concurrencia también es una elección frecuente para servidores con muchas conexiones simultáneas.

### 2.3 Persistencia — Redis (Clave-Valor)

**Qué es / qué no es:** Redis es un motor de datos en memoria de tipo clave-valor (in-memory data structure store); no es una base de datos relacional ni garantiza por defecto durabilidad ACID completa como Postgres, aunque ofrece persistencia opcional (RDB/AOF).

**Características principales:** estructuras de datos ricas (strings, hashes, lists, sets, sorted sets, streams), latencia sub-milisegundo por operar en memoria, soporte nativo de Pub/Sub (clave para notificar eventos a clientes WebSocket), expiración automática de claves (TTL) y persistencia opcional a disco.

**Historia y evolución:** creado por Salvatore Sanfilippo en 2009 para resolver problemas de escalabilidad de un sitio en tiempo real; hoy es el motor clave-valor/caché más usado de la industria y ha incorporado módulos para búsqueda, JSON y vectores (RedisJSON, RediSearch).

**Ventajas y desventajas:** Ventajas: velocidad extrema, estructuras de datos versátiles, Pub/Sub nativo que encaja directamente con un backend de WebSocket para difundir eventos en tiempo real. Desventajas: al vivir en memoria, el tamaño de datos está limitado por RAM disponible; la durabilidad y las relaciones complejas entre entidades son más débiles que en una base relacional.

**Casos de uso:** ideal como almacén principal de estado efímero/tiempo real (sesiones, presencia de usuarios, contadores, colas) y como canal Pub/Sub para difundir mensajes a través de WebSocket. No es la mejor opción como única fuente de verdad para datos transaccionales complejos con integridad referencial estricta.

**Casos de aplicación:** caché, gestión de sesiones, colas ligeras, contadores y difusión de mensajes en tiempo real. La Stack Overflow Developer Survey 2025 lo ubica como la quinta base de datos más usada y con un crecimiento de 8 puntos frente al año anterior.

### 2.4 Protocolo de Integración — WebSocket

**Qué es / qué no es:** WebSocket es un protocolo de comunicación full-duplex sobre una única conexión TCP persistente (RFC 6455); no es una variante de REST ni de HTTP en sentido estricto, aunque el handshake inicial se hace sobre HTTP antes de "actualizar" (upgrade) la conexión.

**Características principales:** conexión persistente bidireccional (servidor y cliente pueden enviar mensajes en cualquier momento, sin que el cliente tenga que solicitar), baja sobrecarga por mensaje frente a peticiones HTTP repetidas (polling), y soporte nativo en navegadores modernos y en la librería estándar de Go.

**Historia y evolución:** estandarizado por el IETF en 2011 como respuesta a las limitaciones de técnicas como long-polling para simular tiempo real sobre HTTP; hoy convive con alternativas como SSE (unidireccional) y WebTransport (más reciente, sobre QUIC).

**Ventajas y desventajas:** Ventajas: latencia mínima para eventos en tiempo real, menor overhead que hacer polling constante por HTTP, ideal para notificaciones push, chats y dashboards en vivo. Desventajas: requiere manejar el ciclo de vida de la conexión (reconexión, heartbeats), no cachea como HTTP tradicional, y algunos proxies/balanceadores necesitan configuración especial para sostener conexiones persistentes.

**Casos de uso:** aplicaciones que requieren actualizaciones en tiempo real en ambas direcciones: chats, juegos, dashboards colaborativos, notificaciones. No es la opción adecuada para operaciones puntuales tipo petición-respuesta donde REST es más simple y cacheable.

**Casos de aplicación:** aplicaciones de mensajería y colaboración en tiempo real (por ejemplo chats, tableros compartidos y paneles en vivo) usan WebSocket para sincronizar el estado entre clientes de forma instantánea.

### 2.5 Relación entre el estilo y las tecnologías seleccionadas

Arquitectura Hexagonal encaja de forma natural con este stack porque cada tecnología ocupa exactamente el rol de un adaptador, dejando el dominio (reglas de negocio) libre de dependencias concretas:

| Componente del hexágono | Tecnología | Rol |
|---|---|---|
| Adaptador de entrada (driving) | Vite + React + TS (cliente WebSocket) | Consume el puerto de entrada del dominio a través de mensajes WebSocket; no contiene lógica de negocio, solo UI y estado de presentación. |
| Adaptador de entrada (driving) | Go — handler de WebSocket | Recibe el mensaje entrante, lo traduce a un comando del dominio y lo invoca a través de un puerto (interfaz de caso de uso). |
| Núcleo de dominio | Go — paquete `domain/` puro | Contiene las reglas de negocio y los puertos (interfaces), sin importar `net/http`, `gorilla/websocket` ni el cliente de Redis. |
| Adaptador de salida (driven) | Go — repositorio Redis | Implementa el puerto de persistencia usando el cliente de Redis; el dominio solo conoce la interfaz, no el driver. |
| Canal de eventos | Redis Pub/Sub | Permite difundir eventos del dominio hacia múltiples conexiones WebSocket activas, incluso si hay varias instancias del backend. |

En otras palabras: el protocolo de integración (WebSocket) y la base de datos (Redis) son intercambiables sin tocar el dominio, siempre que se respete el contrato definido por los puertos; esa sustituibilidad es precisamente lo que Hexagonal promete y lo que se debe demostrar en la sustentación.

### 2.6 Qué tan común es este stack

Cada tecnología por separado es muy común; lo menos frecuente es la combinación exacta:

| Tecnología | Evidencia de adopción | Fuente |
|---|---|---|
| React | 85 % de uso entre los encuestados de State of JS 2025 (satisfacción 72 %); primera librería de frontend en las encuestas recientes | State of JS 2025 |
| Redis | Quinta base de datos más usada; +8 puntos de uso en 2025 | Stack Overflow Developer Survey 2025 |
| Go | Lenguaje consolidado en infraestructura y servicios; aparece entre los más admirados en la encuesta de 2025 | Stack Overflow Developer Survey 2025 (según el resumen de KORE1) |
| WebSocket | Estándar del IETF (RFC 6455) soportado por todos los navegadores | RFC 6455 |

La combinación React/TS + Go + Redis + WebSocket es menos frecuente como "paquete cerrado" que MERN o Next.js + PostgreSQL, pero es un stack coherente para tiempo real: Go sostiene muchas conexiones concurrentes, Redis aporta persistencia rápida y Pub/Sub, y WebSocket entrega los eventos al navegador. La comparación de demanda laboral y salarios se detalla en la matriz 3.4.

## 3. Análisis Arquitectónico

Se presentan las cuatro matrices solicitadas. Las matrices 3.1 a 3.3 se contrastan con el código real del proyecto (carpeta `backend/`); la matriz 3.4 usa fuentes públicas citadas al final.

### 3.1 Matriz de atributos de calidad vs. estilo

| Atributo de calidad | Cómo lo soporta Hexagonal | Cómo lo limita / riesgo | Evidencia en el proyecto |
|---|---|---|---|
| Testeabilidad | El dominio se prueba sustituyendo los adaptadores por dobles de prueba de los puertos; no requiere Redis ni un socket real. | Si un puerto se diseña demasiado técnico, las pruebas dependen indirectamente de infraestructura. | `internal/app/event_service_test.go` prueba el flujo completo con fakes en memoria. |
| Mantenibilidad / Modificabilidad | Cambiar Redis por otra base, o WebSocket por otro protocolo, solo implica un nuevo adaptador; el dominio no se toca. | Más archivos e indirecciones; en equipos pequeños cuesta ubicar dónde vive cada regla. | Las reglas de validación viven solo en `internal/domain/entities.go`. |
| Portabilidad | El núcleo no depende de ningún SDK ni framework. | Se pierde si la infraestructura se filtra al dominio. | `domain/`, `ports/` y `app/` no importan `redis` ni `websocket`. |
| Escalabilidad | Al pasar los eventos por Redis Pub/Sub, se pueden levantar varias instancias del backend sin duplicar lógica. | El estilo por sí solo no escala; depende de cómo se diseñe el adaptador de Redis. | El `Hub` recibe los eventos del puerto `EventSubscriber`, no de memoria local. |
| Rendimiento | En Go las interfaces son baratas; el costo de la indirección es mínimo. | Cada llamada pasa por una capa adicional (adaptador → puerto → dominio). | Goroutines por conexión; pipeline de Redis (`TxPipeline`) para guardar en una sola ida y vuelta. |
| Seguridad | La validación y los límites se concentran en los adaptadores de entrada y en el dominio. | El dominio no debe asumir que todo llegó validado. | Límite de mensaje (`SetReadLimit`), contenido máximo de 2000 caracteres, contenedor sin privilegios de root. |
| Disponibilidad / Resiliencia | Si falla un adaptador, el dominio decide cómo degradar. | Requiere diseñar políticas de error explícitas. | Si falla la difusión el evento ya quedó guardado; el cliente reconecta con backoff exponencial. |
| Despliegue (Deployability) | Backend y frontend se despliegan por separado; el backend es un único binario. | Más piezas que orquestar (backend, Redis, frontend). | `docker-compose.yml` levanta las 3 piezas con un comando. |

### 3.2 Matriz de principios vs. estilo

| Principio | Cómo lo cumple el estilo | Cómo se aplica en el proyecto | Valoración |
|---|---|---|---|
| **S** — Responsabilidad única | Cada zona tiene un solo motivo de cambio: el handler traduce, el servicio orquesta, las entidades validan, los repositorios persisten. | `handler.go` no contiene reglas de negocio; `entities.go` no conoce Redis. | Cumple |
| **O** — Abierto/Cerrado | Se agregan adaptadores nuevos sin modificar el servicio. | Un repositorio PostgreSQL implementaría `RoomRepository` sin tocar `event_service.go`. | Cumple |
| **L** — Sustitución de Liskov | Cualquier implementación de un puerto es intercambiable. | Los fakes de las pruebas y los adaptadores de Redis satisfacen las mismas interfaces. | Cumple |
| **I** — Segregación de interfaces | Puertos pequeños y específicos. | Cuatro puertos de salida separados (`RoomRepository`, `EventRepository`, `EventBroadcaster`, `EventSubscriber`) en vez de uno grande. | Cumple |
| **D** — Inversión de dependencias | Es el núcleo del estilo: la infraestructura depende del dominio. | `NewEventService` recibe interfaces; solo `main.go` conoce las implementaciones. | Cumple (principio central) |
| **KISS** | El estilo agrega capas, lo que puede ir contra la simplicidad. | Se evitaron frameworks: solo dos dependencias externas en el backend (`gorilla/websocket`, `go-redis`). | Cumple con matices |
| **DRY** | La validación se define una vez en el dominio y sirve para cualquier canal de entrada. | Los DTOs del adaptador repiten campos de las entidades (duplicación deliberada para no acoplar el dominio al formato JSON). | Cumple con matices |
| **YAGNI** | El estilo invita a diseñar puertos "por si acaso", lo que puede sobredimensionar. | No se implementó autenticación, CQRS ni persistencia de participantes: no lo exige el caso de uso. | Cumple |
| **PoLA** (mínima sorpresa) | Los puertos hacen explícito el contrato. | Los errores de negocio llegan al cliente como mensajes legibles; los errores de infraestructura no exponen detalles internos. | Cumple |
| **Ley de Demeter** | El adaptador de entrada solo habla con el puerto, no con repositorios. | `handler.go` llama únicamente a `ports.EventService`. | Cumple |
| **STUPID** (anti-principios) | Evita Singleton, acoplamiento fuerte, no testeabilidad, optimización prematura, nombres poco descriptivos y duplicación. | Sin variables globales (la inyección se hace en `main.go`); cada pieza es testeable; no hay optimizaciones prematuras. | Se evita (ver detalle abajo) |
| **Composición sobre herencia** | Los componentes se ensamblan por interfaces. | Go no tiene herencia: `eventService` se compone de tres puertos inyectados. | Cumple |

Detalle de STUPID: **S**ingleton (no hay estado global), **T**ight coupling (dependencias por interfaz), **U**ntestability (18 pruebas automáticas), **P**remature optimization (ninguna), **I**ndescriptive naming (nombres de dominio en español/inglés consistentes con el negocio), **D**uplication (validación centralizada; la única duplicación son los DTOs, justificada arriba).

### 3.3 Matriz de tácticas vs. estilo y stack (justificación de ADR)

Las tácticas siguen el catálogo de atributos de calidad de Bass, Clements y Kazman. Cada fila justifica un ADR (Architecture Decision Record) completo en `docs/adr/`.

| Atributo | Táctica | Justificación por el estilo | Justificación por el stack | Alternativa descartada | ADR |
|---|---|---|---|---|---|
| Modificabilidad | Intermediario (puertos y adaptadores) | Es la esencia de Hexagonal: aísla el dominio de la infraestructura. | Las interfaces de Go se satisfacen implícitamente, sin declarar "implements". | Arquitectura en capas con el dominio dependiendo del acceso a datos. | ADR-001 |
| Rendimiento | Concurrencia | El adaptador de entrada puede atender muchas conexiones sin afectar el dominio. | Goroutines ligeras: una por conexión. | Hilos del sistema por conexión (más memoria). | ADR-002 |
| Rendimiento / Persistencia | Mantener datos en memoria + estructuras adecuadas | El repositorio es un puerto: la tecnología es intercambiable. | Redis ofrece LIST, ZSET y Pub/Sub nativos. | PostgreSQL (más pesado para este caso de eventos efímeros). | ADR-003 |
| Interoperabilidad / Latencia | Canal bidireccional persistente | El protocolo es un adaptador, no una decisión del dominio. | WebSocket es nativo en el navegador y en `gorilla/websocket`. | REST con polling (más latencia y tráfico). | ADR-004 |
| Escalabilidad | Bus de eventos compartido | El puerto `EventBroadcaster` permite reemplazar la difusión local por una compartida. | Redis Pub/Sub entrega el evento a todas las instancias. | Difusión solo en memoria (no escala horizontalmente). | ADR-005 |
| Seguridad / Modificabilidad | Encapsular y validar entradas | Las entidades de dominio no se exponen al exterior. | DTOs propios del adaptador WebSocket con etiquetas JSON. | Serializar directamente las entidades. | ADR-006 |
| Despliegue | Contenerización | Cada adaptador se ensambla en `main.go`, listo para empaquetar. | Go compila a un binario estático: imagen final mínima. | Despliegue manual sin contenedores. | (en 5.4) |
| Disponibilidad | Reintento y monitoreo | El dominio no conoce las fallas de red; los adaptadores las gestionan. | Reconexión con backoff en el cliente; `/healthz` y healthcheck de Docker. | Sin reconexión automática. | (en 5.4) |

### 3.4 Matriz de mercado laboral vs. estilo y stack

**Cómo leerla:** los datos provienen de fuentes públicas consultadas en septiembre de 2026 y son referenciales. Los salarios cambian por empresa, ciudad y experiencia; verifiquen las cifras en las fuentes antes de la sustentación.

**Demanda y adopción**

| Elemento | Indicador | Fuente |
|---|---|---|
| React | 85 % de uso, satisfacción 72 % | State of JS 2025 |
| TypeScript | Lenguaje en expansión junto con Go y Rust, según el análisis de la encuesta | Stack Overflow Developer Survey 2025 |
| Go | Lenguaje admirado y de buena remuneración; posición 13 con 1,20 % en el índice TIOBE de junio de 2026 (−1,08 puntos interanual). *Dato de segunda mano: confirmar en tiobe.com.* | Stack Overflow 2025 (vía KORE1); TIOBE |
| Redis | +8 puntos de uso en 2025; quinta base de datos más usada | Stack Overflow Developer Survey 2025 |
| Arquitectura Hexagonal | No es una tecnología con estadísticas propias: se valora como competencia de diseño (Clean Architecture, DDD, microservicios) que aparece en ofertas de nivel semi-senior y senior. *No se encontraron cifras específicas del estilo.* | — |
| WebSocket | Competencia transversal (tiempo real); sin estadísticas propias | — |

**Salarios de referencia: desarrollador backend en Colombia (2026)**

| Nivel | Salario mensual (COP) | Equivalente aprox. (USD) | Fuente |
|---|---|---|---|
| Junior | 4.200.000 – 5.200.000 | 1.100 – 1.350 | Coderhouse, Sueldo Backend Colombia 2026 |
| Semi senior | 6.000.000 – 8.500.000 | 1.600 – 2.250 | Coderhouse |
| Senior | 10.000.000 o más (más variable) | 2.650 o más | Coderhouse |
| Promedio publicado | 4.525.953 por mes (13 sueldos reportados, feb. 2026) | — | Indeed Colombia |
| Junior / Senior (contratos en dólares) | — | 1.155 promedio junior; hasta 6.000 senior | Talently 2026 |
| Medellín (Glassdoor) | Promedio de 5.858.681 al año según el portal; rango típico 4.105.692 – 8.958.333 | — | Glassdoor. *Las cifras del portal parecen mensuales pese a la etiqueta anual: interpretarlas con cautela.* |

**Salarios de referencia en Estados Unidos (2025)**

| Rol | Mediana anual (USD) | Fuente |
|---|---|---|
| Desarrollador backend | 175.000 | Stack Overflow Developer Survey 2025 (EE. UU., 5.239 respuestas) |
| Desarrollador full-stack | 138.000 | Stack Overflow Developer Survey 2025 |
| Ingeniero de software (compensación total) | 192.500 | Levels.fyi 2025 |

**Proyección.** Redis sube en adopción; Go y TypeScript siguen creciendo en la encuesta de 2025, aunque Go pierde posición relativa en TIOBE frente a Rust (dato por confirmar). Las bandas mejoran cuando el candidato domina inglés y trabaja para clientes en dólares (nearshoring), y los salarios más altos se concentran en Bogotá y Medellín. Conclusión: es un stack empleable, con demanda sólida en React/TypeScript y Redis, y con Go como diferenciador de nicho en backend e infraestructura.

---

## 4. Diseño: Ejemplo Práctico y Funcional

El sistema se modeló con HLD y C4 Model. Todos los diagramas obligatorios están incluidos, además del diagrama de componentes y el modelo de datos (opcionales). Los archivos están en `docs/diagramas/`.

### 4.1 Diagrama de Alto Nivel (HLD)

![Diagrama de Alto Nivel](./diagramas/hld.png)

El usuario accede al cliente web (capa de presentación), que se comunica con el backend en Go (capa de negocio) mediante WebSocket. El backend delega la persistencia y la difusión de eventos a Redis (capa de datos). Cada color identifica una capa y cada flecha indica el protocolo y el sentido de la comunicación.

### 4.2 Diagrama de Contexto — C4 Nivel 1

![Diagrama de Contexto C4 Nivel 1](./diagramas/c4-contexto.png)

El sistema se modela como una única caja `[Software System]` que interactúa con personas `[Person]`. Siguiendo la norma de C4, este nivel no muestra tecnología: Go, Redis y la arquitectura interna aparecen en los niveles 2 y 3.

### 4.3 Diagrama de Contenedores — C4 Nivel 2

![Diagrama de Contenedores C4 Nivel 2](./diagramas/c4-contenedores.png)

El sistema se descompone en cuatro contenedores: el servidor web (nginx) que entrega los archivos estáticos, la aplicación web (React), el backend API (Go) y Redis. Cada relación indica su protocolo entre corchetes.

### 4.4 Diagrama de Componentes — C4 Nivel 3 (Arquitectura Hexagonal)

![Diagrama de Componentes C4 Nivel 3](./diagramas/c4-componentes.png)

La Arquitectura Hexagonal se modela en el Nivel 3, que es el nivel normativo de C4 para este estilo. El contenedor Backend API se divide en adaptadores de entrada (`WebSocket Handler`, `Hub`), núcleo (`EventService` y puertos de salida) y adaptadores de salida (`Redis Store`, `Redis PubSub`). Las flechas naranjas apuntan hacia el puerto: la infraestructura depende del dominio, nunca al revés.

### 4.5 Diagrama Dinámico (flujo principal)

![Diagrama Dinámico](./diagramas/c4-dinamico.png)

Flujo principal: publicar un evento. El usuario envía el mensaje, el backend valida y guarda el evento en Redis, lo publica por Pub/Sub, y el `Hub` lo reenvía por WebSocket a todos los clientes de la sala, incluido el autor.

### 4.6 Diagrama de Despliegue

![Diagrama de Despliegue](./diagramas/c4-despliegue.png)

Tres contenedores en una red de Docker Compose: `frontend` (nginx, puerto publicado 5173), `backend` (puerto publicado 8080) y `redis` (puerto interno 6379, con volumen `redis-data` para persistir). El navegador carga la SPA por HTTP y abre el WebSocket directamente contra el backend.

### 4.7 Modelo de datos

![Modelo de datos](./diagramas/modelo-datos.png)

Tres entidades interrelacionadas: `Room` (1) — (N) `Participant` y `Room` (1) — (N) `Event`. `Room` y `Event` se persisten en Redis; `Participant` es efímero y vive solo en la memoria del `Hub`. Como Redis no tiene integridad referencial, el servicio de aplicación verifica que la sala exista antes de aceptar un evento.

---

## 5. Implementación: Ejemplo Práctico y Funcional

### 5.1 Qué hace el sistema

Un sistema de **salas en tiempo real**: los usuarios crean o se unen a una sala y todo mensaje publicado llega al instante a todos los conectados a esa sala. El historial se guarda en Redis, así que un usuario que entra después ve los mensajes anteriores.

Caso de uso de extremo a extremo: `crear sala → unirse → publicar evento → difusión en tiempo real → consultar historial`.

### 5.2 Estructura del repositorio

```
proyecto-hexagonal/
├── backend/                         # Go 1.22
│   ├── cmd/server/main.go           # raíz de composición (inyección de dependencias)
│   ├── internal/
│   │   ├── domain/                  # entidades + validaciones (sin dependencias externas)
│   │   ├── ports/                   # interfaces de entrada y de salida
│   │   ├── app/                     # servicio de aplicación (implementa el puerto de entrada)
│   │   ├── adapters/
│   │   │   ├── in/websocket/        # adaptador de entrada: handler, hub y DTOs
│   │   │   └── out/redis/           # adaptadores de salida: repositorios y Pub/Sub
│   │   ├── platform/id/             # generador de identificadores
│   │   └── integration/             # prueba end-to-end con Redis real
│   ├── Dockerfile
│   └── go.mod
├── frontend/                        # Vite + React + TypeScript
│   ├── src/domain/                  # tipos, puerto RealtimeGateway y reductor de estado
│   ├── src/adapters/                # cliente WebSocket (implementa el puerto)
│   ├── src/components/              # Lobby y ChatRoom
│   └── Dockerfile
├── docs/                            # documento técnico, diagramas y ADRs
├── docker-compose.yml               # Redis + backend + frontend
└── README.md
```

### 5.3 Decisiones de implementación

- **Composición en un solo lugar:** `cmd/server/main.go` es el único archivo que conoce a la vez el dominio y los adaptadores concretos.
- **Protocolo WebSocket (JSON):** mensajes del cliente `list_rooms`, `create_room`, `join`, `publish`; respuestas `rooms`, `room_created`, `joined`, `event`, `error`.
- **Persistir antes de difundir:** el evento se guarda en Redis (`RPUSH`) y después se publica (`PUBLISH`). Si la difusión falla, el dato no se pierde.
- **Robustez:** validaciones en el dominio (vacío, longitud máxima), límite de tamaño de mensaje, reconexión del cliente con backoff, reintentos del backend al arrancar hasta que Redis responda, y cierre ordenado del servidor.
- **Frontend con el mismo estilo:** la interfaz depende del puerto `RealtimeGateway`; el cliente WebSocket es su adaptador.

### 5.4 Contenedores

`docker-compose.yml` define tres servicios: `redis` (con persistencia AOF, volumen y healthcheck), `backend` (imagen multi-etapa, usuario sin privilegios, healthcheck contra `/healthz`, arranca solo cuando Redis está sano) y `frontend` (compilación con Vite y servido con nginx).

### 5.5 Pruebas y resultados

| Nivel | Qué se prueba | Cantidad |
|---|---|---|
| Unitarias del dominio | Validaciones de entidades | 5 |
| Unitarias de aplicación | Flujo completo con dobles de prueba en memoria, errores y fallo de difusión | 6 |
| Integración (Redis real) | Dos clientes WebSocket reales: difusión en tiempo real, persistencia del historial, validaciones y salas inexistentes | 1 |
| Frontend | Reductor de estado | 6 |

Verificado durante el desarrollo: `go vet`, `go test -race` con Redis real, `npm run build` (TypeScript estricto) y `vitest`; además, el binario real (`cmd/server`) se probó con dos clientes WebSocket externos (difusión, historial, error por mensaje vacío y por sala inexistente). Las claves generadas en Redis (`hexagonal:rooms`, `hexagonal:room:{id}`, `hexagonal:room:{id}:events`) se comprobaron con `redis-cli`.

**Limitación honesta:** en el entorno donde se desarrolló no había Docker instalado, por lo que las imágenes y el `docker-compose.yml` no se construyeron allí; se validó la sintaxis del YAML y el backend se probó de forma nativa contra Redis. La primera ejecución de `docker compose up --build` la deben hacer y registrar ustedes (ver `README.md`).

---

## 6. Lecciones aprendidas

**Técnicas (surgieron al construir el sistema)**

1. **El estilo se paga y se cobra.** Hexagonal agregó archivos, puertos y DTOs, pero permitió escribir pruebas del flujo completo sin infraestructura y una prueba de integración con Redis real sin cambiar el dominio.
2. **Go no tiene sobrecarga de métodos.** `RoomRepository.Save(Room)` y `EventRepository.Save(Event)` no podían vivir en el mismo tipo; se separaron en `Store` y `EventStore`. Los puertos condicionan el diseño de los adaptadores.
3. **Pub/Sub no persiste.** Un mensaje publicado sin suscriptores se pierde. Por eso el historial se guarda aparte en una lista y solo la difusión en vivo usa Pub/Sub.
4. **Un error sutil de Go:** pasar `rdb.Ping(ctx).Err` como función ejecuta el `Ping` una sola vez, por lo que el bucle de reintentos nunca reintentaba. Se corrigió pasando una función anónima. Las pruebas automáticas no lo habrían detectado; sí una revisión del código.
5. **Concurrencia:** cerrar un canal mientras otra goroutine escribe provoca pánico. Se resolvió retirando al cliente del `Hub` (con candado) antes de cerrar su canal, y se verificó con `go test -race`.
6. **`depends_on` no basta:** en Compose hay que esperar a que Redis esté *sano* (`condition: service_healthy`) y además reintentar la conexión en el backend.
7. **Modelar en C4 obliga a separar niveles.** El contexto no debe mencionar tecnología, y la Arquitectura Hexagonal se muestra en el nivel de componentes, no en el de contenedores.

**Qué haríamos distinto con más tiempo**

- Usar **Redis Streams** en lugar de lista + Pub/Sub para tener reproducción y entrega al menos una vez.
- Agregar **autenticación** y validar el `Origin` del WebSocket (hoy acepta cualquier origen, solo apto para entorno académico).
- Persistir los participantes y agregar métricas y trazas (observabilidad).

**Trabajo en equipo (completar entre los tres antes de la sustentación)**

- ¿Cómo repartimos el trabajo y qué funcionó?
- ¿Qué conflictos de Git tuvimos y cómo los resolvimos?
- ¿Qué parte nos costó más y por qué?

---

## 7. Fuentes

**Estilo y tecnologías**
- Cockburn, A. — "Hexagonal Architecture" (alistair.cockburn.us).
- Netflix Tech Blog — "Ready for changes with Hexagonal Architecture": https://netflixtechblog.com/ready-for-changes-with-hexagonal-architecture-b315ec967749
- Martin, R. C. — *Clean Architecture* (2017).
- Documentación oficial: react.dev, vite.dev, typescriptlang.org, go.dev, redis.io; RFC 6455 (WebSocket).
- Brown, S. — C4 Model: https://c4model.com
- Bass, L., Clements, P., Kazman, R. — *Software Architecture in Practice* (catálogo de tácticas).
- Nygard, M. — "Documenting Architecture Decisions" (formato de ADR).

**Mercado laboral (consultadas en septiembre de 2026)**
- Stack Overflow Developer Survey 2025: https://survey.stackoverflow.co/2025/technology y https://survey.stackoverflow.co/2025/work
- State of JS 2025 (front-end frameworks): https://2025.stateofjs.com/en-US/libraries/front-end-frameworks/
- KORE1 — Go Developer Salary Guide 2026: https://www.kore1.com/go-developer-salary-guide/
- Coderhouse — Sueldo Desarrollador Backend en Colombia 2026: https://www.coderhouse.com/co/sueldos/sueldo-desarrollador-backend-colombia-2025
- Indeed Colombia — Sueldo de Backend developer: https://co.indeed.com/career/backend-developer/salaries
- Talently — Salarios de developers en Colombia 2026: https://talently.tech/herramientas/colombia/salario
- Glassdoor — Backend Developer en Medellín: https://www.glassdoor.com/Salaries/medell%C3%ADn-colombia-backend-developer-salary-SRCH_IL.0,17_IM3063_KO18,35.htm
- Levels.fyi — End of Year Pay Report 2025 (citado en la recopilación https://github.com/alihesari/awesome-high-paying-languages).
- TIOBE Index (junio de 2026, dato tomado de una recopilación de terceros; confirmar en https://www.tiobe.com/tiobe-index/).
