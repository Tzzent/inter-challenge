# 02 — Decisiones para defender en la entrevista

Cada decisión con su "por qué".

## 1. QR y no "rotación"

El PDF dice "rotación" en la arquitectura y "factorización QR" en la funcionalidad. Implementamos QR porque:

- El PDF habla de "matrices devueltas" (plural) y pregunta si "alguna matriz es diagonal": eso exige más de una matriz, y QR devuelve dos.
- **Q es una rotación o reflexión**: no contradice la palabra "rotación".
- Las **rotaciones de Givens** son justamente una forma de calcular QR. "Rotar" es el medio; QR es el resultado.

## 2. Householder y no Gram-Schmidt

- Gram-Schmidt clásico (lo que se hace a mano) pierde ortogonalidad con números de punto flotante: al restar vectores casi paralelos se cancelan dígitos.
- Householder aplica reflexiones (espejos), que son ortogonales por construcción, y elige el signo que evita cancelaciones. Es el estándar en librerías numéricas (LAPACK).
- Está escrito a mano en `go-api/internal/qr` porque esa es la lógica que el reto quiere ver implementada.

## 3. QR completo y no reducido

- Completo: Q es m×m y R es m×n. Reducido: Q es m×k y R es k×n con k = min(m,n).
- El completo funciona igual para cualquier forma (m>n, m<n, m=n) y Q siempre es cuadrada y ortogonal, lo cual es fácil de verificar (QᵀQ = I).
- Costo: con m ≫ n, Q es más grande. Con el límite de 100×100 no importa.

## 4. Epsilon = 1e-9

- El punto flotante deja residuos como `1e-16` donde matemáticamente hay 0.
- Tras el QR forzamos a 0 exacto lo que está debajo de la diagonal de R (es 0 por definición).
- Para "¿es diagonal?" usamos `|x| ≤ 1e-9`: muy por encima del ruido de redondeo (~1e-15 con valores ≤ 1e6) y muy por debajo de cualquier valor "real". Es una constante documentada en `node-api/src/stats.ts`.

## 5. Límites de entrada

- Matriz no vacía, rectangular, solo números finitos, máximo 100×100, y límite de tamaño del body.
- Por qué: QR es O(m²n); sin límites, una petición enorme consume CPU y memoria (DoS). Se valida en el borde (la API) y se responde 400 con `{"error": "..."}`.

## 6. gonum solo como oráculo de pruebas

- Usar gonum en producción reduciría el reto a llamar una librería.
- En los tests compara nuestro resultado con uno independiente y confiable. Como Householder puede elegir signos distintos, comparamos tras normalizar signos (que la diagonal de R sea ≥ 0).
- Además, los tests verifican propiedades sin oráculo: Q·R ≈ A, QᵀQ ≈ I, R triangular superior.

## 7. Diseño de JWT

- HS256 con un secreto compartido (`JWT_SECRET`) entre ambas APIs: simple y suficiente para dos servicios del mismo dueño.
- Go emite el token (`POST /auth/token`, credenciales demo por variables de entorno, comparación en tiempo constante, expira en 1 h) y reenvía el header `Authorization` a Node: Node también protege su API (defensa en profundidad).
- El algoritmo se fija en HS256: se rechaza `alg=none` y cualquier otro (ataque clásico de confusión de algoritmo).
- `/health` y `/auth/token` son públicos.
- Mejora futura: RS256 (solo Go tiene la clave privada) y un proveedor de identidad real.

## 8. Todo en Docker

- Go no está instalado en la máquina; Docker garantiza que lo que corre en local es lo mismo que corre en la nube (Cloud Run ejecuta contenedores).
- Dockerfiles multi-etapa: la etapa `build` tiene el toolchain y corre los tests; la etapa `runtime` es mínima (distroless / alpine) y corre como usuario no root.
- Los tests también corren en contenedores (`docker compose --profile test ...`), así cualquiera los reproduce con un solo comando.

## 9. TypeScript en Node

- Tipos para el contrato (`number[][][]`, la forma de las estadísticas): los errores aparecen al compilar, no en producción.
- Costo bajo: `tsx` ejecuta los tests sin compilar; `tsc` compila para la imagen final.

## 10. Respuesta 502 cuando Node falla

- Si Node no responde, responde con error o rechaza el token, Go devuelve 502 (Bad Gateway): el problema es del servicio de atrás, no del cliente. Timeout de 5 s para no colgar peticiones.
- `GET /health` de Go responde 200 aunque Node esté caído, e informa su estado en un campo: Go sigue vivo y lo reporta.
