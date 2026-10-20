# Construcción y uso paso a paso

Autor del trabajo: Juan José Flórez López.

## 1. Preparar Go

Instala Go 1.21 o posterior desde [go.dev](https://go.dev/dl/). Abre una terminal nueva y comprueba `go version`. Este proyecto no necesita dependencias externas: `go.mod` declara el módulo y la versión mínima, y no hay `go.sum` porque no hay módulos descargados.

Para reconstruirlo desde una carpeta vacía, los primeros comandos serían:

```sh
mkdir costenos-narcolepsia
cd costenos-narcolepsia
go mod init costenos-narcolepsia
```

En la carpeta entregada ya se hicieron esos pasos: no hace falta repetirlos. Abre esa carpeta completa en tu editor.

## 2. Entender los estados

Lee `hospital/states.go`. Los tres tipos nombrados evitan mezclar conceptos: un `PatientState` no es un `RoomState`. `String()` devuelve un texto legible para los consumidores.

```text
Awake --ataque--> AsleepInHallway --cama libre--> AsleepInBed
                    |                               |
                    +----------- WakeUp ------------+
                                      |
                                    Awake
```

Si no hay cama, la transición intermedia se conserva y se devuelve `ErrNoBed`. `WakeUp()` también libera la cama en la misma operación. `Room.Release()` por separado deja a un paciente todavía dormido en el pasillo.

## 3. Construir las entidades

Empieza por `Person`, con identidad privada y getters. `Patient`, `Doctor` y `Orderly` incluyen `Person` sin nombre de campo explícito: eso es embedding. Cada entidad mutable tiene un puntero al hospital que protege su estado.

Lee después `Room`: capacidad positiva, lista de ocupantes, métodos de ocupación y liberación. `EpisodeRecord` es una copia de datos históricos; se registra incluso si no hay cama.

## 4. Relacionar las entidades

`Hospital` conserva slices para recorrido ordenado y maps para buscar y rechazar duplicados. Una entidad se construye primero y luego se admite, contrata o registra. Crear un objeto no lo incorpora automáticamente.

`Doctor.DiagnosePatient` mantiene ambos lados de la relación médico-paciente. Cambiar el médico elimina la relación anterior. El historial del médico se guarda dentro del propio médico: `MyEpisodes()` no recibe el hospital.

## 5. Usar la interfaz

Lee `Attender` y las comprobaciones `var _ Attender = (*Doctor)(nil)` y `(*Orderly)(nil)`. Obligan al compilador a comprobar los métodos requeridos. La llamada importante está en `Hospital.RegisterEpisode`:

```go
a := h.staff[h.nextStaff%len(h.staff)]
// ... liberar el mutex antes de delegar ...
_, err := a.Attend(p, location)
```

`a` puede contener un médico o un camillero. El despacho no necesita preguntar el tipo concreto. El camillero no tiene el método de diagnóstico.

## 6. Ejecutar el escenario obligatorio

Ejecuta `go run .`. Sigue `main.go` y `internal/scenario/scenario.go` junto a la salida. Verifica cada paso:

| Momento | Camas ocupadas | Pacientes en pasillo | Episodios |
|---|---:|---:|---:|
| Hospital recién creado | 0 | 0 | 0 |
| Primer ataque | 1 | 0 | 1 |
| Segundo ataque | 2 | 0 | 2 |
| Tercer ataque | 3 | 0 | 3 |
| Cuarto ataque, sin cama | 3 | 1 | 4 |
| P-001 despierta | 2 | 1 | 4 |
| P-004 recibe la cama | 3 | 0 | 4 |

Las cuatro consultas son métodos que devuelven datos. Solo `main` formatea su resultado para la consola.

## 7. Ejecutar las pruebas

```sh
go test ./...
go test -v ./hospital
go vet ./...
gofmt -l .
```

`gofmt -l .` debe devolver una lista vacía. Para aplicar formato, usa `gofmt -w .`.

Lee las pruebas de última cama, cambio de día y copia de slices. Son más útiles que comprobar solamente que un constructor no devuelve `nil`.

## 8. Abrir el bonus gráfico

Ejecuta `go run ./cmd/web` y visita [127.0.0.1:8080](http://127.0.0.1:8080). El HTML y JavaScript se comunican con los handlers del paquete `web`; estos llaman al mismo modelo que la consola. No hay reglas de ocupación de camas implementadas en JavaScript.

La tabla muestra acciones válidas según el último estado recibido. Aun así, el servidor valida nuevamente cada operación: otro usuario o una goroutine pudo cambiar el estado después de dibujar la tabla.

## 9. Probar la simulación

Ejecuta `go run . -simulate -duration 8s`. Observa el escenario obligatorio seguido por estadísticas de la simulación. Repite con `-race` según las instrucciones del README. En la GUI también puedes iniciarla y detenerla.

No agregues pausas dentro del mutex: impedirían progresar a todos los pacientes. Las esperas aleatorias ocurren fuera del modelo, en cada goroutine.

## 10. Preparar la entrega

Revisa `AI_USAGE.md`, practica la guía de sustentación, confirma `git status` y conserva los commits reales. Publica en tu repositorio y entrega su enlace. La fecha y los requisitos administrativos pertenecen al enunciado del curso; esta carpeta no envía automáticamente la entrega a ninguna plataforma.
