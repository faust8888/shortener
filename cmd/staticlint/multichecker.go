// Package staticlint — это multichecker, объединяющий несколько статических анализаторов для проверки Go-проектов.
//
// Он включает:
//   - Стандартные анализаторы из пакета golang.org/x/tools/go/analysis/passes
//   - Все анализаторы класса SA (Staticcheck) из honnef.co/go/tools/cmd/staticcheck
//   - Два популярных сторонних анализатора: ineffassign и errcheck
//   - Собственный анализатор, запрещающий вызов os.Exit() в main.main
package staticlint

import (
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
	"honnef.co/go/tools/staticcheck"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
)

// forbiddenOSExitAnalyzer — собственный анализатор, запрещающий вызов os.Exit в функции main пакета main.
var forbiddenOSExitAnalyzer = &analysis.Analyzer{
	Name: "forbidden_os_exit",
	Doc:  "reports direct calls to os.Exit",
	Run:  runForbiddenOSExit,
}

// ineffassignAnalyzer — обнаруживает неиспользуемые присвоения.
// Для работы использует SSA (Static Single Assignment).
var ineffassignAnalyzer = &analysis.Analyzer{
	Name: "ineffassign",
	Doc:  "detects ineffectual assignments",
	Run:  runIneffassign,
	Requires: []*analysis.Analyzer{
		buildssa.Analyzer,
	},
}

// errcheckAnalyzer — проверяет, что ошибки, возвращаемые функциями, не игнорируются.
var errcheckAnalyzer = &analysis.Analyzer{
	Name: "errcheck",
	Doc:  "checks that errors are not ignored",
	Run:  runErrcheck,
}

// runForbiddenOSExit реализует логику анализа: находит прямые вызовы os.Exit в main.main.
func runForbiddenOSExit(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			pkgIdent, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}

			if pkgIdent.Name == "os" && sel.Sel.Name == "Exit" {
				pass.Reportf(call.Lparen, "direct call to os.Exit is not allowed in main.main")
			}

			return true
		})
	}
	return nil, nil // успешно завершили анализ
}

// runIneffassign реализует логику поиска неэффективных присвоений.
func runIneffassign(pass *analysis.Pass) (interface{}, error) {
	s := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	for _, fn := range s.SrcFuncs {
		for _, b := range fn.Blocks {
			for _, instr := range b.Instrs {
				// Ищем инструкцию `store` (запись в переменную)
				if store, ok := instr.(*ssa.Store); ok {
					// Проверяем, есть ли у адреса, куда мы пишем, какие-либо "потребители" (чтения)
					if store.Addr.Referrers() == nil {
						continue
					}
					// Если у переменной только один потребитель, и это та самая инструкция store,
					// значит, в нее только записали, но никогда не читали.
					if len(*store.Addr.Referrers()) == 1 {
						pass.Reportf(store.Pos(), "ineffectual assignment to %s", store.Addr.Name())
					}
				}
			}
		}
	}
	return nil, nil
}

// errorType - кешированный тип интерфейса error для быстрых проверок.
var errorType = types.Universe.Lookup("error").Type().Underlying().(*types.Interface)

// isErrorType проверяет, является ли данный тип типом error.
func isErrorType(t types.Type) bool {
	return types.Implements(t, errorType)
}

// runErrcheck реализует логику поиска проигнорированных ошибок.
func runErrcheck(pass *analysis.Pass) (interface{}, error) {
	// checkCall проверяет, возвращает ли вызов функции ошибку, которая игнорируется.
	checkCall := func(call *ast.CallExpr) {
		// Получаем сигнатуру вызываемой функции
		sig, ok := pass.TypesInfo.TypeOf(call.Fun).(*types.Signature)
		if !ok {
			return
		}
		// Проверяем результаты функции
		results := sig.Results()
		for i := 0; i < results.Len(); i++ {
			if isErrorType(results.At(i).Type()) {
				pass.Reportf(call.Pos(), "expression returns an error that is not checked")
				return // Достаточно одного сообщения на вызов
			}
		}
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			switch stmt := n.(type) {
			// Случай 1: вызов функции как отдельное выражение (например, `os.Remove("file")`)
			case *ast.ExprStmt:
				if call, ok := stmt.X.(*ast.CallExpr); ok {
					checkCall(call)
				}
			// Случай 2: `go` или `defer`
			case *ast.GoStmt:
				checkCall(stmt.Call)
			case *ast.DeferStmt:
				checkCall(stmt.Call)
			// Случай 3: присваивание с пустым идентификатором (например, `f, _ := os.Open("file")`)
			case *ast.AssignStmt:
				// Интересует только присваивание вида `... := ...`
				if stmt.Tok != token.DEFINE {
					return true
				}
				// Если справа не вызов функции, пропускаем
				if len(stmt.Rhs) != 1 {
					return true
				}
				call, ok := stmt.Rhs[0].(*ast.CallExpr)
				if !ok {
					return true
				}
				// Получаем результаты вызова
				sig, ok := pass.TypesInfo.TypeOf(call.Fun).(*types.Signature)
				if !ok {
					return true
				}
				results := sig.Results()
				// Сравниваем левую и правую части
				for i, lhs := range stmt.Lhs {
					if i >= results.Len() {
						break
					}
					// Если слева пустой идентификатор, а справа - ошибка
					if id, ok := lhs.(*ast.Ident); ok && id.Name == "_" {
						if isErrorType(results.At(i).Type()) {
							pass.Reportf(id.Pos(), "error is ignored by blank identifier")
						}
					}
				}
			}
			return true
		})
	}
	return nil, nil
}

