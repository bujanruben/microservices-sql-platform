# Sistema Meteorológico Distribuido con Go, Docker y PostgreSQL

## Requisitos previos:

* Contar con docker y docker compose instalado en el sistema.

Sistema compuesto por dos microservicios:

* Servicio A: Ingesta y almacenamiento de datos meteorológicos
* Servicio B: API de consulta con caché y agregaciones

## Estructura del proyecto:
* docker-compose.yml
* script_prueba.ps1
* En Carpetas A y B
    * Dockerfile en cada carpeta: define la imagen en docker del proyecto en go.
    * cmd/api/: contiene los archivos de las funciones en go.
    * data/: almacena el csv con los datos meteorológicos.
    * docs/: contiene los archivos openAPI con documentación de la API.
    * go.mod: define los proyectos de go.
    * go.sum: define las dependencias externas a go.

### Archivos Go del Servicio A (A/cmd/api/)

* main.go: Servidor HTTP principal del Servicio A
* handlers.go: Endpoints /ingest/csv, /cities, /raw
* handlers_test.go: Tests de los endpoints del Servicio A
* database.go: Operaciones con PostgreSQL (conexión, inserción, consultas)
* database_test.go: Tests de operaciones de base de datos
* input.go: Procesamiento de archivos CSV
* input_test.go: Tests de procesamiento CSV

### Archivos Go del Servicio B (B/cmd/api/)

* main.go: Servidor HTTP principal del Servicio B
* handlers.go: Endpoint /weather/{city} con parámetros de consulta
* handlers_test.go: Tests del endpoint meteorológicos
* database.go: Consultas complejas y agregaciones (daily, rolling7)
* database_test.go: Tests de consultas meteorológicos
* cache.go: Sistema de caché con TTL 10 minutos
* cache_test.go: Tests del sistema de caché

## Como ejecutar y usar:

1. Construir y ejecutar todos los servicios ( mediantela consola)

        docker-compose up -d --build

Con esto se inician los contenedores de los servicios A y B

2. Verificar que los servicios estén ejecutándose

        docker-compose ps

# Servicio A - Ingesta de Datos

0. Consulta de los contenedores del servicio A

    En docker-compose ps deben aparecer, entre otros,dos contenedores:
    
            meteo-db     postgres:18         "docker-entrypoint.s…"   meteo-db    13 hours ago   Up 7 minutes   0.0.0.0:5432->5432/tcp, [::]:5432->5432/tcp
            servicio-a   servicio-a:latest   "./main"                 servicioA   13 hours ago   Up 7 minutes   0.0.0.0:8080->8080/tcp, [::]:8080->8080/tcp


1. Ingestar archivo CSV

        curl -X POST -F "file=@A/data/datos.csv" http://localhost:8080/ingest/csv

(El archivo puede ser un csv con una estructura equivalente)

Respuesta esperada:

    {"rows_inserted": 118, "rows_rejected": 6, "elapsed_ms": 1760, "file_checksum": "sha256:6f71d53ff0f64e9567b372ad603c855c4be31b099b505b4d57b2c8f6c2a84ac5"}

2. Obtener ciudades disponibles

        curl http://localhost:8080/cities

Respuesta esperada:

        ["Gijón","Madrid","Sevilla","Valencia"]

3. Consultar datos crudos 

        curl "http://localhost:8080/raw?city=Madrid&from=2025-01-01&to=2025-12-15&page=1&limit=5"

Los parámetros introducidos son:
*   city (obligatorio): La ciudad a consultar
*   from (obligatorio): La fecha de inicio de los datos
*   to (obligatorio): La fecha final de los datos
    (Ambos con formato YYYY-MM-DD)
*   page (opcional, por defecto: 1): La página
*   limit (opcional, por defecto: 5): El número de elementos en la página

Respuesta esperada:

    [{"Fecha":"2025-10-12T00:00:00Z","Ciudad":"Madrid","TempMax":11.55,"TempMin":6.25,"Precipitacion":0,"Nubosidad":10},{"Fecha":"2025-10-12T00:00:00Z","Ciudad":"Madrid","TempMax":11.55,"TempMin":6.25,"Precipitacion":0,"Nubosidad":10},{"Fecha":"2025-10-13T00:00:00Z","Ciudad":"Madrid","TempMax":12.35,"TempMin":5.25,"Precipitacion":0.2,"Nubosidad":60},{"Fecha":"2025-10-13T00:00:00Z","Ciudad":"Madrid","TempMax":12.35,"TempMin":5.25,"Precipitacion":0.2,"Nubosidad":60},{"Fecha":"2025-10-14T00:00:00Z","Ciudad":"Madrid","TempMax":12.65,"TempMin":7.15,"Precipitacion":0,"Nubosidad":30}]

# Servicio B - Consultas meteorológicas

0. Consulta del contenedor del servicio B

        En docker-compose ps deben aparecer:
            servicio-b   servicio-b:latest   "./main"                 servicioB   13 hours ago   Up 7 minutes   0.0.0.0:8081->8081/tcp, [::]:8081->8081/tcp

1. Consulta meteorológica básica

        curl "http://localhost:8081/weather/Madrid?date=2025-10-15&days=5&unit=C&agg=daily"

