# Go Cheat Sheet — Learn Go with Tests (Go Fundamentals)

Resumen de referencia rápida basado en
[quii.gitbook.io/learn-go-with-tests](https://quii.gitbook.io/learn-go-with-tests/go-fundamentals/install-go),
sección "Go fundamentals" completa (21 capítulos). Pensado para consulta
rápida, no para leer de principio a fin.

Marcadas con ⭐ las secciones más relevantes para el proyecto actual
(middleware/cómputo distribuido): **Concurrency, Select, Sync, Context**.

---

## Índice

1. [Instalación y setup](#1-instalación-y-setup)
2. [Hello World — estructura básica](#2-hello-world--estructura-básica)
3. [Integers — funciones y testing](#3-integers--funciones-y-testing)
4. [Iteration — el único loop de Go](#4-iteration--el-único-loop-de-go)
5. [Arrays y Slices](#5-arrays-y-slices)
6. [Structs, Methods e Interfaces](#6-structs-methods-e-interfaces)
7. [Pointers y Errors](#7-pointers-y-errors)
8. [Maps](#8-maps)
9. [Dependency Injection](#9-dependency-injection)
10. [Mocking](#10-mocking)
11. [⭐ Concurrency — goroutines y channels](#11--concurrency--goroutines-y-channels)
12. [⭐ Select — esperar en múltiples channels](#12--select--esperar-en-múltiples-channels)
13. [Reflection](#13-reflection)
14. [⭐ Sync — Mutex, WaitGroup, atomic](#14--sync--mutex-waitgroup-atomic)
15. [⭐ Context — cancelación y timeouts](#15--context--cancelación-y-timeouts)
16. [Roman Numerals — testing basado en propiedades](#16-roman-numerals--testing-basado-en-propiedades)
17. [Math](#17-math)
18. [Reading Files](#18-reading-files)
19. [HTML Templates](#19-html-templates)
20. [Generics](#20-generics)
21. [Revisiting Arrays/Slices con Generics](#21-revisiting-arraysslices-con-generics)

---

## 1. Instalación y setup

```sh
mkdir my-project
cd my-project
go mod init <modulepath>
```

Genera `go.mod`:
```
module cmd

go 1.24
```

La versión de Go en `go.mod` determina la semántica del lenguaje usada, aunque
tengas instalada una versión más nueva.

```sh
go help mod
go help mod init
```

**Linter recomendado:** [golangci-lint](https://golangci-lint.run)
```sh
brew install golangci-lint
```

**Editor:** necesitas soporte para extraer/inlinear variables y funciones,
renombrar símbolos entre archivos, `go fmt` al guardar, correr tests rápido,
ver firma/definición de funciones, encontrar usos de un símbolo.

---

## 2. Hello World — estructura básica

```go
package main

import "fmt"

func main() {
	fmt.Println("Hello, world")
}
```

Todo programa ejecutable necesita un paquete `main` con una función `main`.

**Funciones:**
```go
func Hello(name string) string {
	return "Hello, " + name
}

// múltiples parámetros del mismo tipo
func Hello(name, greeting string) string {
	return greeting + name
}

// named return values (usar solo cuando el significado no es obvio)
func greetingPrefix(language string) (prefix string) {
	return prefix
}
```

**Tests** (archivo `xxx_test.go`):
```go
package main

import "testing"

func TestHello(t *testing.T) {
	got := Hello("Chris")
	want := "Hello, Chris"
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
```

**Helper de aserción** (usa `testing.TB` para servir tanto a `*testing.T`
como a `*testing.B`):
```go
func assertCorrectMessage(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}
```

**Subtests:**
```go
func TestHello(t *testing.T) {
	t.Run("scenario 1", func(t *testing.T) { /* ... */ })
	t.Run("scenario 2", func(t *testing.T) { /* ... */ })
}
```

**Variables y constantes:**
```go
varName := value   // declaración corta, tipo inferido
var x int = 5      // tipo explícito

const englishHelloPrefix = "Hello, "
const (
	spanish = "Spanish"
	french  = "French"
)
```

**Control de flujo:**
```go
if name == "" {
	name = "World"
}

switch language {
case "Spanish":
	prefix = "Hola, "
case "French":
	prefix = "Bonjour, "
default:
	prefix = "Hello, "
}
```

**Visibilidad:** Mayúscula inicial = exportado/público. Minúscula = privado
al paquete.

**Comandos clave:**

| Comando | Qué hace |
|---|---|
| `go run archivo.go` | Ejecuta directo |
| `go test` | Corre tests del módulo actual |
| `go build` | Compila a binario |
| `go doc fmt` | Ver documentación de un paquete offline |

**Ciclo TDD:** test que falla → código mínimo que compila → mensaje de error
útil → código que pasa → refactor con red de tests.

---

## 3. Integers — funciones y testing

- Cada carpeta es un solo `package`.
- Parámetros del mismo tipo se acortan: `(x, y int)` en vez de `(x int, y int)`.

```go
// Add takes two integers and returns the sum of them.
func Add(x, y int) int {
	return x + y
}
```

**Formato:** `%d` para enteros (no `%q`, que es para strings).

**Testable Examples** — se ejecutan como parte de los tests y validan que la
documentación sea correcta:
```go
func ExampleAdd() {
	sum := Add(1, 5)
	fmt.Println(sum)
	// Output: 6
}
```

---

## 4. Iteration — el único loop de Go

Go **no tiene** `while`, `do`, `until` — solo `for`.

```go
for i := 0; i < 5; i++ {
	// cuerpo
}
```

Requiere llaves, no lleva paréntesis alrededor de las tres partes.

- `var` declara sin inicializar; `:=` declara y asigna a la vez.
- Strings son inmutables — cada concatenación copia memoria.
- Para concatenación repetida, usa `strings.Builder`:
  ```go
  var b strings.Builder
  b.WriteString("hola")
  result := b.String()
  ```

**Benchmarks:**
```go
func BenchmarkRepeat(b *testing.B) {
	for b.Loop() {
		Repeat("a")
	}
}
```
```sh
go test -bench=.
go test -bench=. -benchmem   # + allocs/op, B/op
```

---

## 5. Arrays y Slices

**Arrays** — tamaño fijo, forma parte del tipo (`[4]int` ≠ `[5]int`):
```go
numbers := [5]int{1, 2, 3, 4, 5}
numbers := [...]int{1, 2, 3}   // el compilador infiere el largo
```

**Slices** — tamaño dinámico:
```go
mySlice := []int{1, 2, 3}
mySlice := make([]int, length, capacity)
```

- **Length**: elementos actuales. **Capacity**: máximo antes de reallocar.
- Slicing: `numbers[1:]`, `numbers[low:high]` (high exclusivo).
- `append(slice, value)` devuelve un slice nuevo.
- Comparar slices: `slices.Equal(s1, s2)` (Go 1.21+), nunca `==`.

**Iterar:**
```go
for _, value := range collection {
	// _ ignora el índice
}
```

`len(slice)`, `cap(slice)`.

---

## 6. Structs, Methods e Interfaces

```go
type Rectangle struct {
	Width  float64
	Height float64
}
rectangle.Width // acceso con punto
```

**Methods** — función con receiver, se invocan sobre instancias:
```go
func (r Rectangle) Area() float64 {
	return r.Width * r.Height
}
rectangle.Area()
```

**Interfaces** — la resolución es **implícita**: cualquier tipo con los
métodos correctos satisface la interfaz automáticamente, sin declaración:
```go
type Shape interface {
	Area() float64
}
```

**Formato:** `%f` floats, `%.2f` con 2 decimales, `%g` mayor precisión,
`%#v` struct con nombres de campo.

**Table-driven tests:**
```go
tests := []struct {
	name    string
	shape   Shape
	hasArea float64
}{
	{name: "Rectangle", shape: Rectangle{Width: 12, Height: 6}, hasArea: 72.0},
}
for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) { /* ... */ })
}
```

**Floats:** nunca compares con `==`/`!=` directo (precisión). Usa tolerancia.

---

## 7. Pointers y Errors

Go copia valores al pasarlos a funciones — usa punteros para mutar el original.

```go
func (w *Wallet) Method() { /* receiver puntero */ }
&value   // dirección de memoria
*Type    // tipo puntero
```

Structs con puntero **no requieren** desreferenciar explícitamente:
`w.balance` funciona igual que `(*w).balance`.

**nil:** punteros e interfaces pueden ser `nil`; acceder a un valor `nil`
causa panic en runtime — el compilador no lo detecta.

**Errors** — se devuelven como último valor de retorno:
```go
func (w *Wallet) Withdraw(amount Bitcoin) error {
	if amount > w.balance {
		return ErrInsufficientFunds
	}
	w.balance -= amount
	return nil
}

var ErrInsufficientFunds = errors.New("cannot withdraw, insufficient funds")
```

**Wrapping (Go 1.13+):**
```go
return fmt.Errorf("processing account %s: %w", accountID, err)
```
Comparar con `errors.Is()`, no `==`, para cadenas de errores wrapeados.

**Tipos custom:** `type Bitcoin int` — permite métodos propios e implementar
interfaces (ej. `Stringer`).

---

## 8. Maps

```go
map[keyType]valueType

var dictionary = map[string]string{}
var dictionary = make(map[string]string)   // nunca dejar un map nil
```

La key debe ser comparable. El value puede ser cualquier tipo, incluso otro map.

```go
value := dictionary[key]
definition, ok := dictionary[word]   // ok=false si no existe
dictionary[key] = value              // agregar/actualizar
delete(dictionary, key)
```

**Tipo custom sobre map:**
```go
type Dictionary map[string]string

func (d Dictionary) Search(word string) string {
	return d[word]
}
```

**Error constante:**
```go
type DictionaryErr string
func (e DictionaryErr) Error() string { return string(e) }
const ErrNotFound = DictionaryErr("could not find the word you were looking for")
```

Maps se comportan como referencia — pasarlos a una función no copia los datos.

---

## 9. Dependency Injection

Pasa dependencias como parámetro en vez de crearlas dentro de la función —
sin necesidad de frameworks.

```go
// Antes (no testeable)
func Greet(name string) {
	fmt.Printf("Hello, %s", name)
}

// Después (testeable, reusable)
func Greet(writer io.Writer, name string) {
	fmt.Fprintf(writer, "Hello, %s", name)
}
```

**`io.Writer`** — interfaz fundamental:
```go
type Writer interface {
	Write(p []byte) (n int, err error)
}
```
Implementada por: `os.Stdout`, `bytes.Buffer`, `http.ResponseWriter`, archivos.

**Test:**
```go
buffer := bytes.Buffer{}
Greet(&buffer, "Chris")
got := buffer.String()
```

Acepta interfaces, no tipos concretos, para mantener el código desacoplado.

---

## 10. Mocking

Sustituye dependencias reales por *test doubles*. Un **spy** es un mock que
registra cómo fue usada la dependencia.

```go
type Sleeper interface {
	Sleep()
}

type SpySleeper struct {
	Calls int
}
func (s *SpySleeper) Sleep() { s.Calls++ }
```

Un mismo tipo puede implementar varias interfaces a la vez (`io.Writer` +
`Sleeper`, por ejemplo) con métodos distintos.

**Reglas prácticas:**
- Divide problemas en "thin vertical slices" — software funcional con tests
  antes de refactorizar.
- Testea comportamiento, no detalles de implementación.
- 🚩 Necesitar más de 3 mocks en un test sugiere mal diseño.
- Tests lentos dañan la productividad — mockear habilita feedback rápido.

**Go 1.23 iterators:**
```go
func countDownFrom(from int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := from; i > 0; i-- {
			if !yield(i) {
				return
			}
		}
	}
}

for i := range countDownFrom(3) {
	fmt.Println(i)
}
```

---

## 11. ⭐ Concurrency — goroutines y channels

**Goroutines** — procesos concurrentes ligeros:
```go
go doSomething()
go func() { /* código */ }()
```
> "Una operación que no bloquea en Go corre en un proceso separado llamado
> goroutine."

Las funciones anónimas mantienen acceso al scope léxico externo — útil para
capturar variables al lanzar goroutines en un loop.

**Channels** — comunicación segura entre goroutines:
```go
resultChannel := make(chan result)
resultChannel <- value    // enviar (channel a la izquierda)
value := <-resultChannel  // recibir (channel a la derecha)
```

**Race conditions:** múltiples goroutines escribiendo al mismo map/variable
sin sincronización → crash o corrupción de datos.

```sh
go test -race
```

**Solución típica:** cada goroutine envía su resultado por un channel; el
proceso principal lee secuencialmente del channel (serializa las escrituras).

Con concurrencia bien aplicada en operaciones I/O-bound, se puede lograr
~100x de mejora al eliminar la espera secuencial.

---

## 12. ⭐ Select — esperar en múltiples channels

`select` espera en varias operaciones de channel a la vez; el primer caso
listo "gana" y se ejecuta.

```go
select {
case <-channel1:
	// se ejecuta si channel1 envía
case <-channel2:
	// se ejecuta si channel2 envía
case <-time.After(duration):
	// timeout
}
```

Un channel declarado con `var` (sin `make`) es `nil` y **bloquea para
siempre** — siempre inicializar con `make()`.

```go
ch := make(chan struct{})   // struct{} vacío = eficiente para "señalizar"
go func() {
	// trabajo...
	close(ch)
}()
```

**Patrón "Racer" (dos operaciones compitiendo, con timeout):**
```go
func Racer(a, b string, timeout time.Duration) (string, error) {
	select {
	case <-ping(a):
		return a, nil
	case <-ping(b):
		return b, nil
	case <-time.After(timeout):
		return "", fmt.Errorf("timed out")
	}
}

func ping(url string) chan struct{} {
	ch := make(chan struct{})
	go func() {
		http.Get(url)
		close(ch)
	}()
	return ch
}
```

`time.After()` devuelve un channel que envía tras cierta duración — evita
bloqueo indefinido.

---

## 13. Reflection

**Reflection** — un programa examinando su propia estructura en runtime.

`interface{}` (alias `any`) acepta cualquier tipo, pero pierde ayuda del
compilador, contrato de API claro, y tiene overhead de inspección en runtime.

**Paquete `reflect`:**
```go
v := reflect.ValueOf(x)
v.Kind()        // String, Struct, Slice, Array, Map, Chan, Func, Pointer...
v.NumField()    // # de campos de un struct
v.Field(i)      // campo i
v.String()      // valor subyacente como string
v.Len()         // largo de slice/array
v.Index(i)      // elemento i
v.MapKeys()
v.MapIndex(key)
v.Recv()        // recibir de un channel
v.Call(args)    // invocar función
v.Elem()        // desreferenciar puntero
v.Pointer()     // dirección numérica (detección de ciclos)
```

`reflect.DeepEqual()` — comparar valores complejos en tests.

**Límite conocido:** referencias circulares causan stack overflow — trackear
punteros visitados con `map[uintptr]bool`.

No sobre-diseñes de entrada: construye incrementalmente, empieza por el caso
más simple.

---

## 14. ⭐ Sync — Mutex, WaitGroup, atomic

**Race condition:** operaciones que parecen atómicas (`c.value++`) en
realidad son 3 pasos (leer, incrementar, escribir) — goroutines pueden
intercalarse entre esos pasos y perder actualizaciones.

**`sync.Mutex`:**
```go
type Counter struct {
	mu    sync.Mutex
	value int
}

func (c *Counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}
```

- El valor cero de un `Mutex` ya está desbloqueado (no necesita inicializarse).
- **Un Mutex nunca debe copiarse tras su primer uso** — siempre pasar
  punteros a structs que contengan uno.
- No embeber `sync.Mutex` directo en un struct público — expone `Lock()`/
  `Unlock()` como métodos públicos, generando acoplamiento riesgoso.
- `go vet` detecta copias accidentales de mutex.

**`sync.WaitGroup`:**
```go
var wg sync.WaitGroup
wg.Add(n)     // registrar n goroutines a esperar
// dentro de cada goroutine:
defer wg.Done()
wg.Wait()     // bloquea hasta que todas llamen Done
```

**`sync/atomic`** — alternativa sin locks para un solo valor:
```go
type Counter struct {
	value atomic.Int64
}
func (c *Counter) Inc() {
	c.value.Add(1)
}
```

**Regla práctica:** usa **channels** cuando transfieres *ownership* de un
dato; usa **mutex** para gestionar estado compartido con invariantes entre
varios campos. Para un solo valor protegido, `atomic` suele bastar.

---

## 15. ⭐ Context — cancelación y timeouts

`context.Context` gestiona señales de cancelación y timeouts en procesos de
larga duración (goroutines, handlers de servidor).

> "Cuando un Context se cancela, todos los Contexts derivados de él también
> se cancelan." — estructura de árbol que propaga la cancelación por toda la
> cadena de llamadas.

**Buena práctica (recomendada por Google):** pasar `context.Context` como
primer parámetro en toda la cadena de llamadas, desde el request entrante
hasta las llamadas salientes.

| Función | Qué hace |
|---|---|
| `context.WithCancel(parent)` | contexto derivado + función de cancelación |
| `ctx.Done()` | channel que se cierra cuando el contexto se cancela |
| `ctx.Err()` | razón de la cancelación (ej. `context.Canceled`) |
| `request.Context()` | obtiene el contexto de un `*http.Request` |
| `request.WithContext(ctx)` | adjunta un nuevo contexto a un request |

**Patrón de cancelación con `select`:**
```go
select {
case result := <-dataChan:
	// éxito
case <-ctx.Done():
	return ctx.Err()   // cancelado o timeout
}
```

**Anti-patrón:** usar `ctx.Value()` para control de flujo — no es type-safe,
crea dependencias ocultas. Resérvalo solo para metadata ortogonal (trace ID,
request ID), nunca para pasar parámetros de negocio.

**Handler HTTP típico:**
```go
func Handler(store Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		data, err := store.Fetch(r.Context())
		if err != nil {
			return // contexto cancelado
		}
		fmt.Fprint(w, data)
	}
}
```

---

## 16. Roman Numerals — testing basado en propiedades

**Table-driven tests con subtests nombrados:**
```go
cases := []struct {
	Arabic int
	Roman  string
}{
	{Arabic: 1, Roman: "I"},
	{Arabic: 4, Roman: "IV"},
}
for _, test := range cases {
	t.Run(fmt.Sprintf("%d→%q", test.Arabic, test.Roman), func(t *testing.T) {
		got := ConvertToRoman(test.Arabic)
		if got != test.Roman {
			t.Errorf("got %q, want %q", got, test.Roman)
		}
	})
}
```

`strings.HasPrefix(s, prefix)`, `strings.TrimPrefix(s, prefix)`.

**Reemplazar switch por datos** (desacopla reglas de dominio del algoritmo):
```go
type RomanNumeral struct {
	Value  int
	Symbol string
}
var allRomanNumerals = []RomanNumeral{
	{1000, "M"},
	{900, "CM"},
	{1, "I"},
}
```

**Property-based testing** con `testing/quick` — valida propiedades del
dominio sobre inputs aleatorios en vez de casos fijos:
```go
func TestPropertiesOfConversion(t *testing.T) {
	assertion := func(arabic uint16) bool {
		roman := ConvertToRoman(arabic)
		return ConvertToArabic(roman) == arabic
	}
	if err := quick.Check(assertion, nil); err != nil {
		t.Error("failed checks", err)
	}
}
```

Generación custom de valores (para acotar el rango):
```go
quick.Check(assertion, &quick.Config{
	MaxCount: 1000,
	Values: func(args []reflect.Value, r *rand.Rand) {
		args[0] = reflect.ValueOf(uint16(r.Intn(4000)))
	},
})
```

> "El verdadero reto de las pruebas basadas en propiedades es entender bien
> tu dominio" para poder expresar reglas testeables — no basta con generar
> datos al azar.

---

## 17. Math

**Acceptance tests** primero (definen "condiciones de victoria" a alto
nivel), luego unit tests.

**Paquete `time`:**
```go
tm := time.Date(1337, time.January, 1, 0, 0, 30, 0, time.UTC)
tm.Second()             // int, 0-59
tm.Minute()
tm.Hour()
tm.Format("15:04:05")
```

**Paquete `math`:**
```go
math.Pi
math.Sin(radians)
math.Cos(radians)
math.Abs(value)
```
Trigonometría usa **radianes**, no grados. Círculo completo = 2π.

**Floats:** nunca compares exactos por precisión — define tolerancia:
```go
func roughlyEqualFloat64(a, b float64) bool {
	const equalityThreshold = 1e-7
	return math.Abs(a-b) < equalityThreshold
}
```
División por cero: en compile-time es error; en runtime da `+Inf`.

**Structs son comparables con `==`** si todos sus campos lo son:
```go
type Point struct{ X, Y float64 }
p1 := Point{150, 60}
p2 := Point{150, 60}
p1 == p2 // true
```

**XML:**
```go
type Line struct {
	X1 float64 `xml:"x1,attr"`
}
xml.Unmarshal(b.Bytes(), &svg)
```

**Módulo:** `t.Hour() % 12` — convierte 24h a 12h.

**Filosofía TDD:** "Escribe tests hasta que el miedo se convierta en
aburrimiento." No necesitas ceremonia una vez que tienes confianza.

---

## 18. Reading Files

**`io/fs`** (Go 1.16+) — abstracción de sistema de archivos, desacopla del
filesystem del SO:
```go
type FS interface {
	Open(name string) (File, error)
}
```
Implementaciones: `os.DirFS()`, `embed.FS`, `zip.Reader`,
`testing/fstest.MapFS` (filesystem en memoria para tests).

**Test sin tocar disco:**
```go
fs := fstest.MapFS{
	"hello.md": {Data: []byte("Title: Post 1")},
}
posts := blogposts.NewPostsFromFS(fs)
```

```go
dir, err := fs.ReadDir(fileSystem, ".")
file, err := fileSystem.Open("filename.md")
defer file.Close()

data, err := io.ReadAll(file)

scanner := bufio.NewScanner(file)
for scanner.Scan() {
	line := scanner.Text()
}
```

**Strings:**
```go
strings.TrimPrefix("Title: Hello", "Title: ")
strings.Split("go, rust, c", ", ")
strings.TrimSuffix("hello\n", "\n")
```

**Diseño consumer-driven:** acepta interfaces, no tipos concretos:
```go
// bien — desacoplado
func NewPostsFromFS(fileSystem fs.FS) ([]Post, error)
// mal — acoplado
func NewPostsFromFS(fileSystem fstest.MapFS) ([]Post, error)
```

Archivos de test con sufijo `_test`: `package blogposts_test` — solo accede
a lo exportado, igual que un consumidor real.

---

## 19. HTML Templates

**`html/template`** — escapa datos automáticamente, previene inyección.
Preferir sobre `text/template` para salida HTML. Misma interfaz.

| Sintaxis | Uso |
|---|---|
| `{{.Field}}` | insertar valor de campo |
| `{{range .Items}}...{{end}}` | loop |
| `{{.}}` | valor actual dentro del loop |
| `{{template "name" .}}` | incluir template |
| `{{define "name"}}...{{end}}` | definir sección reusable |
| `{{if .Cond}}...{{end}}` | condicional |
| `{{funcName .Field}}` | llamar función custom |

```go
templ, err := template.New("name").Parse(templateString)
templ.Execute(writer, data)

templ, err := template.ParseFS(embedFS, "templates/*.gohtml")
templ.ExecuteTemplate(writer, "filename.gohtml", data)
```

**Embeber archivos (Go 1.16+):**
```go
import "embed"

//go:embed "templates/*"
var postTemplates embed.FS
```

**Funciones custom:**
```go
templ, err := template.New("index").Funcs(template.FuncMap{
	"sanitiseTitle": func(title string) string {
		return strings.ToLower(strings.Replace(title, " ", "-", -1))
	},
}).Parse(indexTemplate)
```

**Confiar en HTML propio** (evitar el escape automático):
```go
type postViewModel struct {
	HTMLBody template.HTML  // no se escapa
}
```
> "El contenido encapsulado debe venir de una fuente confiable."

**Patrón: renderer reusable** — cachea el template parseado:
```go
type PostRenderer struct {
	templ *template.Template
}
func NewPostRenderer() (*PostRenderer, error) {
	templ, err := template.ParseFS(postTemplates, "templates/*.gohtml")
	return &PostRenderer{templ: templ}, err
}
func (r *PostRenderer) Render(w io.Writer, p Post) error {
	return r.templ.ExecuteTemplate(w, "blog.gohtml", p)
}
```

**Testing:** approval tests en vez de comparar strings frágiles
(`approvals.VerifyString(t, buf.String())`).

---

## 20. Generics

**Type parameters y constraints:**
```go
func AssertEqual[T comparable](t *testing.T, got, want T) {
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}
```
`comparable` permite `==`/`!=`. `any` (alias de `interface{}`) permite
cualquier tipo pero menos operaciones.

**`[T any]` vs `interface{}`:** los generics mantienen type safety — un
mismo call solo trabaja con **un** tipo concreto por invocación; no puedes
mezclar tipos como sí puedes con `interface{}`.

**Struct genérico:**
```go
type Stack[T any] struct {
	values []T
}
func (s *Stack[T]) Push(value T) {
	s.values = append(s.values, value)
}
func (s *Stack[T]) Pop() (T, bool) {
	if s.IsEmpty() {
		var zero T
		return zero, false
	}
	index := len(s.values) - 1
	el := s.values[index]
	s.values = s.values[:index]
	return el, true
}
```

**Inferencia de tipos:**
```go
AssertEqual(t, 1, 1)          // T inferido como int
myStackOfInts := NewStack[int]()  // explícito cuando no se puede inferir
```

**Ventajas sobre `interface{}`:** type safety, sin type assertions, chequeo
en compile-time (vs. `value.(int)` manual con `interface{}`).

**Patrón:** empieza con implementaciones concretas (`StackOfInts`,
`StackOfStrings`) respaldadas por tests; generaliza después de ver el
patrón repetirse — evita abstracción prematura.

---

## 21. Revisiting Arrays/Slices con Generics

**Higher-order functions** — funciones que aceptan otras funciones como
parámetro, para abstraer operaciones sobre colecciones.

**Reduce/Fold** — la abstracción fundamental de "iterar y acumular":
```go
func Reduce[A, B any](collection []A, f func(B, A) B, initialValue B) B {
	var result = initialValue
	for _, x := range collection {
		result = f(result, x)
	}
	return result
}
```

```go
func Sum(numbers []int) int {
	add := func(acc, x int) int { return acc + x }
	return Reduce(numbers, add, 0)
}
```
El valor inicial debe ser el elemento neutro de la operación (0 para suma,
1 para multiplicación) — si no, corrompe el resultado.

**Find:**
```go
func Find[A any](items []A, predicate func(A) bool) (value A, found bool) {
	for _, v := range items {
		if predicate(v) {
			return v, true
		}
	}
	return
}
```

**Buenas prácticas:** usa TDD para detectar oportunidades reales de
refactor antes de abstraer; usa nombres de patrones ya establecidos (`Map`,
`Reduce`, `Filter`) en vez de reinventar terminología; commitea código que
funciona antes de experimentar con refactors.
