# Uso de inteligencia artificial

**Integrante:** Juan José Flórez López.  
**Fecha de elaboración:** 19 de octubre de 2026, hora de Colombia.  
**Estado:** registro técnico preparado con asistencia de IA; la reflexión personal del estudiante requiere su revisión.

## Herramientas y alcance



Se usaron las herramientas Go (`gofmt`, `go test`, `go vet`, `go run`), un lector de PDF y herramientas de navegador/terminal para verificar resultados. Se consultó documentación oficial de Go sobre embedding y el detector de carreras.


## Instrucciones decisivas


1. **Criterio de implementación derivado del enunciado, formulado por el asistente:** separar el paquete `hospital` de las vistas, usar `Person` embebida y un `[]Attender` con `Doctor` y `Orderly`, y proteger el estado con un mutex.
2. **Criterio de validación formulado por el asistente:** comprobar el escenario 3 camas/4 ataques, el rescate sin duplicar episodios, el límite de día en Colombia y la ejecución concurrente con `-race`.


## Dos fallos reales del código generado y su corrección

### 1. El camillero no trasladaba a un paciente que ya estaba dormido

La primera implementación de `Orderly.Attend` siempre entraba al flujo de un ataque nuevo, que rechazaba a un paciente en `AsleepInHallway`. La prueba `TestOrderlyRescuesExistingEpisode` falló con «el paciente ya está dormido». Se corrigió `Orderly.Attend` para reconocer el traslado, asignar una cama dentro del mutex y devolver el episodio existente. La prueba comprueba que queda dormido en cama y que el historial conserva un solo episodio.

### 2. La API aceptaba JSON `null`

La primera versión decodificaba directamente en un struct. `encoding/json` acepta `null` sin error al decodificar en ese destino, por lo que `/api/simulation/start` respondía 202. `TestHTTPValidation/null` esperaba 400 y falló. Se corrigió la lectura mediante un `*command`: se rechaza un puntero nulo y solo después se copia el objeto validado. También se rechazan campos desconocidos y contenido JSON adicional.

## Incidencia de herramientas

El primer `go test -race ./...` no pudo compilar porque el MinGW instalado falló al arrancar con un error de reubicación de DLL. Se utilizó un LLVM-MinGW portátil en la carpeta de trabajo para repetir la comprobación. Esto es un problema del entorno, distinto de los dos errores de código anteriores. La evidencia final se recoge en `docs/VALIDACION.md`.

## Aprendizaje individual:

se aprenio sobre la composicion en vez de herencia ya que es algo que no experimentabamos en java al ser orientado a objetos, el  polimorfismo que se experimento con doctor y orderly, por ultimo el tema de ser un lenguaje con punteros



- ¿Cómo reutilicé los métodos de `Person` mediante embedding y por qué eso no es herencia clásica?
- ¿Qué entendí al seguir una llamada desde `RegisterEpisode` hasta `Attender.Attend`?
- ¿Cómo impide el mutex asignar dos veces la última cama y por qué no debe bloquearse otra vez desde un método `Locked`?
- ¿Qué parte pude modificar y comprobar sin ayuda? ¿Qué parte todavía me cuesta explicar?

La guía `docs/SUSTENTACION.md` ayuda a practicar; leerla no sustituye comprender y modificar el programa.
