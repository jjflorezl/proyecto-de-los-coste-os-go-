# Hospital de los Costeños con Narcolepsia

Proyecto de Programación Orientada a Objetos en **Go 1.21 o posterior**.

**Integrante:** Juan José Flórez López (trabajo individual).  
**Curso:** Programación Orientada a Objetos, Universidad EIA.  
**Docente:** Sebastián Zapata Ramírez.

Implementa el escenario obligatorio del enunciado y ambos bonus: interfaz gráfica mediante REST + HTML/JS (+10 %) y simulación concurrente (+5 %). Los datos son ficticios y se conservan en memoria durante cada ejecución.

## Inicio rápido

Abre una terminal dentro de esta carpeta, donde está `go.mod`. No necesitas instalar módulos, Node, npm ni una base de datos.

```sh
go version
go run .
go test ./...
go vet ./...
```

La consola crea 2 médicos, 1 camillero, 3 habitaciones de capacidad 1 y 5 pacientes. Registra cuatro ataques en lugares distintos; el cuarto falla por falta de cama. P-001 despierta y libera la habitación 101, que recibe P-004. Finalmente imprime las cuatro consultas. El estado final correcto es **cero pacientes en pasillo**, **cuatro episodios** y **tres camas ocupadas**. El reporte severo muestra P-002: 1, P-004: 1 y P-005: 0. Las cifras del ejemplo ilustrativo del PDF no se inventan como episodios adicionales.

## Interfaz gráfica: bonus +10 %

```sh
go run ./cmd/web
```

