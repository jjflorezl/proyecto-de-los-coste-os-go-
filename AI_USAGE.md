# Uso de inteligencia artificial

**Integrante:** Juan José Flórez López.  
**Fecha de elaboración:** 19 de octubre de 2026, hora de Colombia.  
**Estado:** registro técnico preparado con asistencia de IA; la reflexión personal del estudiante requiere su revisión.

## Herramientas y alcance

Se utilizó **OpenAI Codex** en este chat para leer el enunciado, diseñar el modelo, generar el código Go y sus pruebas, construir la interfaz HTML/CSS/JS, ejecutar comprobaciones y redactar documentación. El asistente también creó los commits locales. Los commits identifican a Codex como autor técnico y no simulan una historia previa del estudiante.

Se usaron las herramientas Go (`gofmt`, `go test`, `go vet`, `go run`), un lector de PDF y herramientas de navegador/terminal para verificar resultados. Se consultó documentación oficial de Go sobre embedding y el detector de carreras. La IA hizo una parte sustancial de la implementación; este trabajo no debe describirse como código escrito sin asistencia.

## Instrucciones decisivas

1. **Solicitud literal del estudiante:** «quiero hacer un trabajo basado en go que trata sobre todo lo del documento adjunto, sigue el paso a paso y crea un proyecto desde 0 con todas las expecificaciones del documento incluido el bonus».
2. **Criterio de implementación derivado del enunciado, formulado por el asistente:** separar el paquete `hospital` de las vistas, usar `Person` embebida y un `[]Attender` con `Doctor` y `Orderly`, y proteger el estado con un mutex.
3. **Criterio de validación formulado por el asistente:** comprobar el escenario 3 camas/4 ataques, el rescate sin duplicar episodios, el límite de día en Colombia y la ejecución concurrente con `-race`.

Solo el primer texto es una cita de un prompt del estudiante. Los otros dos describen decisiones de trabajo de la IA, no mensajes que se atribuyan falsamente a Juan.

## Dos fallos reales del código generado y su corrección

### 1. El camillero no trasladaba a un paciente que ya estaba dormido

La primera implementación de `Orderly.Attend` siempre entraba al flujo de un ataque nuevo, que rechazaba a un paciente en `AsleepInHallway`. La prueba `TestOrderlyRescuesExistingEpisode` falló con «el paciente ya está dormido». Se corrigió `Orderly.Attend` para reconocer el traslado, asignar una cama dentro del mutex y devolver el episodio existente. La prueba comprueba que queda dormido en cama y que el historial conserva un solo episodio.

### 2. La API aceptaba JSON `null`

La primera versión decodificaba directamente en un struct. `encoding/json` acepta `null` sin error al decodificar en ese destino, por lo que `/api/simulation/start` respondía 202. `TestHTTPValidation/null` esperaba 400 y falló. Se corrigió la lectura mediante un `*command`: se rechaza un puntero nulo y solo después se copia el objeto validado. También se rechazan campos desconocidos y contenido JSON adicional.

Estas correcciones las realizó el asistente durante esta sesión. El estudiante debe poder explicarlas, pero no se afirma que las haya detectado o corregido personalmente.

## Incidencia de herramientas

El primer `go test -race ./...` no pudo compilar porque el MinGW instalado falló al arrancar con un error de reubicación de DLL. Se utilizó un LLVM-MinGW portátil en la carpeta de trabajo para repetir la comprobación. Esto es un problema del entorno, distinto de los dos errores de código anteriores. La evidencia final se recoge en `docs/VALIDACION.md`.

## Aprendizaje individual: declaración pendiente de Juan

La IA no puede certificar qué aprendió Juan José Flórez López. Hasta que el estudiante revise y pruebe el proyecto, **no se afirma ningún aprendizaje personal ni preparación completada para la sustentación**.

Antes de entregar, Juan debe sustituir esta sección por su reflexión real. Puede responder en primera persona:

- ¿Cómo reutilicé los métodos de `Person` mediante embedding y por qué eso no es herencia clásica?
- ¿Qué entendí al seguir una llamada desde `RegisterEpisode` hasta `Attender.Attend`?
- ¿Cómo impide el mutex asignar dos veces la última cama y por qué no debe bloquearse otra vez desde un método `Locked`?
- ¿Qué parte pude modificar y comprobar sin ayuda? ¿Qué parte todavía me cuesta explicar?

La guía `docs/SUSTENTACION.md` ayuda a practicar; leerla no sustituye comprender y modificar el programa.
