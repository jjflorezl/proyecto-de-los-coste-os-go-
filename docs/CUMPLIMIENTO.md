# Matriz de cumplimiento

Referencia: enunciado «The Hospital de los Costeños con Narcolepsia», 7 páginas, proporcionado por el estudiante. No se redistribuye el PDF dentro del repositorio.

| Sección | Requisito | Implementación / evidencia |
|---|---|---|
| 2 | Go 1.21+, biblioteca estándar | `go.mod`; sin dependencias externas |
| 2 | Modelo separado de `main` | `hospital/`; `main.go` solo coordina escenario y presentación |
| 2 | Campos privados y errores | Constructores y métodos del dominio; `ErrNoBed`, sin `panic` |
| 3 | Hospital y métodos mínimos | `hospital/hospital.go` |
| 3 | Patient y métodos mínimos | `hospital/patient.go` |
| 3 | Doctor y métodos mínimos | `hospital/doctor.go` |
| 3 | Room y métodos mínimos | `hospital/room.go` |
| 3 | EpisodeRecord y Summary | `hospital/episode.go` |
| 3.1 | Constantes tipadas | `PatientState`, `NarcolepsyLevel`, `RoomState` en `states.go` |
| 4.1 | Person embebida | `patient.go`, `doctor.go`, también `orderly.go` |
| 4.2 | Attender con dos tipos | `Doctor` y `Orderly`; `[]Attender` y despacho circular |
| 5.1 | Pacientes en pasillo | `PatientsInHallway() []*Patient` |
| 5.2 | Historial por médico | `MyEpisodes() []EpisodeRecord`, sin parámetro Hospital |
| 5.3 | Primera cama disponible | `AssignRoom(p) (*Room, error)`, sin inventar habitaciones |
| 5.4 | Reporte severo por día | `SevereReport() map[*Patient]int`, fecha COT y ceros incluidos |
| 6 | 2 médicos, 1 camillero, 3 camas, 5 pacientes | `internal/scenario/scenario.go` |
| 6 | Cuatro ataques y falta de cama | `main.go`, salida legible y cuatro episodios |
| 6 | Despertar, liberar y reasignar | `WakeUp()` atómico + `AssignRoom(P-004)` |
| 6 | Imprimir cuatro consultas | Cuatro encabezados explícitos al final de `main.go` |
| 6 | Al menos tres pruebas | `hospital/hospital_test.go`, además pruebas de API y simulación |
| 7 | GUI (+10 %) | `web/static`, `web/server.go`, `cmd/web`; dominio sin código de vista |
| 7 | Concurrencia (+5 %) | `simulation`, una goroutine por paciente y `Hospital.mu` |
| 8.1 | Uso de IA honesto | `AI_USAGE.md`, dos fallos reales y su corrección |
| 8.2 | Preparación para defensa | `docs/SUSTENTACION.md`; el aprendizaje real lo verifica el estudiante |
| 10 | README e integrante | `README.md`, Juan José Flórez López |
| 10 | Historial Git | Repositorio local con commits reales de esta sesión |
| 10 | Enlace remoto de entrega | Pendiente de publicación en la cuenta del estudiante |

Las comprobaciones ejecutadas y sus límites se describen en `VALIDACION.md`. El cumplimiento de la sustentación y la entrega del enlace no puede sustituirse con código ni automatizarse en nombre del estudiante.
