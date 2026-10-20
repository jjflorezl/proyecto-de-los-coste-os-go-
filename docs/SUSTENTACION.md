# Guía de sustentación individual

Esta guía explica decisiones del proyecto. No afirma que el estudiante ya las domine. Practica cerrándola y explicando el código con tus propias palabras.

## Recorrido sugerido de 15 minutos

1. **Minutos 0–2:** ejecuta `go run .`; explica por qué el cuarto ataque no obtiene cama y por qué al final no queda nadie en pasillo.
2. **Minutos 2–5:** muestra `Person`, `Patient`, `Doctor`, `Room` y `EpisodeRecord`. Explica campos privados, constructores y relaciones.
3. **Minutos 5–7:** sigue `RegisterEpisode` y el `[]Attender`. Muestra el episodio atendido por el camillero.
4. **Minutos 7–10:** explica `AssignRoom`, el mutex y una prueba concurrente. Ejecuta `go test -race ./...` si el entorno está preparado.
5. **Minutos 10–12:** abre la GUI y demuestra que el dominio no depende de ella.
6. **Minutos 12–15:** realiza una modificación pequeña y ejecuta sus pruebas.

## ¿Embedding es herencia?

No. `Patient` contiene un valor `Person` y Go promueve sus métodos. No existe una jerarquía de clases ni un mecanismo de sobrescritura como en Java. La sustitución en este proyecto viene de la interfaz `Attender`.

Con `Person` por valor cada paciente contiene su propia identidad y no hay un puntero nulo adicional. Si fuera `*Person`, varias entidades podrían compartir una identidad mutable y habría que atender la posibilidad de `nil`. Aun con valor, copiar un `Patient` completo sería incorrecto: sus relaciones y pertenencia al hospital representan una entidad única, por eso se usa `*Patient`.

## ¿Qué pasa si cambio un receiver de puntero a valor?

Un método que modifica `Patient` por valor cambia una copia del struct: asignar `state` no modifica el original. Además, copiar un struct que contiene un `sync.Mutex`, como `Hospital`, es incorrecto; `go vet` ayuda a detectar copias del bloqueo. Los getters de `Person` usan receiver por valor porque leen tres campos inmutables y no contienen mutex.

## ¿Qué protege una minúscula?

`state`, `occupants` e `history` no están exportados: código fuera del paquete `hospital` no puede escribir `p.state = ...`. Dentro del paquete sí se puede, por lo que la disciplina del mutex sigue siendo necesaria. Las mayúsculas de los DTO de `Snapshot` permiten leer copias sin exponer el estado interno.

La `Person` embebida es exportada porque así se pide en el ejercicio; no debes reemplazarla ni copiar entidades registradas. La identidad se trata como inmutable, y los métodos del modelo nunca la cambian.

## Dos goroutines compiten por una cama

Las dos pueden llegar a `AssignRoom`. Solo una obtiene `h.mu`. Esa goroutine comprueba la capacidad y ocupa la habitación **sin liberar el mutex entre ambas acciones**. La siguiente ve la capacidad agotada y recibe `ErrNoBed`. Proteger únicamente la escritura no bastaría: el chequeo y la actualización deben ser una sola operación.

`RegisterEpisode` libera el mutex antes de invocar `a.Attend`, porque los implementadores toman el mismo bloqueo. `attendLocked` vuelve a validar el estado después de obtenerlo. Nunca debes llamar a `State()` desde código que ya tiene `h.mu`: el mutex no es reentrante. En un método `Locked` se leen los campos directamente.

## ¿El detector demuestra que no existe ninguna carrera posible?

No. Detecta carreras en las rutas que se ejecutan. Por eso se prueban escrituras, lecturas, reasignaciones, simulación y API concurrente. Su resultado debe acompañarse de una explicación del propietario del estado y del bloqueo, no sustituirla.

## ¿Por qué un map y un slice?

El map por ID evita recorrer todos los pacientes para encontrarlos y facilita rechazar duplicados. Los slices conservan el orden de admisión y de habitaciones para que «la primera disponible» sea determinista. `SevereReport` retorna un map porque la consigna lo exige y representa la asociación paciente → cantidad. Al imprimir se usa el orden del slice, ya que la iteración del map no tiene orden estable.

## ¿Por qué no cambia el conteo al rescatar al paciente?

El evento es el ataque de sueño. Conseguir una cama es un resultado posterior de ese evento. `pendingEpisode` relaciona al paciente con ese registro. `updateEpisodeRoomLocked` actualiza la cama en el hospital y en los historiales de médicos. Nunca se elimina el episodio al despertar.

## Médico tratante, despachador e intervención

`assignedDoctor` es quien tiene al paciente bajo cuidado; `attendingDoctor` es el médico que intervino en un episodio. Un camillero no se registra como médico: su ID y nombre quedan como personal que atendió. Si otro médico rescata después al paciente, su historial también incluye el episodio, mientras el hospital conserva la identidad de quien lo atendió inicialmente.

## Ejercicio: añadir una enfermera

Hazlo en una nueva rama y conserva los tests existentes. Dentro del paquete `hospital`, define `Nurse` con `Person` y `hospital *Hospital`. Implementa un constructor que valide la identidad. `ID` y `Name` se promueven desde `Person`. `Attend` debe bloquear `h.mu`, comprobar que está contratada y delegar en `h.attendLocked(p, location, n, nil)`. La comprobación:

```go
var _ Attender = (*Nurse)(nil)
```

te indica si falta algún método. Contrátala con `h.HireAttender(n)` y agrega una prueba que provoque suficientes episodios para llegar a su turno. No agregues un `case Nurse` al despacho: eso anularía el objetivo del polimorfismo. Mantén la pertenencia al hospital validada en el constructor y en su método de atención.

Para extender además el rescate de pacientes ya dormidos, sigue el patrón de `Orderly.Attend`, sin crear un nuevo episodio por el traslado.

## Otros ejercicios útiles

- Cambia una habitación a capacidad 2 y predice `State()` frente a `IsAvailable()`.
- Añade un paciente severo sin ataques y comprueba que aparece con cero.
- Añade un rechazo explícito para una edad inválida y prueba el mensaje.
- Explica por qué `null` es JSON válido, pero no una solicitud válida para esta API.
- Reemplaza una copia de slice por el slice interno en una rama de práctica y observa qué prueba falla.

## Límites que debes reconocer

Es un ejercicio en memoria, con datos ficticios. No incluye persistencia, cuentas, historial clínico real ni horarios del personal. No debe presentarse como software listo para operar un hospital real. El diseño prioriza una regla de bloqueo fácil de explicar sobre máximo rendimiento: un solo mutex por hospital serializa las operaciones del dominio.
