# Validación de la entrega

Fecha: 19 de octubre de 2026, Colombia. Entorno: Windows amd64, Go 1.26.5. El módulo declara compatibilidad con Go 1.21 y usa APIs disponibles desde esa versión; no se ejecutó una segunda instalación de Go 1.21.

## Comprobaciones realizadas

| Comando / comprobación | Resultado |
|---|---|
| `go run .` | Correcto: escenario completo, falta de cama controlada y cuatro consultas |
| `go test ./...` | Correcto: paquetes hospital, simulation y web |
| `go vet ./...` | Correcto, sin diagnósticos |
| `gofmt -l .` | Sin archivos pendientes de formato |
| `go test -race ./...` | Correcto con LLVM-MinGW; sin carreras detectadas en las rutas ejecutadas |
| `go run -race . -simulate -duration 3s` | Correcto: escenario más simulación; sin avisos del detector |
| `go build -o ... ./cmd/web` | Ejecutable web compilado correctamente |
| `node --check web/static/app.js` | Sintaxis JavaScript válida |
| Escenario completo a través de HTTP | Probado con `httptest`, incluidos el 409 y la reasignación |
| Archivos `/`, `/app.js`, `/style.css` | Entregados con HTTP 200 en las pruebas |
| Lectores HTTP concurrentes con simulación | Prueba ejecutada con el detector de carreras |
| Exportación limpia de `HEAD` | `go run .`, `go test ./...` y `go vet ./...` correctos, sin archivos locales auxiliares |
| `git fsck --full` | Correcto; integridad de los objetos y del historial local |

Una ejecución de tres segundos produjo 21 ataques nuevos, 9 intentos sin cama y 22 despertares, con 25 episodios históricos en total. Los despertares pueden superar los ataques de la simulación porque esta comienza después del escenario obligatorio, con tres pacientes ya dormidos. Estos números son evidencia de una ejecución, no valores que deban repetirse: el planificador de goroutines introduce variaciones.

## Pruebas incluidas

En `hospital/hospital_test.go`:

- Asignación con cama disponible y registro histórico.
- Falta de cama, error tipado y paciente en pasillo.
- Consulta de pasillo, despertar y reasignación sin duplicar episodios.
- Reporte severo al cruzar medianoche en Colombia, incluyendo cero episodios.
- Despacho polimórfico a médico y camillero.
- Capacidad mayor que uno, liberación y doble liberación inválida.
- Validación de duplicados, entradas, pertenencia y copias de slices.
- Veinte goroutines compitiendo por una cama, con lecturas simultáneas.
- Diagnóstico y rescate por médico.
- Rescate de un episodio existente por el camillero.

En `simulation/simulation_test.go`: ejecución concurrente con cancelación, consistencia de estadísticas, capacidad, ausencia de doble ocupación y configuración inválida.

En `web/server_test.go`: escenario obligatorio mediante handlers HTTP, rechazo de `null`, JSON inválido, campos desconocidos y contenido adicional; métodos, IDs y tipo de contenido; recursos estáticos; inicio único, lecturas concurrentes y detención de simulación.

## Entorno del detector de carreras

El MinGW preinstalado no arrancó por una reubicación de DLL. Se descargó LLVM-MinGW 20261006 UCRT x86_64 desde las publicaciones de su proyecto y se comprobó su SHA-256:

```text
317492c456aa27ee607a5919f1d2d38dcdc1112516a24d0bf4b00d078f52d17a
```

Se configuraron `PATH`, `CC=x86_64-w64-mingw32-gcc` y `CGO_ENABLED=1` solo para el proceso de comprobación. El sandbox impidió al enlazador escribir en su carpeta temporal predeterminada; se repitió con `GOTMPDIR`, `TEMP` y `TMP` apuntando a una carpeta de trabajo autorizada. Después pasaron las pruebas y la simulación con `-race`. No se cambió la configuración global del equipo.

## Límite de la revisión de interfaz

El servidor se compiló y los handlers y recursos se comprobaron con `httptest`. El navegador integrado de este entorno devolvió `ERR_CONNECTION_TIMED_OUT` al intentar abrir el servidor local. Por ello **no se afirma una revisión visual completada ni se entrega una captura como si se hubiera observado la interfaz funcionando**. Para verla en una sesión normal del equipo: `go run ./cmd/web` y abrir `http://127.0.0.1:8080`.

Las pruebas de servidor no sustituyen una prueba visual del navegador. El diseño incluye estilos adaptables, búsqueda, filtros, mensajes de estado y controles accesibles; su apariencia debe revisarse en el navegador del estudiante antes de la entrega.

El intento de `git clone` local encontró una restricción del proceso MSYS de Git en el sandbox (`NtCreateDirectoryObject`, acceso denegado). Se verificó una exportación limpia creada con `git archive HEAD` como alternativa, sin afirmar que el comando de clonación hubiera funcionado en este entorno. El ZIP de entrega conserva además `.git` y sus commits.
