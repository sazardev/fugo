(function () {
  "use strict";

  var translations = {
    es: {
      meta: {
        title: "Fugo — Server-Driven UI en Go",
        description: "Fugo: framework local de Server-Driven UI. Escribes toda la lógica en Go; un cliente Flutter precompilado la renderiza. Documentación de proyecto, estado, alcance y seguridad."
      },
      a11y: {
        skip: "Saltar al contenido",
        openNav: "Abrir navegación",
        search: "Buscar / saltar a sección",
        searchTitle: "Buscar (Ctrl K)",
        theme: "Cambiar tema",
        sections: "Secciones",
        paletteLabel: "Paleta de comandos"
      },
      nav: { search: "Buscar", changelog: "Changelog" },
      side: {
        group1: "Proyecto", queEs: "Qué es Fugo", objetivo: "Objetivo", paraQuien: "Para quién", comoFunciona: "Cómo funciona",
        group2: "Estado", estado: "Estado y progreso", alcance: "Alcance", ventajas: "Ventajas y desventajas",
        group3: "Proyecto abierto", seguridad: "Seguridad", openSource: "Código abierto y contribución", licencia: "Licencia",
        group4: "Recursos", contribuir: "Contribuir"
      },
      hero: {
        eyebrow: "Server-Driven UI · Go + Flutter",
        tagline: "Server-Driven UI local para escritorio.<br>Escribe <strong>toda</strong> tu lógica en Go — un cliente Flutter precompilado se encarga de dibujarla.",
        viewGithub: "Ver en GitHub",
        whatIsLink: "Qué es Fugo →"
      },
      queEs: {
        h2: "Qué es Fugo",
        p1: "Fugo es un framework local de <strong>Server-Driven UI (SDUI)</strong>: construyes aplicaciones de escritorio nativas escribiendo <strong>exclusivamente en Go</strong>. La lógica de negocio, el estado y el enrutamiento viven enteramente en un proceso Go; un motor Flutter precompilado actúa como una terminal de renderizado pura — sin lógica de aplicación, sin estado propio.",
        p2: "Los dos procesos se comunican por <strong>gRPC</strong> con streaming bidireccional, sobre <strong>Unix Domain Sockets</strong> (TCP en Windows) y <strong>Protocol Buffers</strong> como formato de cable.",
        diagramLabel: "Diagrama de arquitectura",
        diagram: "┌──────────────────────┐     IPC (UDS/TCP)    ┌──────────────────────┐\n│      Proceso Go      │◄════════════════════►│   Proceso Flutter    │\n│                      │   gRPC + Protobuf     │                      │\n│  Lógica de negocio    │                       │  Registro de widgets │\n│  Árbol retenido       │──── Diff de árbol ───►│  Pipeline de render  │\n│  Motor de diffing     │                       │  Debounce de eventos │\n│  Servidor gRPC        │◄──── Eventos de UI ───│  Cliente gRPC        │\n└──────────────────────┘                       └──────────────────────┘",
        callout: "<strong>Go es la fuente de verdad absoluta.</strong> Flutter es una terminal muda: sin lógica de negocio, sin estado — solo píxeles a 60/120&nbsp;fps vía Impeller."
      },
      objetivo: {
        h2: "Objetivo",
        p1: "Fugo busca ser el framework <strong>\"pilas incluidas\"</strong> para construir aplicaciones de escritorio ultrarrápidas y eficientes usando <strong>únicamente Go</strong>: aplicaciones corriendo en segundos, combinando la velocidad de iteración y el conocimiento que ya tienes de Go con el poder de renderizado y el ecosistema visual de Flutter — sin tener que aprender Dart, sin gestionar estado en dos lenguajes, sin el consumo de memoria de un motor basado en Chromium.",
        p2: "No es un puente genérico Go↔Dart ni un binding de FFI: es una arquitectura deliberada de <em>terminal tonta</em> donde el 100% de la superficie de programación que tocas es Go — <code>fg.Text</code>, <code>fg.Button</code>, <code>fg.Router</code> — y Flutter nunca ve una decisión de negocio, solo un árbol de widgets y sus parches."
      },
      paraQuien: {
        h2: "Para quién",
        c1h: "Desarrolladores Go", c1p: "Que quieren una GUI de escritorio nativa sin salir de Go ni aprender Dart/Flutter para poder usarlo.",
        c2h: "Herramientas internas", c2p: "Equipos que ya tienen backends/CLIs en Go y necesitan una interfaz de escritorio ligera para operarlos, sin el peso de Electron.",
        c3h: "Apps de un solo lenguaje", c3p: "Proyectos donde mantener <em>un único</em> lenguaje de aplicación (Go) es una prioridad de arquitectura y de equipo, no una preferencia estética."
      },
      comoFunciona: {
        h2: "Cómo funciona",
        p1: "El árbol de widgets se construye <strong>una sola vez</strong> y se <strong>retiene</strong>. Los manejadores de eventos son closures de Go que mutan los campos del widget directamente (<code>counterText.SetText(\"5\")</code>) y luego llaman a <code>ctx.Update()</code>. El scheduler (ticker de 60&nbsp;fps) vuelve a recorrer ese <em>mismo</em> árbol cada frame cuando hay cambios pendientes, re-serializa las props, calcula el diff contra el snapshot anterior y transmite <strong>solo los parches</strong>.",
        p2: "Esto <strong>no</strong> es un \"rebuild desde cero\" al estilo React — es mutación directa de un árbol retenido, con diffing por ID/posición (<code>CREATE</code>/<code>UPDATE</code>/<code>DELETE</code>/<code>REPLACE</code>/<code>REORDER</code>) y streaming incremental sobre gRPC."
      },
      estado: {
        h2: "Estado y progreso",
        intro: "<span class=\"badge badge-flame\">v3.44.8-fugo.0</span> — el motor, la API de widgets, el transporte, el supervisor de proceso, el CLI y el cliente Flutter están implementados y funcionan de extremo a extremo.",
        li1: "Instalable vía <code>go install github.com/sazardev/fugo/cmd/fugo@latest</code> — los bindings de protobuf están commiteados, compila en un fetch limpio",
        li2: "Material&nbsp;3 nativo (claro por defecto), variantes de botón, Card/Scaffold/FAB/ListTile/Chip/Progress",
        li3: "Motor de diffing, reconciliador, scheduler de 60&nbsp;fps con prioridad (<code>Update</code> / <code>UpdateNow</code>)",
        li4: "Transporte gRPC (UDS / TCP en Windows), health check, keepalive, token de autenticación opcional",
        li5: "83 widgets en <code>fg/</code> con API fluida y prefix-free, un sistema de <code>Theme</code>, y un <code>fg.Store[S]</code> opcional para estado compartido",
        li6: "Cliente Flutter (isolate gRPC en segundo plano, registro de widgets, reconexión automática)",
        li7: "CLI: <code>fugo init</code> / <code>run</code> (hot reload por defecto) / <code>build</code> / <code>doctor</code> (<code>--fix</code>) / <code>widgets</code> / <code>upgrade</code>",
        li8: "Control de ventana en tiempo de ejecución, portapapeles, diálogos de archivo nativos",
        li9: "Overlays imperativos: <code>ShowSnackBar</code>, <code>ShowDialog</code>, <code>ShowBottomSheet</code>, selectores de fecha/hora",
        li10: "Rendimiento: diffing con pool de objetos, ajuste de GC, benchmarks Go + Dart con perf gate en CI",
        outro: "Ver <a href=\"https://github.com/sazardev/fugo/tree/main/ROADMAP\" target=\"_blank\" rel=\"noopener\">ROADMAP</a> y <a href=\"https://github.com/sazardev/fugo/blob/main/SPEC.md\" target=\"_blank\" rel=\"noopener\">SPEC.md</a> para la visión de diseño completa."
      },
      alcance: {
        h2: "Alcance",
        inH: "Dentro del alcance",
        in1: "Apps de escritorio: Linux, Windows, macOS",
        in2: "IPC estrictamente local (UDS/TCP en <code>127.0.0.1</code>)",
        in3: "Un binario Go + un cliente Flutter precompilado por plataforma",
        in4: "Material&nbsp;3 como sistema de diseño del cliente de render",
        outH: "Fuera del alcance (por ahora)",
        out1: "Móvil (iOS/Android) y web",
        out2: "SDUI remoto sobre red (el diseño asume IPC local de baja latencia, no una WAN)",
        out3: "Persistencia de estado entre hot reloads (el estado en memoria se reinicia)",
        out4: "Un sistema de diseño propio — el cliente es intencionalmente Material&nbsp;3 \"vanilla\""
      },
      pros: {
        h2: "Ventajas y desventajas",
        prosH: "Ventajas",
        p1: "Sin el consumo de memoria de Electron/Chromium — renderizado nativo vía Flutter/Impeller",
        p2: "Un único lenguaje de aplicación: toda la lógica, el estado y el enrutamiento en Go",
        p3: "Material&nbsp;3 real (tipografía, layout, animaciones) sin reinventar un design system",
        p4: "Latencia de IPC local de 5–10&micro;s, no 50–200&nbsp;ms de un SDUI remoto",
        p5: "Solo se transmiten diffs — el framing en Protobuf evita el costo de parsear JSON en cada frame",
        p6: "CLI \"pilas incluidas\": <code>init</code> → <code>run</code> (hot reload) → <code>build</code> → <code>doctor</code>",
        consH: "Desventajas",
        c1: "Pre-1.0: la API pública puede tener cambios incompatibles entre versiones",
        c2: "<code>fugo run</code>/<code>build</code> requieren el SDK de Flutter instalado (o un binario precompilado vía <code>FUGO_FLUTTER_BINARY</code>)",
        c3: "Sin objetivo móvil o web todavía — desktop-only por diseño actual",
        c4: "El hot reload no persiste el estado en memoria entre recargas",
        c5: "Ecosistema y comunidad jóvenes — menos respuestas de terceros que con Flutter/Dart puro",
        c6: "Estás limitado a la superficie que expone <code>fg/</code>, no al 100% de los widgets de Flutter/Dart"
      },
      seguridad: {
        h2: "Seguridad",
        p1: "El transporte es <strong>local por defecto</strong> — Unix Domain Sockets (o TCP en <code>127.0.0.1</code> en Windows) — no expuesto a la red salvo que el usuario lo configure explícitamente. <code>FUGO_AUTH=1</code> añade un token por ejecución que endurece aún más el transporte local.",
        c1h: "Análisis estático", c1p: "<a href=\"https://codeql.github.com/\" target=\"_blank\" rel=\"noopener\">CodeQL</a> corre en cada push/PR y semanalmente vía <code>codeql.yml</code>.",
        c2h: "Reporte responsable", c2p: "Vulnerabilidades vía <a href=\"https://github.com/sazardev/fugo/security/advisories/new\" target=\"_blank\" rel=\"noopener\">security advisories privados</a> — nunca en issues públicos.",
        c3h: "Compromiso de respuesta", c3p: "Confirmación en 48h, actualizaciones cada 7 días, crédito en el advisory y en el CHANGELOG.",
        p2: "Ver <a href=\"https://github.com/sazardev/fugo/blob/main/SECURITY.md\" target=\"_blank\" rel=\"noopener\">SECURITY.md</a> para la política completa, incluyendo versiones soportadas."
      },
      openSource: {
        h2: "Código abierto y contribución",
        p1: "Fugo es <strong>abierto a contribuciones</strong> bajo licencia MIT. El flujo de trabajo está documentado en <a href=\"https://github.com/sazardev/fugo/blob/main/CONTRIBUTING.md\" target=\"_blank\" rel=\"noopener\">CONTRIBUTING.md</a> y el proyecto sigue el <a href=\"https://github.com/sazardev/fugo/blob/main/CODE_OF_CONDUCT.md\" target=\"_blank\" rel=\"noopener\">Contributor Covenant</a>.",
        li1: "Plantillas de issue estructuradas (bug report / feature request) y de pull request",
        li2: "<code>CODEOWNERS</code> define la revisión requerida en todo el repositorio",
        li3: "<code>.golangci.yml</code>: 80+ linters activos, sin umbral de tolerancia (<code>max-issues-per-linter: 0</code>)",
        li4: "Git hooks (Lefthook): pre-commit auto-arregla formato/lint; pre-push corre la puerta completa",
        li5: "Dependabot semanal para módulos Go y GitHub Actions"
      },
      cicd: {
        p1: "Todo push y pull request a <code>main</code> corre la misma puerta de calidad — nada se fusiona sin pasarla:",
        hLint: "Lint", pLint: "golangci-lint + staticcheck",
        hVet: "Vet", pVet: "<code>go vet ./...</code>",
        hBuild: "Build", pBuild: "<code>go build ./...</code>",
        hTest: "Test", pTest: "<code>go test -race -shuffle=on</code>",
        hFormat: "Format", format: "gofumpt, sin diferencias",
        hBench: "Bench", bench: "Perf gates + benchmarks del motor de diffing",
        hDart: "Dart", dart: "<code>flutter analyze</code> + <code>flutter test</code> en <code>flutter_client/</code>",
        hCodeQL: "CodeQL", codeql: "Análisis de seguridad, push/PR + semanal",
        p2: "Aparte, <code>check-flutter-version.yml</code> vigila semanalmente nuevas versiones estables de Flutter y abre un PR (nunca auto-merge) para actualizar el pin, y <code>release-flutter-client.yml</code> compila el cliente Flutter para Linux/Windows y lo adjunta al release en cada tag <code>v*</code>.",
        callout: "Esta misma página se despliega automáticamente: un workflow de GitHub Actions publica <code>site/</code> en GitHub Pages en cada cambio a <code>main</code>, sin pasos manuales."
      },
      licencia: {
        h2: "Licencia",
        p: "<strong>MIT</strong> — permisiva, sin restricciones de uso comercial ni de distribución. Ver <a href=\"https://github.com/sazardev/fugo/blob/main/LICENSE\" target=\"_blank\" rel=\"noopener\">LICENSE</a> en el repositorio."
      },
      footer: {
        p1: "Fugo — hecho con Go + Flutter. <a href=\"https://github.com/sazardev/fugo\" target=\"_blank\" rel=\"noopener\">github.com/sazardev/fugo</a>",
        p2: "La guía completa del framework (arquitectura, catálogo de widgets, theming, estado) vive en docs/FUGO_IDIOMATICO.md — ver el enlace \"Guía completa\" arriba.",
        colProject: "Proyecto", repo: "Repositorio", issues: "Issues", prs: "Pull requests", discussions: "Discussions",
        colDocs: "Documentación", guide: "Guía completa",
        colSupport: "Apoyar", star: "★ Star en GitHub", releases: "Releases"
      },
      palette: {
        placeholder: "Ir a una sección… (Home, Seguridad, CI/CD…)",
        groupSection: "Sección", groupLink: "Enlace", groupAction: "Acción", empty: "Sin resultados",
        openGithub: "Abrir en GitHub", viewChangelog: "Ver Changelog", toggleTheme: "Cambiar tema claro/oscuro"
      },

      contribMeta: {
        title: "Contribuir — Fugo",
        description: "Cómo configurar el entorno, el estilo de código, los tests y el flujo de pull request para contribuir a Fugo."
      },
      contrib: {
        eyebrow: "Código abierto · MIT",
        group1: "En esta página",
        sAntes: "Antes de empezar", sSetup: "Entorno de desarrollo", sEstilo: "Estilo de código", sTests: "Tests",
        sFlujo: "Commit & pull request", sReleases: "Releases", sReportar: "Reportar bugs / seguridad", sConducta: "Código de conducta",
        backHome: "← Volver al inicio",
        h1: "Contribuir a Fugo",
        lead: "Fugo es un proyecto abierto bajo licencia MIT. Esto es todo lo que necesitas para configurar el entorno, seguir el estilo de código y abrir un pull request que pase la puerta de CI a la primera.",
        fork: "Hacer fork en GitHub", issues: "Ver issues abiertos",
        antesH: "Antes de empezar",
        antesP1: "Lee primero <code>CLAUDE.md</code> en la raíz del repositorio — es el documento de arquitectura canónico (paquetes, flujo de datos, formato de cable, convenciones) y se mantiene al día con el código, a diferencia de algunos documentos de diseño más antiguos en español bajo <code>ROADMAP/</code>.",
        antesP2: "Para cualquier cambio no trivial (un widget nuevo, un cambio de transporte, un comando de CLI), abre primero un issue o discussion — ahorra retrabajo si el enfoque necesita ajustes. Fixes pequeños (typos, docs, un bug claro con test de regresión) pueden ir directo a un PR.",
        setupH: "Entorno de desarrollo",
        setupP1: "Requiere <strong>Go 1.26+</strong>. Los bindings de protobuf generados están commiteados, así que <code>go build ./...</code> funciona de inmediato sin <code>protoc</code>.",
        setupP2: "Si vas a tocar <code>transport/proto/fugo/v1/fugo.proto</code>, además necesitas <code>protoc</code>, <code>protoc-gen-go</code>, <code>protoc-gen-go-grpc</code> y <code>protoc-gen-dart</code> (el plugin de Dart debe ser <code>protoc_plugin</code> 21.1.x — ver <code>CLAUDE.md</code>).",
        setupP3: "Si vas a tocar <code>flutter_client/</code> (el cliente Dart), necesitas el SDK de Flutter fijado en <code>FLUTTER_VERSION</code> — ver <code>flutter_client/README.md</code>.",
        estiloH: "Estilo de código",
        estilo1: "El formateador es <code>gofumpt</code>, no <code>gofmt</code> — corre <code>gofumpt -w .</code> antes de commitear (Lefthook lo hace automáticamente en los archivos en stage)",
        estilo2: "<code>make lint</code> corre <code>golangci-lint</code> (80+ linters en <code>.golangci.yml</code>) y <code>staticcheck</code> — ambos deben pasar limpio, sin umbral de tolerancia",
        estilo3: "<code>make vet</code> (<code>go vet ./...</code>) debe pasar",
        estilo4: "Cambios acotados: un bug fix no debería traer un refactor sin relación. Sin abstracciones especulativas para necesidades hipotéticas futuras",
        estilo5: "Los widgets siguen las convenciones de <code>fg/</code>: constructores sin prefijo (<code>fg.Text</code>, no <code>NewText</code>), setters encadenables, una bandera <code>*Set bool</code> para props opcionales",
        testsH: "Tests",
        testsP1: "<code>make test</code> corre la suite completa con <code>-race -shuffle=on</code>. El código nuevo necesita cobertura de tests; los bug fixes necesitan un test de regresión que falle antes del fix y pase después.",
        testsP2: "Los tests table-driven son el estilo de la casa en el lado Go. Los benchmarks viven en <code>*_bench_test.go</code> (<code>engine/</code> tiene las puertas de rendimiento que CI verifica en cada push — no las regresiones sin una buena razón documentada en el PR).",
        flujoH: "Commit & pull request",
        paso1: "Crea una rama a partir de <code>main</code>.",
        paso2: "Haz tu cambio, con commits acotados.",
        paso3: "<code>gofumpt -w .</code>, <code>make lint</code>, <code>make vet</code>, <code>make test</code>, <code>go mod tidy</code> — todo limpio.",
        paso4: "Actualiza <code>CHANGELOG.md</code> bajo <code>[Unreleased]</code> si el cambio es visible para el usuario.",
        paso5: "Abre el PR — <code>.github/pull_request_template.md</code> te pedirá el checklist anterior. <code>make pr MSG=\"type: description\"</code> automatiza el branch/push/apertura del PR si estás trabajando directo desde <code>main</code>.",
        flujoP: "CI corre lint, vet, build, test (<code>-race -shuffle=on</code>), un chequeo de formato con gofumpt, el perf gate/benchmarks, y el job de Dart (<code>flutter analyze</code> + <code>flutter test</code>) en cada push y PR — todo debe pasar antes de mergear. CodeQL corre semanalmente y en cada push para análisis de seguridad.",
        releasesH: "Releases",
        releasesP: "No edites a mano <code>VERSION</code>, <code>FLUTTER_VERSION</code>, ni <code>CHANGELOG.md</code> — son responsabilidad de quien mantiene el proyecto vía <code>make release</code> / <code>make release-flutter</code>. Ver el <a href=\"changelog.html#versionamiento\">esquema de versionamiento</a> para los detalles.",
        releasesLink: "esquema de versionamiento",
        reportarH: "Reportar bugs / seguridad",
        reportarP1: "Usa las <a href=\"https://github.com/sazardev/fugo/issues/new/choose\" target=\"_blank\" rel=\"noopener\">plantillas de issue</a> — piden el contexto correcto de entrada (pasos de reproducción, versiones de Go/Flutter, comportamiento esperado vs. real).",
        reportarP2: "Para vulnerabilidades de seguridad usa un <a href=\"https://github.com/sazardev/fugo/security/advisories/new\" target=\"_blank\" rel=\"noopener\">security advisory privado</a> en lugar de un issue público — ver <a href=\"https://github.com/sazardev/fugo/blob/main/SECURITY.md\" target=\"_blank\" rel=\"noopener\">SECURITY.md</a>.",
        conductaH: "Código de conducta",
        conductaP: "Este proyecto sigue el <a href=\"https://github.com/sazardev/fugo/blob/main/CODE_OF_CONDUCT.md\" target=\"_blank\" rel=\"noopener\">Contributor Covenant</a>. Al participar, se espera que lo respetes."
      },

      chMeta: {
        title: "Changelog — Fugo",
        description: "Historial completo de versiones de Fugo y el esquema de versionamiento ligado a Flutter."
      },
      ch: {
        eyebrow: "Historial de versiones",
        sVersion: "Versionamiento", sHistorial: "Historial de versiones",
        h1: "Changelog",
        lead: "Historial completo de Fugo, versión por versión — traducido del <code>CHANGELOG.md</code> del repositorio.",
        viewSource: "Ver CHANGELOG.md en GitHub", viewReleases: "Ver releases",
        unreleased: "Sin publicar",
        grpAdded: "Añadido", grpChanged: "Cambiado", grpFixed: "Corregido", grpInfra: "Infraestructura",
        versionH: "Versionamiento",
        versionP1: "La versión de Fugo sigue intencionalmente a Flutter: <code>VERSION</code> es <code>&lt;versión-de-flutter&gt;-fugo.&lt;build&gt;</code> (por ejemplo <code>3.44.8-fugo.0</code>) — los primeros tres números siempre son la versión exacta del SDK de Flutter contra la que están compilados el cliente precompilado y <code>flutter_client/</code>, así que nunca hay ambigüedad entre un release de Fugo y el Flutter con el que renderiza.",
        versionP2: "<code>FLUTTER_VERSION</code> es la única fuente de verdad para ese pin: <code>release-flutter-client.yml</code> compila exactamente contra esa versión (no un canal flotante como <code>stable</code>), y <code>fugo doctor</code> avisa (no falla — el SDK de Flutter es opcional) si tu Flutter instalado no coincide.",
        versionP3: "Un workflow semanal, <code>check-flutter-version.yml</code>, revisa si hay una versión estable de Flutter más reciente y abre un PR para actualizar el pin — Fugo está pensado para seguir a Flutter de forma continua, no quedarse atrás. Ese PR nunca se mergea automáticamente, porque una actualización de Flutter puede romper el código Dart de <code>flutter_client/</code>.",
        versionP4: "Dos comandos cubren un release, ninguno edita <code>VERSION</code>/<code>CHANGELOG.md</code> a mano: <code>make release MSG=\"...\"</code> sube solo el sufijo <code>-fugo.N</code> (un fix de Fugo contra la misma versión de Flutter); <code>make release-flutter FLUTTER=&lt;version&gt; MSG=\"...\"</code> sube la versión de Flutter objetivo (reinicia a <code>-fugo.0</code>). Ya no existe <code>TYPE=patch|minor|major</code> — mayor/menor no significan nada bajo este esquema.",
        historialH: "Historial de versiones",
        historialLead: "Formato basado en <a href=\"https://keepachangelog.com/en/1.1.0/\" target=\"_blank\" rel=\"noopener\">Keep a Changelog</a>. Traducido del <a href=\"https://github.com/sazardev/fugo/blob/main/CHANGELOG.md\" target=\"_blank\" rel=\"noopener\">CHANGELOG.md</a> original en inglés.",
        hist: {
          r2: {
            c1: "<strong>La versión de Fugo ahora sigue exactamente la versión del SDK de Flutter que usa.</strong> <code>VERSION</code> es <code>&lt;versión-de-flutter&gt;-fugo.&lt;build&gt;</code> (ej. <code>3.44.8-fugo.0</code>) — los primeros tres números siempre indican exactamente contra qué versión de Flutter están compilados el cliente precompilado y <code>flutter_client/</code>, así que nunca hay ambigüedad entre un release de Fugo y el Flutter con el que renderiza. <code>FLUTTER_VERSION</code> es la nueva fuente única de verdad para ese pin: CI (<code>release-flutter-client.yml</code>) compila exactamente contra esa versión en vez de un canal flotante <code>channel: stable</code>, y <code>fugo doctor</code> avisa si tu SDK de Flutter instalado no coincide.",
            c2: "<code>make release MSG=\"...\"</code> ahora solo sube el sufijo de build <code>-fugo.N</code> (un fix de Fugo contra la misma versión de Flutter) — <code>TYPE=patch|minor|major</code> desapareció, ya que mayor/menor no significan nada bajo este esquema. Subir la versión de Flutter objetivo es un paso separado y explícito: <code>make release-flutter FLUTTER=&lt;version&gt; MSG=\"...\"</code>.",
            c3: "Un nuevo workflow programado (<code>check-flutter-version.yml</code>) revisa semanalmente si hay una versión estable de Flutter más nueva y abre un PR que sube <code>FLUTTER_VERSION</code> — Fugo está pensado para seguir a Flutter actual, no quedarse atrás. Ese PR siempre necesita revisión y merge humano (una actualización de Flutter puede afectar el código Dart de <code>flutter_client/</code>), nunca se auto-mergea."
          },
          r3: {
            a1: "<code>fugo doctor</code> ahora verifica <strong>coherencia del proyecto</strong>: lee la ruta del módulo de <code>go.mod</code> y el import <code>&lt;module&gt;/ui</code> de <code>main.go</code>, y reporta un ✗ preciso cuando no coinciden, cuando falta el paquete <code>ui</code> importado, o cuando no tiene un <code>Build</code> exportado — detectado antes del error de compilación genérico.",
            a2: "<code>fugo doctor --fix</code> repara las partes auto-corregibles dentro de un proyecto antes de diagnosticar: escribe un <code>fugo.toml</code> por defecto si falta, corre <code>git init</code> si no hay repo, y <code>go mod tidy</code> para resolver dependencias.",
            c1: "<strong><code>fugo run</code> hace hot-reload por defecto.</strong> Ahora vigila los archivos <code>.go</code> y reconstruye/reconecta el servidor Go en cada cambio mientras la ventana de Flutter permanece abierta — así que los cambios en texto, handlers, layout, etc. se ven en vivo (el estado en memoria se reinicia entre recargas). Pasa <code>--no-watch</code> para un solo build-and-run; el antiguo flag <code>--watch</code> ahora está implícito (oculto, no-op).",
            f1: "Cliente de render Flutter: la reconexión tras un reinicio del servidor (hot reload) ya no falla con \"Bad state: Stream has already been listened to\". El isolate ahora se suscribe a su <code>ReceivePort</code> una sola vez y enruta eventos al stream activo, y reinicia su backoff tras una sesión exitosa para reconectar en ~500ms."
          },
          r4: {
            a1: "<code>fugo doctor</code> ahora es un chequeo de salud en dos partes. <strong>Toolchain</strong>: Go, Flutter, git, protoc, gofumpt (Go/Flutter son requeridos → ✗; el resto son advertencias) más la plataforma. <strong>Proyecto</strong> (al correr dentro de un proyecto Fugo): importa <code>fugo.toml</code> y reporta la ventana/dirección resuelta, verifica que la dirección gRPC configurada esté libre (o sea un unix socket), verifica la estructura (<code>main.go</code>, <code>ui/</code>, <code>go.mod</code>), que el módulo <code>github.com/sazardev/fugo</code> resuelva, si el cliente Flutter está compilado (o es compilable), y finalmente compila el proyecto (<code>go build ./...</code>).",
            c1: "<code>fugo doctor</code> ahora sale con código distinto de cero cuando hay un problema bloqueante (falta una herramienta requerida, un módulo sin resolver, o un proyecto que no compila), para poder usarlo como gate en scripts/CI. Las advertencias por sí solas mantienen un código de salida cero."
          },
          r5: {
            a1: "<code>fugo init</code> ahora genera una estructura de proyecto recomendada en vez de un <code>main.go</code> solitario: un entrypoint <code>main.go</code> delgado, un paquete <code>ui</code> para tus pantallas (<code>ui.Build</code>), una config <code>fugo.toml</code>, un <code>README.md</code>, un <code>.gitignore</code>, y <code>logs/</code> (con <code>bin/</code>/<code>dist/</code> reservados para la salida del build). También corre <code>git init</code> y hace un commit inicial (omítelo con <code>--no-git</code>).",
            a2: "<code>fugo.toml</code> — configuración declarativa de proyecto leída por el CLI <strong>y</strong> la app: <code>[window] title/width/height</code> y <code>[server] addr</code>. El <code>main.go</code> generado lo carga vía el nuevo <code>fugo.ConfigOptions(\"fugo.toml\")</code>; <code>fugo run</code> usa <code>[server] addr</code> como dirección por defecto (aún se puede sobreescribir con <code>--addr</code>), y <code>fugo build</code> incluye <code>fugo.toml</code> en <code>dist/</code>.",
            a3: "Nuevo paquete <code>config</code> libre de dependencias (<code>github.com/sazardev/fugo/config</code>) que carga <code>fugo.toml</code> (un subconjunto pequeño y fijo de TOML), compartido entre el runtime y el CLI.",
            a4: "<code>fugo run</code> ahora copia los logs de runtime de la app a <code>logs/run.log</code> (además de seguir mostrándolos en la consola).",
            c1: "Las plantillas iniciales <code>app</code> y <code>showcase</code> ahora viven en el paquete <code>ui</code> generado (<code>ui.Build</code>); el tema se configura en <code>main.go</code> antes de <code>RunStandalone</code>."
          },
          r6: {
            a1: "<code>fg.RichText(fg.Span(\"a\").Bold(), fg.Span(\"b\").Color(...).Size(...))</code> — un párrafo de fragmentos de texto con estilos mixtos.",
            a2: "<code>fg.DataTable().Columns(...).Row(...)</code> — una tabla de datos Material (con scroll horizontal).",
            a3: "<code>fg.Stepper().Step(title, content).Active(i).OnStep(fn)</code> — un asistente paso a paso."
          },
          r7: { a1: "Más overlays imperativos desde Go: <code>ctx.ShowBottomSheet(title, message)</code> (bottom sheet modal) y los selectores nativos <code>ctx.PickDate(func(date string))</code> / <code>ctx.PickTime(func(t string))</code>, que devuelven el valor elegido (ISO <code>YYYY-MM-DD</code> / 24 horas <code>HH:MM</code>) a un callback — vacío si se cancela." },
          r8: {
            a1: "Ayudantes de layout: <code>fg.AspectRatio(ratio, child)</code>, <code>fg.ClipRRect(radius, child)</code>, <code>fg.FittedBox(child)</code>, y <code>fg.Flexible(child).Flex(n)</code>.",
            a2: "<code>fg.ExpansionTile(title)</code> — un acordeón colapsable (<code>.Subtitle</code>, <code>.Leading</code>, <code>.Children</code>, <code>.InitiallyExpanded</code>).",
            a3: "<code>fg.PopupMenuButton(icon)</code> — un menú contextual/overflow (<code>.Item(value, label)</code>, <code>.OnSelected</code>)."
          },
          r9: { a1: "Más widgets Material comunes: <code>fg.Tooltip</code>, <code>fg.Badge(child).Label(...)</code>, <code>fg.CircleAvatar</code> (<code>.Text</code> / <code>.Icon</code> / <code>.BgColor</code> / <code>.Radius</code>), y <code>fg.SegmentedButton</code>." },
          r10: { a1: "Overlays imperativos manejados desde Go sobre el canal de comandos fuera de banda: <code>ctx.ShowSnackBar(text)</code> y <code>ctx.ShowDialog(title, message)</code>." },
          r11: { a1: "<code>fg.Tabs</code> — una barra de pestañas Material con una vista por pestaña (<code>.Tab(label, content)</code>, <code>.InitialIndex</code>). El cambio de pestaña se maneja en el cliente, así que no necesita ida y vuelta a Go." },
          r12: {
            a1: "<code>Scaffold.Drawer(widget)</code> — un panel lateral deslizante. Con un app bar presente y sin leading explícito, el botón de menú que lo abre aparece automáticamente.",
            a2: "<code>fg.NavigationBar</code> — una barra de navegación inferior Material 3, conectada vía <code>Scaffold.BottomBar(...)</code>."
          },
          r13: {
            a1: "<code>fg.AppBar</code> — un app bar Material completo: un título más un widget <code>.Leading</code> opcional, <code>.Actions(...)</code> al final, <code>.CenterTitle</code>, y <code>.BgColor</code>.",
            c1: "<strong>Cambio incompatible:</strong> <code>Scaffold.AppBar</code> ahora recibe un widget <code>*fg.AppBar</code> en vez de un string de título. Migra <code>fg.Scaffold(body).AppBar(\"X\")</code> → <code>fg.Scaffold(body).AppBar(fg.AppBar(\"X\"))</code>."
          },
          r14: {
            a1: "Bancos de constantes al estilo Flutter: <strong><code>fg.Icons.*</code></strong> (~2.200 iconos base), <strong><code>fg.Colors.*</code></strong> (la paleta Material), <strong><code>fg.TextSize.*</code></strong> (la escala tipográfica de Material 3).",
            a2: "<code>cmd/gen-icons</code>: una herramienta de desarrollo que regenera <code>fg/icons_gen.go</code> y <code>flutter_client/lib/icons_gen.dart</code> desde el SDK de Flutter instalado.",
            c1: "El cliente Flutter resuelve nombres de íconos a través de la tabla generada <code>materialIcons</code> en vez de un switch mantenido a mano de ~20 íconos."
          },
          r15: {
            a1: "<code>fugo upgrade</code> — auto-actualiza el CLI al último release vía <code>go install …@latest</code> (pasa una versión para fijarla, ej. <code>fugo upgrade v0.4.2</code>). En Windows el binario en ejecución se mueve a un lado (<code>&lt;exe&gt;.old</code>) para que <code>go install</code> pueda reemplazarlo.",
            c1: "Higiene del repositorio: se consolidó <code>.gitignore</code> para que los artefactos de prueba nunca lleguen al árbol. Se documentó la estructura del repositorio en <code>AGENTS.md</code>."
          },
          r16: {
            a1: "<code>FloatingActionButton</code> ahora usa un hero tag único por nodo, así una app puede mostrar múltiples FABs sin colisión de hero tag. Se añadió el ícono <code>remove</code> (menos).",
            c1: "La plantilla de <code>fugo init</code> ahora es un contador mínimo y elegante: un conteo centrado, un app bar titulado \"Fugo\", y dos FABs (decrementar / incrementar)."
          },
          r17: {
            f1: "Un <code>FloatingActionButton</code> disparaba su <code>OnClick</code> repetidamente por su cuenta — el cliente envolvía cada app en un <code>Scaffold</code> exterior, desviando los gestos del FAB anidado. La superficie exterior ahora es un <code>Material</code> plano.",
            c1: "Look más limpio y plano: el cliente aplana las superficies con tinte de seed de Material 3 a un fondo neutro; el seed sigue coloreando los elementos interactivos.",
            c2: "La plantilla de contador de <code>fugo init</code> ahora es la app Material responsiva canónica."
          },
          r18: {
            a1: "Renderizado nativo <strong>Material 3</strong>, claro por defecto, sembrado vía <code>ColorScheme.fromSeed</code> desde el <code>fg.Theme</code> activo.",
            a2: "Variantes de botón Material como constructores separados: <code>fg.FilledButton</code>, <code>fg.FilledTonalButton</code>, <code>fg.OutlinedButton</code>, <code>fg.TextButton</code>, <code>fg.ElevatedButton</code>, <code>fg.IconButton</code>.",
            a3: "Widgets Material centrales: <code>fg.Card</code>, <code>fg.Scaffold</code>, <code>fg.FloatingActionButton</code>, <code>fg.ListTile</code>, <code>fg.Chip</code>, <code>fg.ProgressCircular</code> / <code>fg.ProgressLinear</code>.",
            a4: "Controles de alineación de <code>fg.Column</code>: <code>.MainAlign</code>, <code>.CrossAlign</code>, <code>.MainAxisSize</code>, <code>.Expand</code>.",
            c1: "El renderer auto-centra una raíz de tamaño intrínseco; las raíces que llenan el viewport se dejan igual.",
            c2: "Los widgets ya no inyectan colores hex predeterminados — heredan el <code>ColorScheme</code> de Material 3 a menos que se llame un setter de color."
          },
          r19: { a1: "Archivo <code>LICENSE</code> (MIT) — restaura la documentación renderizada en pkg.go.dev." },
          r20: { f1: "<code>fugo --version</code> ahora reporta la versión correcta al instalarse con <code>go install</code>, usando como respaldo la versión del módulo y los datos de VCS de <code>runtime/debug.ReadBuildInfo()</code>." },
          r21: {
            a1: "<strong>Instalable vía <code>go install</code></strong>: los bindings de protobuf de Go generados están commiteados, así que un fetch limpio compila el CLI sin <code>protoc</code>.",
            a2: "Servicios nativos del SO: portapapeles (<code>Context.Clipboard()</code>) y diálogos de archivo nativos (<code>Context.Files().Open/Save</code>).",
            a3: "Control de ventana en tiempo de ejecución vía <code>Context.Window()</code>.",
            a4: "Nuevos widgets: <code>fg.AnimatedPositioned</code> y <code>fg.WindowDragArea</code>.",
            a5: "Ruta de prioridad inmediata del scheduler: <code>Context.UpdateNow()</code>.",
            c1: "Los bindings de protobuf de Go generados ya no están en gitignore — commiteados y mantenidos limpios con gofumpt.",
            c2: "Rendimiento: mapa de búsqueda de diff con pool de objetos, ajuste de GC, benchmarks Go + Dart detrás de un gate de regresión de rendimiento en CI."
          },
          r22: {
            a1: "Renovación de DX del CLI: logging verbose/quiet nivelado, pasos animados, ayuda enriquecida, un catálogo de widgets, <code>FUGO_LOG</code>.",
            f1: "Data race en <code>App.handlers</code> entre el scheduler y las goroutines de transporte (ahora protegido con mutex).",
            f2: "Desincronización de diff con keys — la identidad del nodo ahora es posicional por id, así que cada patch es aplicable en el cliente.",
            f3: "Propiedades de Text y Container que se podían configurar pero nunca llegaban al cable.",
            c1: "Se re-establecieron los presupuestos de rendimiento sobre protobuf estándar con un gate de diff determinista de cero-allocs."
          },
          r23: {
            a1: "Declaración del paquete raíz, git hooks de Lefthook, CI de GitHub Actions (lint/vet/build/test/format), <code>.golangci.yml</code> con 80+ linters, <code>Makefile</code>, <code>VERSION</code>, <code>CHANGELOG.md</code>.",
            i1: "Módulo Go inicializado en <code>github.com/sazardev/fugo</code> (Go 1.26.3)."
          }
        }
      }
    },

    en: {
      meta: {
        title: "Fugo — Server-Driven UI in Go",
        description: "Fugo: a local Server-Driven UI framework. Write all your logic in Go; a precompiled Flutter client renders it. Project docs, status, scope, and security."
      },
      a11y: {
        skip: "Skip to content",
        openNav: "Open navigation",
        search: "Search / jump to section",
        searchTitle: "Search (Ctrl K)",
        theme: "Toggle theme",
        sections: "Sections",
        paletteLabel: "Command palette"
      },
      nav: { search: "Search", changelog: "Changelog" },
      side: {
        group1: "Project", queEs: "What is Fugo", objetivo: "Goal", paraQuien: "Who it's for", comoFunciona: "How it works",
        group2: "Status", estado: "Status & progress", alcance: "Scope", ventajas: "Pros & cons",
        group3: "Open source", seguridad: "Security", openSource: "Open source & contributing", licencia: "License",
        group4: "Resources", contribuir: "Contributing"
      },
      hero: {
        eyebrow: "Server-Driven UI · Go + Flutter",
        tagline: "A local Server-Driven UI framework for desktop.<br>Write <strong>all</strong> your logic in Go — a precompiled Flutter client takes care of drawing it.",
        viewGithub: "View on GitHub",
        whatIsLink: "What is Fugo →"
      },
      queEs: {
        h2: "What is Fugo",
        p1: "Fugo is a local <strong>Server-Driven UI (SDUI)</strong> framework: you build native desktop applications writing <strong>exclusively in Go</strong>. Business logic, state, and routing live entirely in a Go process; a precompiled Flutter engine acts as a pure rendering terminal — no application logic, no state of its own.",
        p2: "The two processes talk over <strong>gRPC</strong> with bidirectional streaming, on <strong>Unix Domain Sockets</strong> (TCP on Windows) with <strong>Protocol Buffers</strong> as the wire format.",
        diagramLabel: "Architecture diagram",
        diagram: "┌──────────────────────┐     IPC (UDS/TCP)    ┌──────────────────────┐\n│      Go Process       │◄════════════════════►│   Flutter Process    │\n│                      │   gRPC + Protobuf     │                      │\n│  Business logic       │                       │  Widget registry     │\n│  Retained tree        │──── Tree diff ───────►│  Render pipeline     │\n│  Diffing engine       │                       │  Event debouncer     │\n│  gRPC server          │◄──── UI events ───────│  gRPC client         │\n└──────────────────────┘                       └──────────────────────┘",
        callout: "<strong>Go is the absolute source of truth.</strong> Flutter is a dumb terminal: no business logic, no state — just pixels at 60/120&nbsp;fps via Impeller."
      },
      objetivo: {
        h2: "Goal",
        p1: "Fugo aims to be the <strong>\"batteries included\"</strong> framework for building ultra-fast, efficient desktop applications using <strong>only Go</strong>: apps running in seconds, combining the iteration speed and Go knowledge you already have with Flutter's rendering power and visual ecosystem — no Dart to learn, no state split across two languages, none of the memory footprint of a Chromium-based engine.",
        p2: "It's not a generic Go↔Dart bridge or an FFI binding: it's a deliberate <em>dumb-terminal</em> architecture where 100% of the programming surface you touch is Go — <code>fg.Text</code>, <code>fg.Button</code>, <code>fg.Router</code> — and Flutter never sees a business decision, only a widget tree and its patches."
      },
      paraQuien: {
        h2: "Who it's for",
        c1h: "Go developers", c1p: "Who want a native desktop GUI without leaving Go or learning Dart/Flutter to use it.",
        c2h: "Internal tools", c2p: "Teams that already have Go backends/CLIs and need a lightweight desktop interface to operate them, without Electron's weight.",
        c3h: "Single-language apps", c3p: "Projects where keeping <em>one</em> application language (Go) is an architecture and team priority, not an aesthetic preference."
      },
      comoFunciona: {
        h2: "How it works",
        p1: "The widget tree is built <strong>once</strong> and <strong>retained</strong>. Event handlers are Go closures that mutate widget fields directly (<code>counterText.SetText(\"5\")</code>) and then call <code>ctx.Update()</code>. The scheduler (a 60&nbsp;fps ticker) re-walks that <em>same</em> tree every frame when there are pending changes, re-serializes props, diffs against the previous snapshot, and streams <strong>only the patches</strong>.",
        p2: "This is <strong>not</strong> a React-style \"rebuild from scratch\" — it's direct mutation of a retained tree, with ID/positional diffing (<code>CREATE</code>/<code>UPDATE</code>/<code>DELETE</code>/<code>REPLACE</code>/<code>REORDER</code>) and incremental streaming over gRPC."
      },
      estado: {
        h2: "Status & progress",
        intro: "<span class=\"badge badge-flame\">v3.44.8-fugo.0</span> — the engine, widget API, transport, process supervisor, CLI, and Flutter client are implemented and run end to end.",
        li1: "Installable via <code>go install github.com/sazardev/fugo/cmd/fugo@latest</code> — protobuf bindings are committed, builds on a clean fetch",
        li2: "Native Material&nbsp;3 (light by default), button variants, Card/Scaffold/FAB/ListTile/Chip/Progress",
        li3: "Diffing engine, reconciler, 60&nbsp;fps scheduler with priority (<code>Update</code> / <code>UpdateNow</code>)",
        li4: "gRPC transport (UDS / TCP on Windows), health check, keepalive, opt-in auth token",
        li5: "83 widgets in <code>fg/</code> with a fluent, prefix-free API, a <code>Theme</code> system, and an opt-in <code>fg.Store[S]</code> for shared state",
        li6: "Flutter render client (background gRPC isolate, widget registry, auto-reconnect)",
        li7: "CLI: <code>fugo init</code> / <code>run</code> (hot reload by default) / <code>build</code> / <code>doctor</code> (<code>--fix</code>) / <code>widgets</code> / <code>upgrade</code>",
        li8: "Runtime window control, clipboard, native file dialogs",
        li9: "Imperative overlays: <code>ShowSnackBar</code>, <code>ShowDialog</code>, <code>ShowBottomSheet</code>, date/time pickers",
        li10: "Performance: pooled-object diffing, GC tuning, Go + Dart benchmarks with a CI perf gate",
        outro: "See the <a href=\"https://github.com/sazardev/fugo/tree/main/ROADMAP\" target=\"_blank\" rel=\"noopener\">ROADMAP</a> and <a href=\"https://github.com/sazardev/fugo/blob/main/SPEC.md\" target=\"_blank\" rel=\"noopener\">SPEC.md</a> for the full design vision."
      },
      alcance: {
        h2: "Scope",
        inH: "In scope",
        in1: "Desktop apps: Linux, Windows, macOS",
        in2: "Strictly local IPC (UDS/TCP on <code>127.0.0.1</code>)",
        in3: "One Go binary + one precompiled Flutter client per platform",
        in4: "Material&nbsp;3 as the render client's design system",
        outH: "Out of scope (for now)",
        out1: "Mobile (iOS/Android) and web",
        out2: "Remote SDUI over a network (the design assumes low-latency local IPC, not a WAN)",
        out3: "State persistence across hot reloads (in-memory state resets)",
        out4: "A design system of its own — the client is intentionally \"vanilla\" Material&nbsp;3"
      },
      pros: {
        h2: "Pros & cons",
        prosH: "Pros",
        p1: "None of Electron/Chromium's memory footprint — native rendering via Flutter/Impeller",
        p2: "A single application language: all logic, state, and routing in Go",
        p3: "Real Material&nbsp;3 (typography, layout, animations) without reinventing a design system",
        p4: "5–10&micro;s local IPC latency, not the 50–200&nbsp;ms of a remote SDUI",
        p5: "Only diffs are sent over the wire — Protobuf framing avoids the cost of parsing JSON every frame",
        p6: "\"Batteries included\" CLI: <code>init</code> → <code>run</code> (hot reload) → <code>build</code> → <code>doctor</code>",
        consH: "Cons",
        c1: "Pre-1.0: the public API can still have breaking changes between versions",
        c2: "<code>fugo run</code>/<code>build</code> need the Flutter SDK installed (or a prebuilt binary via <code>FUGO_FLUTTER_BINARY</code>)",
        c3: "No mobile or web target yet — desktop-only by current design",
        c4: "Hot reload doesn't persist in-memory state across reloads",
        c5: "Young ecosystem and community — fewer third-party answers than plain Flutter/Dart",
        c6: "You're limited to the surface <code>fg/</code> exposes, not 100% of Flutter/Dart's widgets"
      },
      seguridad: {
        h2: "Security",
        p1: "Transport is <strong>local by default</strong> — Unix Domain Sockets (or TCP on <code>127.0.0.1</code> on Windows) — never exposed to the network unless the user explicitly configures it. <code>FUGO_AUTH=1</code> adds a per-run token that further hardens the local transport.",
        c1h: "Static analysis", c1p: "<a href=\"https://codeql.github.com/\" target=\"_blank\" rel=\"noopener\">CodeQL</a> runs on every push/PR and weekly via <code>codeql.yml</code>.",
        c2h: "Responsible disclosure", c2p: "Vulnerabilities via <a href=\"https://github.com/sazardev/fugo/security/advisories/new\" target=\"_blank\" rel=\"noopener\">private security advisories</a> — never public issues.",
        c3h: "Response commitment", c3p: "Confirmation within 48h, updates every 7 days, credit in the advisory and the CHANGELOG.",
        p2: "See <a href=\"https://github.com/sazardev/fugo/blob/main/SECURITY.md\" target=\"_blank\" rel=\"noopener\">SECURITY.md</a> for the full policy, including supported versions."
      },
      openSource: {
        h2: "Open source & contributing",
        p1: "Fugo is <strong>open to contributions</strong> under the MIT license. The workflow is documented in <a href=\"https://github.com/sazardev/fugo/blob/main/CONTRIBUTING.md\" target=\"_blank\" rel=\"noopener\">CONTRIBUTING.md</a>, and the project follows the <a href=\"https://github.com/sazardev/fugo/blob/main/CODE_OF_CONDUCT.md\" target=\"_blank\" rel=\"noopener\">Contributor Covenant</a>.",
        li1: "Structured issue templates (bug report / feature request) and a pull request template",
        li2: "<code>CODEOWNERS</code> defines required review across the whole repository",
        li3: "<code>.golangci.yml</code>: 80+ linters enabled, no tolerance threshold (<code>max-issues-per-linter: 0</code>)",
        li4: "Git hooks (Lefthook): pre-commit auto-fixes format/lint; pre-push runs the full gate",
        li5: "Weekly Dependabot for Go modules and GitHub Actions"
      },
      cicd: {
        p1: "Every push and pull request to <code>main</code> runs the same quality gate — nothing merges without passing it:",
        hLint: "Lint", pLint: "golangci-lint + staticcheck",
        hVet: "Vet", pVet: "<code>go vet ./...</code>",
        hBuild: "Build", pBuild: "<code>go build ./...</code>",
        hTest: "Test", pTest: "<code>go test -race -shuffle=on</code>",
        hFormat: "Format", format: "gofumpt, no diffs",
        hBench: "Bench", bench: "Diffing engine perf gates + benchmarks",
        hDart: "Dart", dart: "<code>flutter analyze</code> + <code>flutter test</code> in <code>flutter_client/</code>",
        hCodeQL: "CodeQL", codeql: "Security analysis, push/PR + weekly",
        p2: "Separately, <code>check-flutter-version.yml</code> checks weekly for newer stable Flutter releases and opens a PR (never auto-merged) to bump the pin, and <code>release-flutter-client.yml</code> builds the Flutter client for Linux/Windows and attaches it to the release on every <code>v*</code> tag.",
        callout: "This very page deploys itself: a GitHub Actions workflow publishes <code>site/</code> to GitHub Pages on every change to <code>main</code>, with no manual steps."
      },
      licencia: {
        h2: "License",
        p: "<strong>MIT</strong> — permissive, no restrictions on commercial use or distribution. See <a href=\"https://github.com/sazardev/fugo/blob/main/LICENSE\" target=\"_blank\" rel=\"noopener\">LICENSE</a> in the repository."
      },
      footer: {
        p1: "Fugo — built with Go + Flutter. <a href=\"https://github.com/sazardev/fugo\" target=\"_blank\" rel=\"noopener\">github.com/sazardev/fugo</a>",
        p2: "The complete framework guide (architecture, widget catalog, theming, state) lives in docs/FUGO_IDIOMATICO.md — see the \"Complete guide\" link above.",
        colProject: "Project", repo: "Repository", issues: "Issues", prs: "Pull requests", discussions: "Discussions",
        colDocs: "Docs", guide: "Complete guide",
        colSupport: "Support", star: "★ Star on GitHub", releases: "Releases"
      },
      palette: {
        placeholder: "Jump to a section… (Home, Security, CI/CD…)",
        groupSection: "Section", groupLink: "Link", groupAction: "Action", empty: "No results",
        openGithub: "Open on GitHub", viewChangelog: "View changelog", toggleTheme: "Toggle light/dark theme"
      },

      contribMeta: {
        title: "Contributing — Fugo",
        description: "How to set up the environment, code style, tests, and the pull request flow to contribute to Fugo."
      },
      contrib: {
        eyebrow: "Open source · MIT",
        group1: "On this page",
        sAntes: "Before you start", sSetup: "Development setup", sEstilo: "Code style", sTests: "Tests",
        sFlujo: "Commit & pull request", sReleases: "Releases", sReportar: "Reporting bugs / security", sConducta: "Code of conduct",
        backHome: "← Back home",
        h1: "Contributing to Fugo",
        lead: "Fugo is open under the MIT license. This is everything you need to set up the environment, follow the code style, and open a pull request that passes the CI gate on the first try.",
        fork: "Fork on GitHub", issues: "View open issues",
        antesH: "Before you start",
        antesP1: "Read <code>CLAUDE.md</code> at the repo root first — it's the canonical architecture doc (packages, data flow, wire format, conventions) and stays current with the code, unlike some of the older Spanish-language design docs under <code>ROADMAP/</code>.",
        antesP2: "For anything non-trivial (a new widget, a transport change, a CLI command), open an issue or discussion before writing code — it saves rework if the approach needs adjusting. Small fixes (typos, docs, a clear bug fix with a regression test) can go straight to a PR.",
        setupH: "Development setup",
        setupP1: "Requires <strong>Go 1.26+</strong>. The generated protobuf bindings are committed, so <code>go build ./...</code> works out of the box without <code>protoc</code>.",
        setupP2: "If you're changing <code>transport/proto/fugo/v1/fugo.proto</code>, you'll additionally need <code>protoc</code>, <code>protoc-gen-go</code>, <code>protoc-gen-go-grpc</code>, and <code>protoc-gen-dart</code> (the Dart plugin must be <code>protoc_plugin</code> 21.1.x — see <code>CLAUDE.md</code>).",
        setupP3: "If you're touching <code>flutter_client/</code> (the Dart render client), you'll need the Flutter SDK pinned in <code>FLUTTER_VERSION</code> — see <code>flutter_client/README.md</code>.",
        estiloH: "Code style",
        estilo1: "The formatter is <code>gofumpt</code>, not <code>gofmt</code> — run <code>gofumpt -w .</code> before committing (Lefthook does this automatically for staged files)",
        estilo2: "<code>make lint</code> runs <code>golangci-lint</code> (80+ linters in <code>.golangci.yml</code>) and <code>staticcheck</code> — both must pass clean, no tolerance threshold",
        estilo3: "<code>make vet</code> (<code>go vet ./...</code>) must pass",
        estilo4: "Keep changes scoped: a bug fix shouldn't carry an unrelated refactor. No speculative abstractions for hypothetical future needs",
        estilo5: "Widgets follow <code>fg/</code> conventions: prefix-free constructors (<code>fg.Text</code>, not <code>NewText</code>), chainable setters, a <code>*Set bool</code> flag for optional props",
        testsH: "Tests",
        testsP1: "<code>make test</code> runs the full suite with <code>-race -shuffle=on</code>. New code needs test coverage; bug fixes need a regression test that fails before the fix and passes after.",
        testsP2: "Table-driven tests are the house style on the Go side. Benchmarks live in <code>*_bench_test.go</code> (<code>engine/</code> has the perf gates CI checks on every push — don't regress them without a good reason documented in the PR).",
        flujoH: "Commit & pull request",
        paso1: "Branch off <code>main</code>.",
        paso2: "Make your change, keep commits focused.",
        paso3: "<code>gofumpt -w .</code>, <code>make lint</code>, <code>make vet</code>, <code>make test</code>, <code>go mod tidy</code> — all clean.",
        paso4: "Update <code>CHANGELOG.md</code> under <code>[Unreleased]</code> if the change is user-facing.",
        paso5: "Open the PR — <code>.github/pull_request_template.md</code> will prompt you for the checklist above. <code>make pr MSG=\"type: description\"</code> automates the branch/push/PR-open steps if you're working from <code>main</code> directly.",
        flujoP: "CI runs lint, vet, build, test (<code>-race -shuffle=on</code>), a gofumpt format check, the perf gate/benchmarks, and the Dart job (<code>flutter analyze</code> + <code>flutter test</code>) on every push and PR — all must pass before merge. CodeQL runs weekly and on every push for security analysis.",
        releasesH: "Releases",
        releasesP: "Don't hand-edit <code>VERSION</code>, <code>FLUTTER_VERSION</code>, or <code>CHANGELOG.md</code> — those are maintainer-run via <code>make release</code> / <code>make release-flutter</code>. See the <a href=\"changelog.html#versionamiento\">versioning scheme</a> for details.",
        releasesLink: "versioning scheme",
        reportarH: "Reporting bugs / security",
        reportarP1: "Use the <a href=\"https://github.com/sazardev/fugo/issues/new/choose\" target=\"_blank\" rel=\"noopener\">issue templates</a> — they ask for the right context up front (repro steps, Go/Flutter versions, expected vs actual behavior).",
        reportarP2: "For security vulnerabilities, use a <a href=\"https://github.com/sazardev/fugo/security/advisories/new\" target=\"_blank\" rel=\"noopener\">private security advisory</a> instead of a public issue — see <a href=\"https://github.com/sazardev/fugo/blob/main/SECURITY.md\" target=\"_blank\" rel=\"noopener\">SECURITY.md</a>.",
        conductaH: "Code of conduct",
        conductaP: "This project follows the <a href=\"https://github.com/sazardev/fugo/blob/main/CODE_OF_CONDUCT.md\" target=\"_blank\" rel=\"noopener\">Contributor Covenant</a>. By participating, you're expected to uphold it."
      },

      chMeta: {
        title: "Changelog — Fugo",
        description: "Fugo's full version history and the versioning scheme tied to Flutter."
      },
      ch: {
        eyebrow: "Version history",
        sVersion: "Versioning", sHistorial: "Version history",
        h1: "Changelog",
        lead: "Fugo's complete history, version by version — translated from the repository's <code>CHANGELOG.md</code>.",
        viewSource: "View CHANGELOG.md on GitHub", viewReleases: "View releases",
        unreleased: "Unreleased",
        grpAdded: "Added", grpChanged: "Changed", grpFixed: "Fixed", grpInfra: "Infrastructure",
        versionH: "Versioning",
        versionP1: "Fugo's version deliberately tracks Flutter: <code>VERSION</code> is <code>&lt;flutter-version&gt;-fugo.&lt;build&gt;</code> (e.g. <code>3.44.8-fugo.0</code>) — the first three numbers are always the exact Flutter SDK version the precompiled client and <code>flutter_client/</code> are built against, so there's never ambiguity between a Fugo release and the Flutter it renders with.",
        versionP2: "<code>FLUTTER_VERSION</code> is the single source of truth for that pin: <code>release-flutter-client.yml</code> builds against it exactly (not a floating channel like <code>stable</code>), and <code>fugo doctor</code> warns (doesn't fail — the Flutter SDK is optional) if your installed Flutter doesn't match.",
        versionP3: "A weekly workflow, <code>check-flutter-version.yml</code>, checks for a newer stable Flutter release and opens a PR bumping the pin — Fugo is meant to track current Flutter continuously, not drift behind it. That PR is never auto-merged, since a Flutter upgrade can break <code>flutter_client/</code>'s Dart code.",
        versionP4: "Two commands cover a release, neither hand-edits <code>VERSION</code>/<code>CHANGELOG.md</code>: <code>make release MSG=\"...\"</code> bumps only the <code>-fugo.N</code> suffix (a Fugo-only fix against the same Flutter version); <code>make release-flutter FLUTTER=&lt;version&gt; MSG=\"...\"</code> bumps the targeted Flutter version (resets to <code>-fugo.0</code>). There's no <code>TYPE=patch|minor|major</code> anymore — major/minor don't mean anything under this scheme.",
        historialH: "Version history",
        historialLead: "Format based on <a href=\"https://keepachangelog.com/en/1.1.0/\" target=\"_blank\" rel=\"noopener\">Keep a Changelog</a>.",
        hist: {
          r2: {
            c1: "<strong>Fugo's version now tracks the exact Flutter SDK version it targets.</strong> <code>VERSION</code> is <code>&lt;flutter-version&gt;-fugo.&lt;build&gt;</code> (e.g. <code>3.44.8-fugo.0</code>) — the first three numbers always tell you exactly which Flutter release the precompiled client and <code>flutter_client/</code> were built against, so there's never ambiguity between a Fugo release and the Flutter it renders with. <code>FLUTTER_VERSION</code> is the new single source of truth for that pin: CI (<code>release-flutter-client.yml</code>) builds against it exactly instead of a floating <code>channel: stable</code>, and <code>fugo doctor</code> warns if your installed Flutter SDK doesn't match it.",
            c2: "<code>make release MSG=\"...\"</code> now only bumps the <code>-fugo.N</code> build suffix (a Fugo-only fix against the same Flutter version) — <code>TYPE=patch|minor|major</code> is gone, since major/minor no longer mean anything under this scheme. Upgrading the targeted Flutter version is a separate, explicit step: <code>make release-flutter FLUTTER=&lt;version&gt; MSG=\"...\"</code>.",
            c3: "A new scheduled workflow (<code>check-flutter-version.yml</code>) checks weekly for a newer stable Flutter release and opens a PR bumping <code>FLUTTER_VERSION</code> — Fugo is meant to track current Flutter, not drift behind it. The PR always needs a human to review and merge (a Flutter upgrade can affect <code>flutter_client/</code>'s Dart code), it's never auto-merged."
          },
          r3: {
            a1: "<code>fugo doctor</code> now checks <strong>project coherence</strong>: it reads <code>go.mod</code>'s module path and <code>main.go</code>'s <code>&lt;module&gt;/ui</code> import and reports a precise ✗ when they don't match, when the imported <code>ui</code> package is missing, or when it has no exported <code>Build</code> — caught before the generic compile error.",
            a2: "<code>fugo doctor --fix</code> repairs the auto-fixable bits inside a project before diagnosing: writes a default <code>fugo.toml</code> if missing, runs <code>git init</code> if there's no repo, and <code>go mod tidy</code> to resolve dependencies.",
            c1: "<strong><code>fugo run</code> hot-reloads by default.</strong> It now watches <code>.go</code> files and rebuilds/reconnects the Go server on every change while the Flutter window stays open — so edits to text, handlers, layout, etc. show up live (in-memory state resets across reloads). Pass <code>--no-watch</code> for a single build-and-run; the old <code>--watch</code> flag is now implied (hidden, no-op).",
            f1: "Flutter render client: reconnection after a server restart (hot reload) no longer fails with \"Bad state: Stream has already been listened to\". The isolate now subscribes to its <code>ReceivePort</code> once and routes events to the active stream, and resets its backoff after a successful session so it reconnects within ~500ms."
          },
          r4: {
            a1: "<code>fugo doctor</code> is now a two-part health check. <strong>Toolchain</strong>: Go, Flutter, git, protoc, gofumpt (Go/Flutter are required → ✗; the rest are warnings) plus the platform. <strong>Project</strong> (when run inside a Fugo project): it imports <code>fugo.toml</code> and reports the resolved window/address, checks the configured gRPC address is free (or a unix socket), verifies the structure (<code>main.go</code>, <code>ui/</code>, <code>go.mod</code>), that the <code>github.com/sazardev/fugo</code> module resolves, whether the Flutter client is built (or buildable), and finally compiles the project (<code>go build ./...</code>).",
            c1: "<code>fugo doctor</code> now exits non-zero when there's a blocking issue (a required tool missing, an unresolved module, or a project that doesn't compile), so it can gate scripts/CI. Warnings alone keep a zero exit."
          },
          r5: {
            a1: "<code>fugo init</code> now scaffolds a recommended project structure instead of a lone <code>main.go</code>: a thin <code>main.go</code> entrypoint, a <code>ui</code> package for your screens (<code>ui.Build</code>), a <code>fugo.toml</code> config, a <code>README.md</code>, a <code>.gitignore</code>, and <code>logs/</code> (with <code>bin/</code>/<code>dist/</code> reserved for build output). It also runs <code>git init</code> and makes an initial commit (skip with <code>--no-git</code>).",
            a2: "<code>fugo.toml</code> — declarative project config read by the CLI <strong>and</strong> the app: <code>[window] title/width/height</code> and <code>[server] addr</code>. Generated <code>main.go</code> loads it via the new <code>fugo.ConfigOptions(\"fugo.toml\")</code>; <code>fugo run</code> uses <code>[server] addr</code> as the default address (still overridable with <code>--addr</code>), and <code>fugo build</code> ships <code>fugo.toml</code> into <code>dist/</code>.",
            a3: "New dependency-free <code>config</code> package (<code>github.com/sazardev/fugo/config</code>) that loads <code>fugo.toml</code> (a small, fixed TOML subset), shared by the runtime and the CLI.",
            a4: "<code>fugo run</code> now tees the app's runtime logs to <code>logs/run.log</code> (still streamed to the console).",
            c1: "The <code>app</code> and <code>showcase</code> starter templates now live in the generated <code>ui</code> package (<code>ui.Build</code>); the theme is set in <code>main.go</code> before <code>RunStandalone</code>."
          },
          r6: {
            a1: "<code>fg.RichText(fg.Span(\"a\").Bold(), fg.Span(\"b\").Color(...).Size(...))</code> — a paragraph of mixed-style text runs.",
            a2: "<code>fg.DataTable().Columns(...).Row(...)</code> — a Material data table (horizontally scrollable).",
            a3: "<code>fg.Stepper().Step(title, content).Active(i).OnStep(fn)</code> — a step-by-step wizard."
          },
          r7: { a1: "More imperative overlays from Go: <code>ctx.ShowBottomSheet(title, message)</code> (modal bottom sheet) and the native pickers <code>ctx.PickDate(func(date string))</code> / <code>ctx.PickTime(func(t string))</code>, which return the chosen value (ISO <code>YYYY-MM-DD</code> / 24-hour <code>HH:MM</code>) to a callback — empty if cancelled." },
          r8: {
            a1: "Layout helpers: <code>fg.AspectRatio(ratio, child)</code>, <code>fg.ClipRRect(radius, child)</code>, <code>fg.FittedBox(child)</code>, and <code>fg.Flexible(child).Flex(n)</code>.",
            a2: "<code>fg.ExpansionTile(title)</code> — a collapsible accordion (<code>.Subtitle</code>, <code>.Leading</code>, <code>.Children</code>, <code>.InitiallyExpanded</code>).",
            a3: "<code>fg.PopupMenuButton(icon)</code> — an overflow/context menu (<code>.Item(value, label)</code>, <code>.OnSelected</code>)."
          },
          r9: { a1: "More common Material widgets: <code>fg.Tooltip</code>, <code>fg.Badge(child).Label(...)</code>, <code>fg.CircleAvatar</code> (<code>.Text</code> / <code>.Icon</code> / <code>.BgColor</code> / <code>.Radius</code>), and <code>fg.SegmentedButton</code>." },
          r10: { a1: "Imperative overlays driven from Go over the out-of-band command channel: <code>ctx.ShowSnackBar(text)</code> and <code>ctx.ShowDialog(title, message)</code>." },
          r11: { a1: "<code>fg.Tabs</code> — a Material tab strip with one view per tab (<code>.Tab(label, content)</code>, <code>.InitialIndex</code>). Tab switching is handled on the client, so it needs no round-trip to Go." },
          r12: {
            a1: "<code>Scaffold.Drawer(widget)</code> — a slide-in side panel. With an app bar present and no explicit leading, the menu button that opens it appears automatically.",
            a2: "<code>fg.NavigationBar</code> — a Material 3 bottom navigation bar, attached via <code>Scaffold.BottomBar(...)</code>."
          },
          r13: {
            a1: "<code>fg.AppBar</code> — a full Material app bar: a title plus an optional <code>.Leading</code> widget, trailing <code>.Actions(...)</code>, <code>.CenterTitle</code>, and <code>.BgColor</code>.",
            c1: "<strong>Breaking:</strong> <code>Scaffold.AppBar</code> now takes an <code>*fg.AppBar</code> widget instead of a title string. Migrate <code>fg.Scaffold(body).AppBar(\"X\")</code> → <code>fg.Scaffold(body).AppBar(fg.AppBar(\"X\"))</code>."
          },
          r14: {
            a1: "Flutter-style constant banks: <strong><code>fg.Icons.*</code></strong> (~2,200 base icons), <strong><code>fg.Colors.*</code></strong> (the Material palette), <strong><code>fg.TextSize.*</code></strong> (the Material 3 type scale).",
            a2: "<code>cmd/gen-icons</code>: a dev tool that regenerates <code>fg/icons_gen.go</code> and <code>flutter_client/lib/icons_gen.dart</code> from the installed Flutter SDK.",
            c1: "The Flutter client resolves icon names through the generated <code>materialIcons</code> table instead of a hand-maintained ~20-icon switch."
          },
          r15: {
            a1: "<code>fugo upgrade</code> — self-update the CLI to the latest release via <code>go install …@latest</code> (pass a version to pin, e.g. <code>fugo upgrade v0.4.2</code>). On Windows the running binary is moved aside (<code>&lt;exe&gt;.old</code>) so <code>go install</code> can replace it.",
            c1: "Repo hygiene: consolidated <code>.gitignore</code> so scratch artifacts never land in the tree. Documented the repository layout in <code>AGENTS.md</code>."
          },
          r16: {
            a1: "<code>FloatingActionButton</code> now uses a unique hero tag per node, so an app can show multiple FABs without a Hero tag collision. Added the <code>remove</code> (minus) icon.",
            c1: "<code>fugo init</code>'s template is now a minimal, elegant counter: a centered count, an app bar titled \"Fugo\", and two FABs (decrement / increment)."
          },
          r17: {
            f1: "A <code>FloatingActionButton</code> fired its <code>OnClick</code> repeatedly on its own — the client wrapped every app in an outer <code>Scaffold</code>, misrouting nested FAB gestures. The outer surface is now a plain <code>Material</code>.",
            c1: "Cleaner, flatter look: the client flattens Material 3 seed-tinted surfaces to a neutral background; the seed still colors interactive elements.",
            c2: "<code>fugo init</code>'s counter template is now the canonical responsive Material app."
          },
          r18: {
            a1: "Native <strong>Material 3</strong> rendering, light by default, seeded via <code>ColorScheme.fromSeed</code> from the active <code>fg.Theme</code>.",
            a2: "Material button variants as separate constructors: <code>fg.FilledButton</code>, <code>fg.FilledTonalButton</code>, <code>fg.OutlinedButton</code>, <code>fg.TextButton</code>, <code>fg.ElevatedButton</code>, <code>fg.IconButton</code>.",
            a3: "Core Material widgets: <code>fg.Card</code>, <code>fg.Scaffold</code>, <code>fg.FloatingActionButton</code>, <code>fg.ListTile</code>, <code>fg.Chip</code>, <code>fg.ProgressCircular</code> / <code>fg.ProgressLinear</code>.",
            a4: "<code>fg.Column</code> alignment controls: <code>.MainAlign</code>, <code>.CrossAlign</code>, <code>.MainAxisSize</code>, <code>.Expand</code>.",
            c1: "The renderer auto-centers an intrinsically-sized root; roots that fill the viewport are left as-is.",
            c2: "Widgets no longer inject opinionated hex colors — they inherit the Material 3 <code>ColorScheme</code> unless a color setter is called."
          },
          r19: { a1: "<code>LICENSE</code> file (MIT) — restores rendered documentation on pkg.go.dev." },
          r20: { f1: "<code>fugo --version</code> now reports the correct version when installed with <code>go install</code>, falling back to module version and VCS stamps from <code>runtime/debug.ReadBuildInfo()</code>." },
          r21: {
            a1: "<strong>Installable via <code>go install</code></strong>: generated protobuf Go bindings are committed, so a clean fetch builds the CLI without <code>protoc</code>.",
            a2: "OS host services: clipboard (<code>Context.Clipboard()</code>) and native file dialogs (<code>Context.Files().Open/Save</code>).",
            a3: "Runtime window control via <code>Context.Window()</code>.",
            a4: "New widgets: <code>fg.AnimatedPositioned</code> and <code>fg.WindowDragArea</code>.",
            a5: "Scheduler immediate-priority path: <code>Context.UpdateNow()</code>.",
            c1: "Generated Go protobuf bindings are no longer gitignored — committed and kept gofumpt-clean.",
            c2: "Performance: object-pooled diff lookup map, GC tuning, Go + Dart benchmarks behind a CI perf-regression gate."
          },
          r22: {
            a1: "CLI DX overhaul: leveled verbose/quiet logging, animated steps, rich help, a widgets catalog, <code>FUGO_LOG</code>.",
            f1: "Data race on <code>App.handlers</code> between the scheduler and transport goroutines (now mutex-guarded).",
            f2: "Keyed diff desync — node identity is now positional by id, so every patch is client-applicable.",
            f3: "Text and Container properties that were settable but never reached the wire.",
            c1: "Re-baselined performance budgets on standard protobuf with a deterministic zero-alloc diff gate."
          },
          r23: {
            a1: "Root package declaration, Lefthook git hooks, GitHub Actions CI (lint/vet/build/test/format), <code>.golangci.yml</code> with 80+ linters, <code>Makefile</code>, <code>VERSION</code>, <code>CHANGELOG.md</code>.",
            i1: "Go module initialized at <code>github.com/sazardev/fugo</code> (Go 1.26.3)."
          }
        }
      }
    }
  };

  function get(lang, key) {
    var parts = key.split(".");
    var obj = translations[lang];
    for (var i = 0; i < parts.length && obj != null; i++) obj = obj[parts[i]];
    return typeof obj === "string" ? obj : null;
  }

  function detectLang() {
    try {
      var stored = localStorage.getItem("fugo-lang");
      if (stored === "es" || stored === "en") return stored;
    } catch (e) { /* storage unavailable */ }
    return navigator.language && navigator.language.toLowerCase().indexOf("en") === 0 ? "en" : "es";
  }

  var currentLang = detectLang();

  function applyLang(lang) {
    currentLang = lang;
    document.documentElement.lang = lang;

    document.querySelectorAll("[data-i18n]").forEach(function (el) {
      var keys = el.getAttribute("data-i18n").split("|");
      var attrsAttr = el.getAttribute("data-i18n-attr");
      if (attrsAttr) {
        attrsAttr.split("|").forEach(function (attr, i) {
          var val = get(lang, keys[i] || keys[0]);
          if (val == null) return;
          if (attr === "text") el.textContent = val;
          else el.setAttribute(attr, val);
        });
      } else {
        var val = get(lang, keys[0]);
        if (val != null) el.innerHTML = val;
      }
    });

    var langLabel = document.getElementById("langLabel");
    if (langLabel) langLabel.textContent = lang === "es" ? "EN" : "ES";

    try { localStorage.setItem("fugo-lang", lang); } catch (e) { /* storage unavailable */ }
    document.dispatchEvent(new CustomEvent("fugo:langchange", { detail: { lang: lang } }));
  }

  window.fugoI18n = {
    t: function (key) { return get(currentLang, key); },
    lang: function () { return currentLang; },
    set: applyLang
  };

  document.addEventListener("DOMContentLoaded", function () {
    applyLang(currentLang);
    var toggle = document.getElementById("langToggle");
    if (toggle) {
      toggle.addEventListener("click", function () {
        applyLang(currentLang === "es" ? "en" : "es");
      });
    }
  });
})();