Los parámetros introducidos son:
*   city (obligatorio): Como parte del path en http://localhost:8081/weather/{city}
*   date (obligatorio): La fecha de inicio (Formato YYYY-MM-DD)
*   days (opcional, por defecto: 5): Días a consultar, 1-10
*   unit (opcional, por defecto: C): Unidad de temperatura C/F (celsius/fahrenheit)
*   agg (opcional): Agregación (daily|rolling7), su ausencia será una consulta cruda


Ejemplos:

1.  Consulta con conversión a fahrenheit

        curl "http://localhost:8081/weather/Madrid?date=2025-10-15&days=1&unit=F"

Respuesta esperada:

    {"agg":"","city":"Madrid","count":2,"data":[{"Fecha":"2025-10-15T00:00:00Z","Ciudad":"Madrid","TempMax":60.35,"TempMin":42.53,"Precipitacion":1.4,"Nubosidad":80},{"Fecha":"2025-10-15T00:00:00Z","Ciudad":"Madrid","TempMax":60.35,"TempMin":42.53,"Precipitacion":1.4,"Nubosidad":80}],"days":1,"from":"2025-10-15","to":"2025-10-15","unit":"F"}

2.  Consulta con agregación rolling7

        curl "http://localhost:8081/weather/Madrid?date=2025-10-15&days=7&unit=C&agg=rolling7"

Respuesta esperada:

    {"agg":"rolling7","city":"Madrid","count":14,"data":[{"Fecha":"2025-10-15T00:00:00Z","Ciudad":"Madrid","TempMax":10.8,"TempMin":10.8,"Precipitacion":1.4,"Nubosidad":80},{"Fecha":"2025-10-15T00:00:00Z","Ciudad":"Madrid","TempMax":10.8,"TempMin":10.8,"Precipitacion":2.8,"Nubosidad":80},{"Fecha":"2025-10-16T00:00:00Z","Ciudad":"Madrid","TempMax":11.333333333333334,"TempMin":11.333333333333334,"Precipitacion":5.4,"Nubosidad":86.66666666666667},{"Fecha":"2025-10-16T00:00:00Z","Ciudad":"Madrid","TempMax":11.6,"TempMin":11.6,"Precipitacion":8,"Nubosidad":90},{"Fecha":"2025-10-17T00:00:00Z","Ciudad":"Madrid","TempMax":11.72,"TempMin":11.72,"Precipitacion":8.3,"Nubosidad":82},{"Fecha":"2025-10-17T00:00:00Z","Ciudad":"Madrid","TempMax":11.799999999999999,"TempMin":11.799999999999999,"Precipitacion":8.600000000000001,"Nubosidad":76.66666666666667},{"Fecha":"2025-10-18T00:00:00Z","Ciudad":"Madrid","TempMax":12.071428571428571,"TempMin":12.071428571428571,"Precipitacion":8.700000000000001,"Nubosidad":68.57142857142857},{"Fecha":"2025-10-18T00:00:00Z","Ciudad":"Madrid","TempMax":12.485714285714286,"TempMin":12.485714285714286,"Precipitacion":7.399999999999999,"Nubosidad":60},{"Fecha":"2025-10-19T00:00:00Z","Ciudad":"Madrid","TempMax":13.157142857142858,"TempMin":13.157142857142858,"Precipitacion":6.099999999999999,"Nubosidad":50},{"Fecha":"2025-10-19T00:00:00Z","Ciudad":"Madrid","TempMax":13.6,"TempMin":13.6,"Precipitacion":3.6,"Nubosidad":37.142857142857146},{"Fecha":"2025-10-20T00:00:00Z","Ciudad":"Madrid","TempMax":14.12857142857143,"TempMin":14.12857142857143,"Precipitacion":0.9999999999999999,"Nubosidad":22.857142857142858},{"Fecha":"2025-10-20T00:00:00Z","Ciudad":"Madrid","TempMax":14.685714285714283,"TempMin":14.685714285714283,"Precipitacion":0.7,"Nubosidad":15.714285714285714},{"Fecha":"2025-10-21T00:00:00Z","Ciudad":"Madrid","TempMax":15.135714285714284,"TempMin":15.135714285714284,"Precipitacion":0.4,"Nubosidad":10},{"Fecha":"2025-10-21T00:00:00Z","Ciudad":"Madrid","TempMax":15.37142857142857,"TempMin":15.37142857142857,"Precipitacion":0.30000000000000004,"Nubosidad":8.571428571428571}],"days":7,"from":"2025-10-15","to":"2025-10-21","unit":"C"}



### Pruebas:

    cd .\A\cmd\api\
    go test -v . 

    cd .\B\cmd\api\
    go test -v . 

Las pruebas cubren los puntos más críticos del proyecto y los pasan satisfactoriamente.
Se comprueban transformaciones de datos críticas, solicitudes incorrectas y su correcta respuesta inválida o el uso de parámetros incorrectos


#### Arquitectura Actual

* 2 servicios + BD en contenedores separados

* Servicio B en puerto 8081 hace consultas a BD con caché

* Bridge network para que se comuniquen entre sí