Abre [la interfaz local](http://127.0.0.1:8080). El servidor comienza con los cinco pacientes despiertos, para que puedas reproducir el escenario manualmente. Cierra con Ctrl+C. Si el puerto está ocupado:

```sh
go run ./cmd/web -addr 127.0.0.1:8081
```

1. En las cuatro primeras filas, pulsa **Registrar episodio**. Puedes cambiar el lugar con el selector.
2. La cuarta operación registra el episodio, pero muestra la falta de cama. La tarjeta de pasillo sube a 1.
3. Pulsa **Despertar** en P-001. Queda una cama libre.
4. Pulsa **Asignar cama** en P-004. No se crea un quinto episodio.
5. Revisa habitaciones, historial por personal y reporte severo.
6. **Iniciar simulación** ejecuta una goroutine por paciente durante 30 segundos. Puedes detenerla antes.

El navegador consulta la API cada segundo. Los archivos estáticos están incluidos en el ejecutable mediante `embed`; no hay CDN ni servicios externos. La lógica clínica no depende de `net/http`, JSON, HTML ni JavaScript. Reiniciar el servidor reinicia los datos; refrescar el navegador los conserva.

## Concurrencia: bonus +5 %

```sh
go run . -simulate -duration 8s -seed 42
go test -race ./...
go run -race .
go run -race . -simulate -duration 8s -seed 42
go run -race ./cmd/web
```

El modo `-simulate` ejecuta primero el escenario obligatorio y después la simulación. Cada paciente espera intervalos aleatorios; si está despierto sufre un ataque, si está en pasillo intenta conseguir cama y si está en cama despierta. `context` cancela, `WaitGroup` espera la terminación y un canal reúne estadísticas. La semilla controla las esperas, pero el orden de ejecución de las goroutines no es determinista.

Para `-race`, Go necesita CGO y un compilador C compatible con la plataforma. En Windows de 64 bits se requiere un toolchain MinGW-w64 compatible; el antiguo MinGW de 32 bits no sirve. Si ya tienes LLVM-MinGW portátil, configura **solo la terminal actual**:

```powershell
$env:PATH = 'C:\ruta\llvm-mingw\bin;' + $env:PATH
$env:CC = 'x86_64-w64-mingw32-gcc'
$env:CGO_ENABLED = '1'
go test -race ./...
go run -race . -simulate -duration 8s
```

Consulta las [instrucciones oficiales del detector](https://go.dev/doc/articles/race_detector) y las [distribuciones de LLVM-MinGW](https://github.com/mstorsjo/llvm-mingw/releases). El compilador C es una herramienta para verificar `-race`, no una dependencia del código hospitalario ni un módulo Go.

## Estructura

```text
main.go                    Escenario obligatorio, flags e impresión
hospital/                  Entidades, reglas, consultas y mutex
  person.go                Identidad compartida por embedding
  patient.go               Estado, ubicación, médico y cama
  doctor.go                Diagnóstico, emergencias e interfaz Attender
  orderly.go               Segundo tipo concreto de Attender
  room.go                  Capacidad, ocupación y liberación
  episode.go               Registro histórico con fecha y hora
  hospital.go              Agregado, despacho y cuatro consultas
  states.go                Constantes tipadas y errores
  snapshot.go              Lectura consistente del modelo
  hospital_test.go         Pruebas de reglas e invariantes
internal/scenario/         Construcción compartida de datos de ejemplo
simulation/                Goroutines, cancelación y estadísticas
cmd/web/main.go            Arranque del servidor y cierre ordenado
web/                       Adaptador REST y sus pruebas
  static/                  HTML, CSS y JavaScript de la vista
docs/                      Guías, trazabilidad y validación
AI_USAGE.md                Uso honesto de IA y correcciones reales
```

## Decisiones del modelo

- **Encapsulación:** los campos del modelo son privados. Los consumidores usan constructores y métodos. Las listas devueltas son copias; los DTO de `Snapshot` son datos separados, no las entidades mutables.
- **Composición:** `Patient`, `Doctor` y `Orderly` embeben `Person` por valor. `ID`, `Name` y `Age` se promueven. La identidad se fija al construir; no se copian entidades ni se reemplaza su `Person` después de registrarlas.
- **Pertenencia:** los constructores reciben `*Hospital`; admitir o contratar es una operación explícita posterior. La pertenencia nunca cambia. Así todas las operaciones de una entidad usan el mismo `sync.Mutex`, sin cambiar mutex durante una simulación.
- **Polimorfismo real:** `Hospital.staff` es `[]Attender` con médicos y camillero. `RegisterEpisode` elige por turnos y llama `a.Attend(...)`. No usa un `switch` de tipos para despachar. El pequeño `switch` de contratación solo conserva las invariantes de pertenencia de los tipos conocidos.
- **Disponibilidad del personal:** la atención es una operación síncrona, sin un tiempo de consulta modelado. Un turno terminado está disponible inmediatamente. El reparto es circular; no se simulan horarios ni turnos ocupados prolongados.
- **Mutex:** `Hospital.mu` protege las colecciones y todo estado mutable de sus entidades. Los métodos privados terminados en `Locked` requieren que el llamador tenga ese bloqueo. No llaman getters que intenten tomarlo otra vez. El hospital no debe copiarse por valor.
- **Habitaciones:** `Occupied` significa que hay al menos un ocupante; `IsAvailable` responde si queda capacidad. En las habitaciones obligatorias de capacidad 1 ambas condiciones son complementarias. Se soportan capacidades mayores sin desperdiciar camas.
- **Errores de negocio:** falta de cama devuelve `ErrNoBed` y conserva al paciente dormido en pasillo y su episodio en el historial. Se identifica con `errors.Is`. Una entrada inválida no crea episodios. No se usa `panic` para estos casos.
- **Historia:** el episodio guarda el lugar original y la habitación asignada. Una asignación posterior actualiza ese dato en el historial del hospital y de los médicos; despertar no borra la cama histórica. Rescatar un paciente no cuenta como un nuevo ataque.
- **Relaciones:** el médico tratante del paciente puede ser distinto del que atiende un episodio. Si el camillero atiende inicialmente, `EpisodeRecord.Doctor()` es `nil` y su identidad queda en `AttenderID/AttenderName`. Un médico que rescata un episodio existente también lo conserva en `MyEpisodes`; el registro central mantiene al despachador inicial.
- **Día del reporte:** se usa UTC-5 (COT, Colombia). Solo se cuentan episodios de la fecha actual en esa zona; los pacientes severos sin ataques aparecen con cero.
- **Slices y maps:** los slices conservan el orden de ingreso, habitaciones y despacho; los maps permiten buscar IDs y verificar duplicados. Se recorre el slice de pacientes al imprimir el reporte, porque Go no garantiza el orden de un map.
- **Separación de la vista:** `Snapshot` es una consulta de datos coherentes bajo un solo bloqueo. La conversión a JSON y a textos visuales está en `web`. La GUI fue añadida en un commit separado que no cambia el paquete `hospital`.

## API local

| Método | Ruta | Entrada / propósito |
|---|---|---|
| GET | `/api/state` | Pacientes, camas, personal, episodios, severidad y simulación |
| POST | `/api/attack` | `{"patient_id":"P-001","location":"Cafetería"}` |
| POST | `/api/wake` | `{"patient_id":"P-001"}` |
| POST | `/api/assign` | `{"patient_id":"P-004"}` |
| POST | `/api/simulation/start` | `{}`; una simulación de 30 segundos |
| POST | `/api/simulation/stop` | `{}`; cancela y espera todas las goroutines |

Los POST requieren `Content-Type: application/json`. Respuestas: 200 éxito; 202 simulación iniciada; 400 entrada inválida; 404 paciente inexistente; 405 método incorrecto; 409 conflicto de estado o falta de cama; 415 tipo de contenido incorrecto. Un 409 con `episode_recorded: true` significa que el ataque sí quedó registrado aunque no haya cama.

## Compilar ejecutables

```sh
go build -o bin/hospital.exe .
go build -o bin/hospital-web.exe ./cmd/web
```

En Linux/macOS puedes omitir `.exe`. No necesitas esos binarios para ejecutar con `go run`.

## Documentación y entrega

- [Paso a paso desde cero](docs/GUIA_PASO_A_PASO.md).
- [Guía para la sustentación individual](docs/SUSTENTACION.md).
- [Matriz de cumplimiento del documento](docs/CUMPLIMIENTO.md).
- [Validación realizada](docs/VALIDACION.md).
- [Declaración de uso de IA](AI_USAGE.md).

El repositorio local contiene commits reales de esta sesión, sin fechas alteradas. **No constituye un enlace remoto de entrega**. Para publicarlo, crea un repositorio vacío en tu cuenta y ejecuta los siguientes comandos usando su URL real:

```sh
git remote add origin URL_DE_TU_REPOSITORIO
git push -u origin main
```

Si ya existe `origin`, revisa `git remote -v` antes de cambiarlo. No ejecutes nuevamente `git init` ni elimines `.git`: ahí está la historia. Antes de entregar, revisa todo el código, completa la reflexión personal de `AI_USAGE.md` con lo que realmente aprendiste y entrega el enlace que solicite el docente.