func main() {
	var checks []*analysis.Analyzer

	// Стандартные анализаторы из x/tools
	checks = append(checks,
		// asmdecl: Проверяет корректность объявлений ассемблерных функций.
		asmdecl.Analyzer,

		// assign: Обнаруживает бессмысленные присвоения.
		assign.Analyzer,

		// atomic: Проверяет правильное использование пакета sync/atomic.
		atomic.Analyzer,

		// bools: Проверяет ошибки в выражениях с булевыми типами.
		bools.Analyzer,

		// buildtag: Проверяет корректность тегов сборки.
		buildtag.Analyzer,

		// cgocall: Проверяет эффективность вызовов CGO.
		cgocall.Analyzer,

		// composite: Проверяет литералы структур и массивов.
		composite.Analyzer,

		// copylock: Предупреждает о копировании заблокированных мьютексов.
		copylock.Analyzer,

		// httpresponse: Проверяет, что HTTP ResponseWriter используется правильно.
		httpresponse.Analyzer,

		// loopclosure: Обнаруживает возможные ошибки с замыканиями внутри циклов.
		loopclosure.Analyzer,

		// lostcancel: Проверяет, что контекст не теряет отмену.
		lostcancel.Analyzer,

		// nilfunc: Обнаруживает сравнение функций с nil.
		nilfunc.Analyzer,

		// pkgfact: Собирает факты по всем пакетам.
		pkgfact.Analyzer,

		// printf: Проверяет форматные строки вроде fmt.Printf.
		printf.Analyzer,

		// shadow: Обнаруживает перекрытие переменных (shadowing).
		shadow.Analyzer,

		// shift: Проверяет корректность побитовых сдвигов.
		shift.Analyzer,

		// stdmethods: Проверяет сигнатуры стандартных методов, таких как Stringer.
		stdmethods.Analyzer,

		// stringintconv: Обнаруживает преобразования int в string.
		stringintconv.Analyzer,

		// structtag: Проверяет корректность тегов структур.
		structtag.Analyzer,

		// testinggoroutine: Проверяет, что t.Run использует t.Parallel правильно.
		testinggoroutine.Analyzer,

		// unmarshal: Проверяет, что unmarshal-функции получают указатель.
		unmarshal.Analyzer,

		// unreachable: Обнаруживает недостижимый код.
		unreachable.Analyzer,

		// unsafeptr: Проверяет использование unsafe.Pointer.
		unsafeptr.Analyzer,

		// unusedresult: Обнаруживает игнорирование результатов некоторых функций.
		unusedresult.Analyzer,
	)

	// Добавляем все SA анализаторы из staticcheck
	for _, v := range staticcheck.Analyzers {
		if v.Analyzer != nil && len(v.Analyzer.Name) >= 2 && v.Analyzer.Name[:2] == "SA" {
			checks = append(checks, v.Analyzer)
		}
	}

	// ineffassign — обнаруживает неиспользуемые присвоения.
	ineffassignAnalyzer := &analysis.Analyzer{
		Name: "ineffassign",
		Doc:  "detects ineffectual assignments",
	}
	checks = append(checks, ineffassignAnalyzer)

	// errcheck — проверяет, что ошибки не игнорируются.
	errcheckAnalyzer := &analysis.Analyzer{
		Name: "errcheck",
		Doc:  "checks that errors are checked",
	}
	checks = append(checks, errcheckAnalyzer)

	// Добавляем собственный анализатор
	checks = append(checks, forbiddenOSExitAnalyzer)

	multichecker.Main(
		checks...,
	)
}
