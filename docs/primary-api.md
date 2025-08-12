# A3 Mercados

**Abril 2025**
## Primary API
**v1.21**

---

## Índice

- **Historia del documento** `4`
- **Introducción** `7`
- **Conectándose a la API por token de autenticación** `9`
- **Consultas a la API REST de Trading** `10`
- **Funcionalidades de la API REST** `10`
- **Funcionalidades de la API WebSocket** `11`
- **Segmentos** `12`
  - Lista de Segmentos disponibles `12`
- **Instrumentos (Securities)** `13`
  - Lista de Segmentos disponibles `13`
  - Lista detallada de Instrumentos disponibles `14`
  - Descripción detallada de un Instrumento `16`
  - Lista de Instrumentos por Código CFI `17`
  - Lista de Instrumentos por Segmento `19`
- **Orden** `20`
  - Ingresar una orden `20`
  - Ingresar una orden a través de WebSocket `23`
  - Reemplazar una orden `25`
  - Cancelar una orden `26`
  - Ver el estado de una orden `27`
  - Consultar último estado por Client Order ID `27`
  - Consultar todos los estados por Client Order ID `29`
  - Consultar Order por OrderID `30`
  - Consultar Ordenes Operadas `31`
  - Estado de orden por ID Cuenta `32`
  - Estado de orden por Execution ID `34`
  - Suscribirse a Execution Reports a través de WebSocket `35`
  - Mensaje para Execution Reports `36`
- **Market Data** `37`
  - MarketData en tiempo real a través de REST `38`
  - Suscribirse a MarketData en tiempo real a través de WebSocket `39`
  - Mensaje de Market Data `40`
  - Descripción de MarketData Entries `41`
  - MarketData Histórica `42`
- **Primary Risk API** `43`
  - Consultar las posiciones de una cuenta `43`
  - Consultar detalle de posiciones `46`
  - Consultar reporte de cuenta `48`
- **Anexo - Errores** `50`
- **Anexo - Diccionario de Campos** `51`



---

## Historia del documento

| Fecha | Versión | Descripción | Autor |
| :--- | :--- | :--- | :--- |
| 03/11/2015 | 1.0 | - Versión inicial | Primary |
| 10/11/2015 | 1.1 | - Se agregó una introducción sobre Primary API.<br>- Se agregó una breve descripción de como conectarse al Web Service mediante un Login web. | Primary |
| 16/11/2015 | 1.2 | - Corrección y modificación gramatical.<br>- Incorporación de métodos:<br>  - Lista de segmentos disponibles<br>  - Lista detallada de todos Instrumentos<br>  - Lista detallada de un instrumento<br>  - Lista de Instrumentos por CFICode | Primary |
| 20/11/2015 | 1.3 | - Se agregó un anexo con información de Primary Risk API | Primary |
| 14/12/2015 | 1.4 | - Se agregó información de nuevos métodos en primary Risk API:<br>  - Consulta de posiciones para una cuenta<br>  - Consulta de posiciones abiertas para una cuenta<br>  - Consulta de reporte de cuenta<br>  - Consulta de reportes de cuentas<br>  - Consulta de detalle de posición<br>  - Consulta de integridad de datos<br>  - Consulta de importaciones realizadas | Primary |
| 03/02/2016 | 1.5 | - Se agregó descripción y ejemplos de la API Web Socket.<br>- Tabla descriptiva de market data entries disponibles para consultar. | Primary |
| 16/02/2016 | 1.6 | - Se modificó la información disponible en Risk API. Solamente se muestra información de los servicios accesibles a través de Primary API. | Primary |
| 23/02/2016 | 1.7 | - Nueva funcionalidad para soportar profundidad en la market data tanto para la API Rest y Web Socket.<br>- Se agregó el campo vencimiento al detalle de instrumentos.<br>- API para listar instrumentos por segmento.<br>- API para listar ordenes por id.<br>- API para listar ordenes activas.<br>- API para listar ordenes operadas.<br>- Descripción del campo "level" en mensaje de suscripción a market data para Web Socket. | Primary |
| 20/04/2016 | 1.8 | - Se agregó un parámetro más al método de ingreso de ordenes para permitir indicar si se cancelan las ordenes previas o no.<br>- Se actualizó la respuesta del Reporte de Cuentas para dar soporte multimoneda. | Primary |
| 05/07/2016 | 1.9 | - Se actualizó la respuesta del Reporte de Cuentas para dar soporte al cambio en la estructura del JSON.<br>- Se agregaron los métodos para obtener los Execution Reports de las ordenes activas (NEW y PARTIALLY_FILLED) por Web Socket. | Primary |
| 16/02/2017 | 1.10 | - Se agregó el parámetro "snapshotOnlyActive” para obtener los Execution Reports de las ordenes activas (NEW y PARTIALLY_FILLED) por Web Socket. | Primary |
| 22/02/2017 | 1.11 | - Se agregó el parámetro Iceberg y displayQty para ordenes Iceberg. Se agregó el timeInForce = GTD (Good Till Date) y el parámetro expireDate para ordenes GTD. | Primary |
| 10/03/2017 | 1.12 | - Se agregaron los mensajes Web Socket para el envío de ordenes Iceberg y GTD. | Primary |
| 14/03/2017 | 1.13 | - Se reemplazó el método de logun de la API. | Primary |
| 17/03/2017 | 1.14 | - Se agregó el método Replace por clOrdId para reemplazar órdenes ingresadas al mercado. | Primary |
| 14/06/2017 | 1.15 | - Se agregó la forma autenticación por token. | Primary |
| 30/08/2017 | 1.16 | - Se agregó el método para consultar las ordenes por Execution ID. | Primary |
| 18/09/2017 | 1.17 | - Se agregó links a la documentación Swagger.<br>- Se incluyen links a ejemplos de login Rest y Websocket en Python. | Primary |
| 11/01/2018 | 1.18 | - Se agregaron los parámetros dateFrom y dateTo en el método de MarketData Histórica (getTrades) para consultar por rango de fecha. | Primary |
| 27/03/2018 | 1.19 | - Se agregó el método de consulta de estado de orden por OrdenID. | Primary |
| 21/09/2018 | 1.20 | - Se agregó el campo wsCIOrdId para poder identificar las ordenes enviadas por WebSocket.<br>- Se agregó el parámetro allOrNone para poder enviar ofertas para contratos mayoristas por WebSocket. | Primary |
| 02/12/2022 | 1.21 | - Se actualiza documentación de la API con nuevo formato y con la información mejor estructurada.<br>- Se agregan links a herramientas, conectores del ecosistema y webinars desarrollados por Primary para facilitar la integración con las APIs.<br>- Se agrega un diccionario de campos al final del documento para tener una referencia de cada campo que devuelve la API. | Primary |



---

## Introducción

Primary API es una solución creada para facilitar la interoperabilidad de software de terceros con la plataforma de negociación electrónica de MATBA ROFEX, Primary Trading Platform (PTP).

El valor agregado de esta solución consiste en simplificar la integración con PTP traduciendo las interfaces basadas en el estándar FIX a un conjunto de Web Services simplificados; estándar tecnológico más adoptado por el mercado y cuya implementación requiere un menor esfuerzo de desarrollo con la consiguiente ventaja en la reducción del costo asociado.

Los Web Services tipo REST trabajan de forma sincrónica y fueron diseñados y optimizados para atender consultas como por ejemplo: ingresar/cancelar una orden, ver Market Data histórica, etc.

Se encuentran disponibles distintas herramientas desarrolladas por el ecosistema, para facilitar aún más las pruebas de integración con la API. A continuación se listan:

- **[APIDoc](https://api.remarkets.primary.com.ar/api-docs/index.html)**
- **[Cliente WebSocket Web](https://www.primary.com.ar/web-socket/)**
- **[Repositorio de Github](https://github.com/matbarofex)**
- **[Repositorio de Postman](https://www.postman.com/primary-com-ar/workspace/primary-api/overview)**
- **[Guía de implementación de MarketData](https://www.youtube.com/watch?v=kYJ5KQiZtT4)**
- **[Guía de implementación de MarketData y OrderRouting](https://www.youtube.com/watch?v=QOzC8yUBR_g)**
- **[Guía de homologación y buenas prácticas](https://www.youtube.com/watch?v=Fj2F23u3d0w)**
- **[Google Colab (Webinar)](https://www.youtube.com/watch?v=uK481KxSTc8)**
- **[Buenas prácticas de consumo](https://www.youtube.com/watch?v=LqUf7K04K-k)**

Adicionalmente, Primary desarrolló y desarrolla distintos webinarios para explicar las funcionalidades de la API y las herramientas creadas para facilitar las pruebas con la API:

- **[Webinar APIs y herramientas del ecosistema Primary para acceso al mercado de Capitales - Grupo MtR](https://www.youtube.com/watch?v=LqUf7K04K-k)**
- **[Implementación de Ruteo de Ofertas y Market Data via GColab con Primary xOMS](https://www.youtube.com/watch?v=uK481KxSTc8)**
- **[Postman Collections en Primary API](https://www.youtube.com/watch?v=r087kI65T3w)**
- **[Webinar: Repositorio POSTMAN](https://www.youtube.com/watch?v=r087kI65T3w)**

¡Te invitamos a sumarte a nuestro MeetUp para enterarte de cada nuevo evento que desarrollemos!
[https://meetup.com/es-ES/PMY-Eventos/](https://meetup.com/es-ES/PMY-Eventos/)

Por último, Primary cuenta brinda soporte gratuito durante el día, por lo que cualquier consulta que tengas, estamos a disposición para ayudarte.

¡Envianos un correo a [mpi@primary.com.ar](mailto:mpi@primary.com.ar) con tus dudas!



---

### Esquema de la solución:



### Conectándose a la API por token de autenticación

A continuación se muestra un ejemplo de conexión a la API, primero indicaremos cómo solicitar el token y luego la respuesta que nos devuelve el request.

Para obtener el Token, deberá hacer una petición POST a la siguiente URL: `https://api.remarkets.primary.com.ar/auth/getToken`.

Enviando los siguientes Headers:
*   X-Username
*   X-Password

Una vez enviada la solicitud, dicho request va a retornar una serie de headers, y entre ellos recibirá el siguiente header con el token de acceso:
*   X-Auth-Token

Ejemplo del Token de Acceso que se obtiene por Postman:

`X-Auth-Token - 8LCfQCvFjj7SGRicXT2HaursZ/4cltAhWL9o0tAjD2U=`

Una vez haya obtenido el token, deberá utilizarlo para realizar los request necesarios hacia los otros métodos. Es importante tener en cuenta que la duración del token es de 24 hs.

A continuación se especifica link con las iniciativas opensource para conectarse al mercado Matba Rofex. Link: [https://github.com/matbarofex/](https://github.com/matbarofex/)



---

## Consultas a la API REST de Trading

El objetivo de esta sección es brindarle al desarrollador una descripción detallada de los métodos disponibles en Primary API.

Se da una breve explicación de la información que devuelve cada método, los parámetros que acepta, un ejemplo de cómo utilizarlo y la respuesta que devuelve la API. Por último se agregó una lista de errores posibles que pueden ocurrir al utilizar la API.

### Funcionalidades de la API REST

Primary API permite realizar lo siguiente:

**Segmentos¹:**
*   Lista de segmentos disponibles Instrumentos
*   Lista de Instrumentos disponibles
*   Lista Detallada de Instrumentos disponibles
*   Descripción detallada de un Instrumento
*   Lista de Instrumentos por Código CFI

**Órdenes:**
*   Ingresar una orden
*   Cancelar una orden
*   Ver estado de una orden
    *   Todos los estados
    *   El último estado
    *   Estado de todas las ordenes llenas
    *   Estado de todas las ordenes activas
    *   Todas las orden asociadas a una cuenta

**Market Data:**
*   Datos en tiempo real
*   Trades históricos para un rango de fechas

---

### Funcionalidades de la API WebSocket

Primary API con la tecnología Web Socket permite realizar lo siguiente:

**Órdenes:**
*   Ingresar una orden
*   Cancelar una orden
*   Suscribirse al Execution Report de las ordenes
    *   Una cuenta
    *   Varias Cuentas
    *   Todas las cuentas

**Market Data:**
*   Suscribirse a Market Data en tiempo real.



---

## Segmentos

### Lista de Segmentos disponibles

Los segmentos son los distintos ambientes o ruedas de negociación en las que está organizada la operatoria de MATBA ROFEX. Algunos de los segmentos que existen hoy en día son:
*   DDF (Instrumentos de la División Derivados Financieros)
*   DDA (Instrumentos de la División Derivados Agropecuarios)
*   DUAL (Instrumentos listados en ambas divisiones)
*   MERV (Instrumentos de mercados externos a Matba Rofex)

**HTTP request**

`GET https://api.remarkets.primary.com.ar/rest/segment/all`

**Parámetros**

No recibe parámetros.

**Ejemplo**

| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/segment/all` |
| **Respuesta** |
```json
{
  "status": "OK",
  "segments": [
    {
      "marketSegmentId": "DDA",
      "marketId": "ROFX"
    },
    {
      "marketSegmentId": "TEST",
      "marketId": "ROFX"
    },
    {
      "marketSegmentId": "DUAL",
      "marketId": "ROFX"
    },
    {
      "marketSegmentId": "MERV",
      "marketId": "ROFX"
    }
  ]
}
```
|

---

## Instrumentos (Securities)

Instrumentos hace referencia a los distintos activos disponibles para negociarse en MATBA ROFEX. Pueden agruparse por Tipos (los cuales se identifican con el estándar CFI Code) y, a su vez, cada uno tiene características propias, las cuales se pueden consultar pidiendo el detalle (tickSize, maxSize, minSize, etc.) del mismo.

### Lista de Segmentos disponibles

Este método nos devuelva una lista con todos los instrumentos disponibles para negociarse en MATBA ROFEX. Por cada instrumento devuelve el símbolo, ID del mercado al que pertenece y el código CFI del instrumento.

**HTTP request**

`GET https://api.remarkets.primary.com.ar/rest/instruments/all`

**Parámetros**

No recibe parámetros.

**Ejemplo**

| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/instruments/all` |
| **Respuesta** |
```json
{
  "status": "OK",
  "instruments": [
    {
      "instrumentId": {
        "marketId": "ROFX",
        "symbol": "TRI.ROS/DIC23 352 C"
      },
      "cficode": "OCAFXS"
    },
    {
      "instrumentId": {
        "marketId": "ROFX",
        "symbol": "DLR/DIC22"
      },
      "cficode": "FXXXSX"
    }
  ]
}
```
|

### Lista detallada de Instrumentos disponibles

Al igual que el método anterior, éste devuelve una lista con todos los instrumentos pero se agrega una descripción detallada de cada uno de ellos. Por cada instrumento devuelve datos de segmento, precio mínimo/máximo, vencimiento, etc.

**HTTP request**

`GET https://api.remarkets.primary.com.ar/rest/instruments/details`

**Parámetros**

No recibe parámetros.

**Ejemplo**
| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/instruments/details` |
| **Respuesta** |
```json
{
  "status": "OK",
  "instruments": [
    {
      "symbol": null,
      "segment": {
        "marketSegmentId": "DDA",
        "marketId": "ROFX"
      },
      "lowLimitPrice": 0.1,
      "highLimitPrice": 100,
      "minPriceIncrement": 0.1,
      "minTradeVol": 1,
      "maxTradeVol": 20,
      "tickSize": 1,
      "contractMultiplier": 100,
      "roundLot": 1,
      "priceConvertionFactor": 1,
      "maturityDate": "20231123",
      "currency": "USD",
      "orderTypes": [
        "STOP LIMIT",
        "MARKET_TO_LIMIT",
        "LIMIT"
      ],
      "timesInForce": [
        "FOK",
        "IOC",
        "DAY",
        "GTD"
      ],
      "securityType": null,
      "settlType": null,
      "instrumentPricePrecision": 1,
      "instrumentSizePrecision": 0,
      "securityId": null,
      "securityIdSource": null,
      "securityDescription": "TRI.ROS/DIC23 352 C",
      "tickPriceRanges": {
        "0": {
          "lowerLimit": 0,
          "upperLimit": null,
          "tick": 0.1
        }
      },
      "cficode": "OCAFXS",
      "instrumentId": {
        "marketId": "ROFX",
        "symbol": "TRI.ROS/DIC23 352 C"
      }
    },
    {
      "symbol": null,
      "segment": {
        "marketSegmentId": "DDF",
        "marketId": "ROFX"
      },
      "lowLimitPrice": 161.2,
      "highLimitPrice": 201.2,
      "minPriceIncrement": 0.05,
      "minTradeVol": 1,
      "maxTradeVol": 10000,
      "tickSize": 1,
      "contractMultiplier": 1000,
      "roundLot": 1,
      "priceConvertionFactor": 1,
      "maturityDate": "20221230",
      "currency": "ARS",
      "orderTypes": [
        "STOP LIMIT",
        "MARKET TO LIMIT",
        "MARKET",
        "LIMIT"
      ],
      "timesInForce": [
        "IOC",
        "DAY",
        "GTD"
      ],
      "securityType": null,
      "settlType": null,
      "instrumentPricePrecision": 2,
      "instrumentSizePrecision": 0,
      "securityId": null,
      "securityIdSource": null,
      "securityDescription": "DLR/DIC22",
      "tickPriceRanges": {
        "0": {
          "lowerLimit": 0,
          "upperLimit": null,
          "tick": 0.05
        }
      },
      "cficode": "FXXXSX",
      "instrumentId": {
        "marketId": "ROFX",
        "symbol": "DLR/DIC22"
      }
    }
  ]
}
```
|

### Descripción detallada de un Instrumento

Este método devuelve una descripción detallada de un solo instrumento, el cual se especifica en los parámetros.

**HTTP request**

`GET https://api.remarkets.primary.com.ar/rest/instruments/detail`

**Parámetros URL**

| Parameter Name | Value | Description |
| :--- | :--- | :--- |
| marketId | String | El id del Mercado en el que se va a cargar la orden. Valores permitidos: `ROFX`. Valido para ordenes nativas a Matba Rofex y ordenes a mercados externos. |
| Symbol | String | El símbolo del instrumento. Ejemplo: DLR/DIC23 - para futuro de dólar vencimiento Diciembre 2023. |

**Ejemplo**

| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/instruments/detail?symbol=DLR/NOV23&marketId=ROFX` |
| **Respuesta** |
```json
{
  "status": "OK",
  "instrument": {
    "symbol": null,
    "segment": {
      "marketSegmentId": "DDF",
      "marketId": "ROFX"
    },
    "lowLimitPrice": 321,
    "highLimitPrice": 370,
    "minPriceIncrement": 0.05,
    "minTradeVol": 1,
    "maxTradeVol": 10000,
    "tickSize": 1,
    "contractMultiplier": 1000,
    "roundLot": 1,
    "priceConvertionFactor": 1,
    "maturityDate": "20231130",
    "currency": "ARS",
    "orderTypes": [
      "STOP LIMIT",
      "MARKET TO LIMIT",
      "MARKET",
      "LIMIT"
    ],
    "timesInForce": [
      "IOC",
      "DAY",
      "GTD"
    ],
    "securityType": null,
    "settlType": null,
    "instrumentPricePrecision": 2,
    "instrumentSizePrecision": 0,
    "securityId": null,
    "securityIdSource": null,
    "securityDescription": "DLR/NOV23",
    "tickPriceRanges": {
      "0": {
        "lowerLimit": 0,
        "upperLimit": null,
        "tick": 0.05
      }
    },
    "cficode": "FXXXSX",
    "instrumentId": {
      "marketId": "ROFX",
      "symbol": "DLR/NOV23"
    }
  }
}
```
|

### Lista de Instrumentos por Código CFI

Este método permite listar todos los instrumentos que pertenezcan al mismo Tipo (como dijimos anteriormente cada tipo se identifica por un código CFI).

**HTTP request**

`GET https://api.remarkets.primary.com.ar/rest/instruments/byCFICode`

**Parámetros URL**

| Parameter Name | Value | Description |
| :--- | :--- | :--- |
| CFICode | String | Codigo que identifica al Tipo de instrumento. Valores posibles:<br>- ESXXXX > Accion<br>- DBXXXX > Bono<br>- OCASPS > Opción Call sobre Acción<br>- OPASPS > Opción Put sobre Acción<br>- FXXXSX > Futuro<br>- OPAFXS > Opción Put sobre Futuro<br>- OCAFXS > Opción Call sobre Futuro<br>- EMXXXX > CEDEAR<br>- DBXXFR > Obligaciones Negociables |

**Ejemplo**

| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/instruments/byCFICode?CFICode=FXXXSX` |
| **Respuesta** |
```json
{
  "status": "OK",
  "instruments": [
    {
      "marketId": "ROFX",
      "symbol": "DLR/DIC23"
    },
    {
      "marketId": "ROFX",
      "symbol": "TRI.ROS/DIC23"
    },
    {
      "marketId": "ROFX",
      "symbol": "SOJ.QQ/ENE23"
    },
    {
      "marketId": "ROFX",
      "symbol": "SOJ.ROS/DIC22"
    }
  ]
}
```
|

### Lista de Instrumentos por Segmento

Este método permite listar todos los instrumentos que pertenezcan al mismo Segmento de Mercado (los segmentos habilitados en el mercado de se encuentran disponibles en la API de segmentos `https://api.remarkets.primary.com.ar/rest/segment/all`).

**HTTP request**

`GET https://api.remarkets.primary.com.ar/rest/instruments/bySegment`

**Parámetros URL**

| Parameter Name | Value | Description |
| :--- | :--- | :--- |
| MarketSegmentID | String | Codigo que identifica al Segmento. Valores posibles: DDF, DDA, DUAL, U-DDF, U-DDA, U-DUAL, MERV |
| MarketID | String | ID del Mercado al que pertenece el segmento. Valores permitidos: ROFX |

**Ejemplo**

| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/instruments/bySegment?MarketSegmentID=DDF&MarketID=ROFX` |
| **Respuesta** |
```json
{
  "status": "OK",
  "instruments": [
    {
      "marketId": "ROFX",
      "symbol": "DLR/DIC22 186 C"
    },
    {
      "marketId": "ROFX",
      "symbol": "DLR/AGO23"
    },
    {
      "marketId": "ROFX",
      "symbol": "DLR/DIC22 188 C"
    },
    {
      "marketId": "ROFX",
      "symbol": "DLR/DIC22/FEB23 A"
    },
    {
      "marketId": "ROFX",
      "symbol": "DLR/ABR23/MAY23"
    }
  ]
}
```
|

---

## Orden

### Ingresar una orden

Es la manera de enviar una orden al Mercado. Vale la pena aclarar que cuando se ingresa una orden hay que verificar si realmente se cargó la orden y no fue rechazada. La secuencia debería ser:
1.  Ingresar la orden a través de la API.
2.  Si la respuesta que recibo es un `"status":"OK"`, entonces debería pedirle a la API el estado de la orden creada (utilizando el ID devuelto al ingresar la orden. Ej: `"clientId":"user14472764..."`). En este momento la orden puede tener 2 estados:
    2.1 `"status":"NEW"` = la orden se ingresó correctamente
    2.2 `"status":"REJECTED"` = la orden fue rechazada. La razón se especifica en la respuesta.

**HTTP request**

`GET https://api.remarkets.primary.com.ar/rest/order/newSingleOrder`

**Parámetros URL**

| Parameter Name | Value | Description |
| :--- | :--- | :--- |
| marketId | String | El id del Mercado en el que se va a cargar la orden. Valores permitidos: ROFX. Valido para órdenes nativas a MATBA ROFEX y ordenes a MERVAL |
| Symbol | String | El símbolo del instrumento. Ejemplo: "DLR/DIC23"- para futuro de dólar vencimiento Diciembre 2023. |
| Price | Float | Precio de la orden. Ejemplo: 12.01 |
| orderQty | Integer | Tamaño de la Orden. Ejemplo: 2400 |
| ordType | String | Tipo de Orden. Valores permitidos: Limit, Market |
| side | String | Lado de la orden. Valores permitidos: Sell, Buy |
| timeInForce | String | Modificador de la orden que define el lapso de tiempo en el que la orden estará activa, por defecto Day. Valores permitidos: Day: Orden solo valida por el día. Al cerrar la rueda se expira; IOC: Immediate or Cancel; FOK: Fill or Kill; GTD: Good Till Date (Es necesario completar el campo expireDate). |
| account | Integer | Número de cuenta. Ejemplo: 30 |
| cancelPrevious | Boolean | Parámetro opcional que indica si se cancelan las órdenes enviadas previamente para el mismo contrato, cuenta y lado. |
| iceberg | Boolean | Parámetro opcional que indica el tipo de orden Iceberg. Por defecto este valor es false. |
| expireDate | Date | Parámetro opcional que indica la fecha de Vencimiento para ordenes GTD. Ejemplo: 20230720 |
| displayQty | Integer | Parámetro opcional que indica la cantidad a divulgar para ordenes Iceberg. |

**Ejemplo**

| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/order/newSingleOrder?marketId=ROFX&symbol=DLR/DIC23&side=BUY&timeInForce=DAY&orderQty=100&ordType=LIMIT&account=REM6771&cancelPrevious=False&iceberg=False&price=210.5` |
| **Respuesta** |
```json
{
  "status": "OK",
  "order": {
    "clientId": "21581341758",
    "proprietary": "PBCP"
  }
}
```
|

**Ejemplo Iceberg**

| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/order/newSingleOrder?marketId=ROFX&symbol=DLR/DIC23&side=BUY&timeInForce=DAY&orderQty=100&ordType=LIMIT&account=REM6771&cancelPrevious=False&iceberg=true&price=210.5&displayQty=5` |
| **Respuesta** |
```json
{
  "status": "OK",
  "order": {
    "clientId": "user125469825632595",
    "proprietary": "PBCP"
  }
}
```
|

**Ejemplo GTD**

| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/order/newSingleOrder?marketId=ROFX&symbol=DLR/DIC23&side=BUY&timeInForce=GTD&orderQty=100&ordType=LIMIT&account=REM2747&cancelPrevious=False&iceberg=False&price=182.5&expireDate=20230505` |
| **Respuesta** |
```json
{
  "status": "OK",
  "order": {
    "clientId": "utfa3256548752365489",
    "proprietary": "api"
  }
}
```
|

### Ingresar una orden a través de WebSocket

Con este mensaje se envía una orden al mercado. Para poder saber que ocurrió con la orden hay que estar suscripto a los Execution Report para la cuenta con la que mandamos la orden, de lo contrario no recibiremos ningún mensaje sobre el estado de la orden.

**Mensaje enviado**
```json
{
  "type":"no",
  "product":{
    "marketId":"ROFX",
    "symbol":"DLR/DIC23"
  },
  "price":185,
  "quantity":23,
  "side": "BUY",
  "account":"20",
  "iceberg":false
}
```

**Mensaje para el envío de una orden para un contrato Todo o Nada**
```json
{
  "type":"no",
  "product":{
    "marketId":"ROFX",
    "symbol":"DLR/DIC23A"
  },
  "price":185,
  "quantity":3000,
  "side": "BUY",
  "account":"20",
  "iceberg":false
}
```

**Mensaje para el envío de una orden con identificador (wsClOrdId):**
```json
{
  "type":"no",
  "product":{
    "marketId":"ROFX",
    "symbol":"DLR/DIC23"
  },
  "price":185,
  "quantity":23,
  "side":"BUY",
  "account":"20",
  "iceberg":false,
  "wsClOrdId":"asdjuej213n1"
}
```

**Mensaje de respuesta al envío de una orden con identificador (wsClOrdId):**
```json
{
  "type": "or",
  "timestamp": 1537212212623,
  "orderReport": {
    "orderId": null,
    "clOrdId": "275772212000001",
    "proprietary": "test",
    "accountId": {
      "id": "100"
    },
    "instrumentId": {
      "marketId": "ROFX",
      "symbol": "DLR/DIC23"
    },
    "price": 40,
    "orderQty": 1,
    "ordType": "LIMIT",
    "side": "BUY",
    "timeInForce": "DAY",
    "transactTime": "20180917-16:23:32",
    "status": "PENDING_NEW",
    "text": "Enviada",
    "wsClOrdId": "12345678fz"
  }
}
```
**Comentarios**
El campo `wsCIOrdId` se utiliza para identificar la orden enviada. Este campo va a venir solamente en el primer Execution Report (con estado `PENDING_NEW` o `REJECT`). En el primer Execution Report recibido el campo `wsClOrdId` se debe referenciar con el campo `clOrdId` para poder seguir los diferentes estados de la orden. El campo `wsClOrdId` es opcional. Si no se envía se debe tomar el primer `clOrdId` recibido en el Execution Report para identificar la orden. El usuario debe asegurarse que el ID ingresado le permita identificar la orden. La API no valida que el ID sea único.

**Mensaje para el envío de una orden Iceberg vía WebSocket:**
```json
{
  "type":"no",
  "product":{
    "marketId":"ROFX",
    "symbol": "DLR/DIC22"
  },
  "price":185,
  "quantity":20,
  "side":"BUY",
  "account":"20",
  "iceberg":true,
  "displayQuantity":8
}
```

**Mensaje para el envío de una orden GTD:**
```json
{
  "type":"no",
  "product":{
    "marketId":"ROFX",
    "symbol":"DLR/DIC23"
  },
  "price":185,
  "quantity":23,
  "side": "BUY",
  "account":"20",
  "timeInForce":"GTD",
  "expireDate":"20231010"
}
```

### Reemplazar una orden

Permite reemplazar una orden ingresada al mercado.

**HTTP request**

`GET https://api.remarkets.primary.com.ar/rest/order/replaceById`

**Parámetros URL**

| Parameter Name | Value | Description |
| :--- | :--- | :--- |
| clOrdId | String | El ID del request que se hizo al Mercado, devuelto por la API de Ingreso y Cancelación de Orden. |
| proprietary | String | ID que identifica al participante del mercado mediante el cual se hace el request. Este parámetro es siempre fijo para una cuenta. |
| orderQty | Integer | Tamaño a modificar de la orden. |
| Price | Float | Precio a modificar de la orden. |

**Ejemplo**
| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/order/replaceById?clOrdId=user144733478280357&proprietary=api&price=17&orderQty=10` |
| **Respuesta** |
```json
{
  "status": "OK",
  "order": {
    "clientId": "user14473450286174Cn15",
    "proprietary": "PBCP"
  }
}
```
|

### Cancelar una orden

Funcionalidad que permite cancelar una orden ingresada en el mercado.

**HTTP request**

`GET https://api.remarkets.primary.com.ar/rest/order/cancelById`

**Ejemplo**
| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/order/cancelById?clOrdId=ajduj3l13ieci2jr4ck&proprietary=PBCP` |
| **Respuesta** |
```json
{
  "status": "OK",
  "order": {
    "clientId": "user14473450286174Cn15",
    "proprietary": "PBCP"
  }
}
```
|

### Cancelar una Orden a través de WebSocket

Mensaje que permite cancelar una orden ingresada en el mercado vía WebSocket.

**Mensaje enviado:**
```json
{
  "type":"co",
  "clientId":"user114121092035207",
  "proprietary":"PBCP"
}
```

### Ver el estado de una orden

APIs para consultar en qué estado se encuentra una orden que se ingresó. Se brindan varias alternativas para esto, se puede consultar por el Client Order ID (explicado más adelante), por ordenes activas u operadas, etc.

Es importante dejar en claro algunos conceptos para poder entender correctamente tanto la consulta a la API como los datos que nos devuelve. Dos términos que suelen confundirse pero que representan cosas distintas son el **Cliente Order ID** y el **Order ID**:
*   **Cliente Order ID**: hace referencia al ID del request que se hace al mercado. Para que quede claro, un ejemplo sería: cuando se quiere mandar dar de alta una orden en el mercado se hace un request de alta de orden que lleva asociado un Cliente Order ID específico, esto permite diferenciar distintos request que se hagan. Para el mercado un request de alta de orden y un request de cancelación de una orden son distintos aunque tengan asociado el mismo ID de orden (ver siguiente), entonces tendríamos un client order ID al ingresar una orden y un client order ID distinto al cancelar esa orden y ambos van a estar asociados al mismo Order ID.
*   **Order ID**: identifica una orden en el mercado, tan simple como esto, como dijimos antes un Order ID va a estar asociado a un request de alta de orden (client order ID). Si se manda otro request para cancelar esa orden entonces también ese order ID va a estar asociado al request de cancelación de orden (client order ID distinto al de alta de orden).

Si vemos lo que devuelve la API de ingreso de orden y la API de cancelación de orden encontramos un campo de la respuesta llamado `clOrdId` (Cliente Order ID) con este dato es con el que vamos a poder realizar varias de las consultas hechas a la API.

### Consultar último estado por Client Order ID

Esta consulta permite ver el estado de ese request, este se identifica con un client order ID.

**HTTP request**

`GET https://api.remarkets.primary.com.ar/rest/order/id`

**Parámetros URL**
| Parameter Name | Value | Description |
| :--- | :--- | :--- |
| clOrdId | String | El ID del request que se hizo al Mercado, devuelto por la API de Ingreso y Cancelación de Orden. |
| proprietary | String | ID que identifica al participante del mercado mediante el cual se hace el request. Este parámetro es siempre fijo para una cuenta. |

**Ejemplo**
| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/order/id?clOrdId=user1144720678549411&proprietary=api` |
| **Respuesta** |
```json
{
  "status": "OK",
  "order": {
    "orderId": "1130835",
    "clOrdId": "user1145712381052053",
    "proprietary": "PBCP",
    "execId": "160229133429-fix1-493",
    "accountId": {
      "id": "10"
    },
    "instrumentId": {
      "marketId": "ROFX",
      "symbol": "DLR/DIC23"
    },
    "price": 183,
    "orderQty": 10,
    "ordType": "LIMIT",
    "side": "BUY",
    "timeInForce": "DAY",
    "transactTime": "20160304-17:37:35",
    "avgPx": 0,
    "lastPx": 0,
    "lastQty": 0,
    "cumQty": 0,
    "leavesQty": 10,
    "status": "NEW",
    "text": "Aceptada"
  }
}
```
|

### Consultar todos los estados por Client Order ID

Consulta que devuelve todos los estados por los que paso una orden que estén asociados a un request hecho al mercado. En el caso de un request de alta de orden devolverá todos los estados de la orden que se ingresó al mercado menos el del estado Cancelled ya este estado estará asociado a un request de cancelación y no de alta de orden.

**HTTP request**

`GET https://api.remarkets.primary.com.ar/rest/order/allById`

**Parámetros URL**
| Parameter Name | Value | Description |
| :--- | :--- | :--- |
| clOrdId | String | El ID del request que se hizo al Mercado, devuelto por la API de Ingreso y Cancelación de Orden. Ejemplo: user144720678549411 |
| proprietary | String | ID que identifica al participante del mercado mediante el cual se hace el request. Este parámetro es siempre fijo para una cuenta. Valor: api |

**Ejemplo**
| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/order/allById?clOrdId=user1144720678549411&proprietary=api` |
| **Respuesta** |
```json
{
  "status": "OK",
  "orders": [
    {
      "orderId": null,
      "clOrdId": "user1145712381052053",
      "proprietary": "PBCP",
      "accountId": {
        "id": "10"
      },
      "instrumentId": {
        "marketId": "ROFX",
        "symbol": "DLR/DIC23"
      },
      "price": 189.54,
      "orderQty": 10,
      "ordType": "LIMIT",
      "side": "BUY",
      "timeInForce": "DAY",
      "transactTime": "20160304-17:36:50",
      "status": "PENDING_NEW",
      "text": "Enviada"
    },
    {
      "orderId": "1130835",
      "clOrdId": "user1145712381052053",
      "proprietary": "PBCP",
      "execId": "160229133429-fix1-493",
      "accountId": {
        "id": "10"
      },
      "instrumentId": {
        "marketId": "ROFX",
        "symbol": "DLR/DIC23"
      },
      "price": 189,
      "orderQty": 10,
      "ordType": "LIMIT",
      "side": "BUY",
      "timeInForce": "DAY",
      "transactTime": "20160304-17:37:35",
      "avgPx": 0,
      "lastPx": 0,
      "lastQty": 0,
      "cumQty": 0,
      "leavesQty": 10,
      "status": "NEW",
      "text": "Aceptada"
    }
  ]
}
```
|

### Consultar Orden por OrderID

Consulta que devuelve el estado de una orden consultando por su ID.

**HTTP request**

`GET https://api.remarkets.primary.com.ar/rest/order/byOrderId`

**Parámetros URL**
| Parameter Name | Value | Description |
| :--- | :--- | :--- |
| orderId | Integer | ID de la orden |

**Ejemplo**
| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/order/byOrderId?orderId=1130813` |
| **Respuesta** |
```json
{
  "status": "OK",
  "orders": [
    {
      "orderId": "1130813",
      "clOrdId": "user1145703678429722",
      "proprietary": "PBCP",
      "execId": "160229133429-fix1-414",
      "accountId": {
        "id": "10"
      },
      "instrumentId": {
        "marketId": "ROFX",
        "symbol": "DLR/DIC23"
      },
      "price": 189,
      "orderQty": 1,
      "ordType": "LIMIT",
      "side": "BUY",
      "timeInForce": "DAY",
      "transactTime": "20160303-17:27:08",
      "avgPx": 0,
      "lastPx": 0,
      "lastQty": 0,
      "cumQty": 0,
      "leavesQty": 1,
      "status": "NEW",
      "text": "Aceptada"
    }
  ]
}
```
|

### Consultar órdenes activas

Consulta que devuelve todas las órdenes activas de una cuenta.

**HTTP request**

`GET https://api.remarkets.primary.com.ar/rest/order/actives`

**Parámetros URL**
| Parameter Name | Value | Description |
| :--- | :--- | :--- |
| accountId | Integer | Número de cuenta. Ejemplo: 30 |

**Ejemplo**
| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/order/actives?accountId=10` |
| **Respuesta** |
```json
{
  "status": "OK",
  "orders": [
    {
      "orderId": "1130813",
      "clOrdId": "user1145703678429722",
      "proprietary": "PBCP",
      "execId": "160229133429-fix1-414",
      "accountId": {
        "id": "10"
      },
      "instrumentId": {
        "marketId": "ROFX",
        "symbol": "DLR/DIC23"
      },
      "price": 189,
      "orderQty": 1,
      "ordType": "LIMIT",
      "side": "BUY",
      "timeInForce": "DAY",
      "transactTime": "20160303-17:27:08",
      "avgPx": 0,
      "lastPx": 0,
      "lastQty": 0,
      "cumQty": 0,
      "leavesQty": 1,
      "status": "NEW",
      "text": "Aceptada"
    }
  ]
}
```
|

### Consultar Ordenes Operadas

Consulta que devuelve todas las ordenes que están total o parcialmente operadas.

**HTTP request**

`GET https://api.remarkets.primary.com.ar/rest/order/filleds`

**Parámetros URL**
| Parameter Name | Value | Description |
| :--- | :--- | :--- |
| accountId | Integer | Número de cuenta. Ejemplo: 30 |

**Ejemplo**
| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/order/filleds?accountId=10` |
| **Respuesta** |
```json
{
  "status": "OK",
  "orders": [
    {
      "orderId": "1130813",
      "clOrdId": "user1145703678429722",
      "proprietary": "PBCP",
      "execId": "T1169011",
      "accountId": {
        "id": "10"
      },
      "instrumentId": {
        "marketId": "ROFX",
        "symbol": "DLR/DIC23"
      },
      "price": 189,
      "orderQty": 1,
      "ordType": "LIMIT",
      "side": "BUY",
      "timeInForce": "DAY",
      "transactTime": "20160303-17:29:53",
      "avgPx": 18.540000,
      "lastPx": 18.500,
      "lastQty": 1,
      "cumQty": 1,
      "leavesQty": 0,
      "status": "FILLED",
      "text": "Operada"
    }
  ]
}
```
|

### Estado de orden por ID Cuenta

Consulta que devuelve el último estado de los request (client order ID) asociadas a una cuenta. Es decir, de los si se hizo un request para dar de alta una orden, y luego se hizo otro para darlo de baja entonces esta API devolverá 2 ordenes, una con el ultimo estado asociado al request de alta y otra con el ultimo estado asociado al request de baja.

**HTTP request**

`GET https://api.remarkets.primary.com.ar/rest/order/all`

**Parámetros URL**
| Parameter Name | Value | Description |
| :--- | :--- | :--- |
| accountId | Integer | Número de cuenta. Ejemplo: 30 |

**Ejemplo**
| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/order/all?accountId=30` |
| **Respuesta** |
```json
{
  "status": "OK",
  "orders": [
    {
      "orderId": "1130835",
      "clOrdId": "user1145712381052053",
      "proprietary": "PBCP",
      "execId": "160229133429-fix1-493",
      "accountId": {
        "id": "10"
      },
      "instrumentId": {
        "marketId": "ROFX",
        "symbol": "DLR/DIC23"
      },
      "price": 189,
      "orderQty": 10,
      "ordType": "LIMIT",
      "side": "BUY",
      "timeInForce": "DAY",
      "transactTime": "20160304-17:37:35",
      "avgPx": 0,
      "lastPx": 0,
      "lastQty": 0,
      "cumQty": 0,
      "leavesQty": 10,
      "status": "NEW",
      "text": "Aceptada"
    },
    {
      "orderId": "1130835",
      "clOrdId": "user1145712381052053Cn154",
      "proprietary": "PBCP",
      "execId": "160229133429-fix1-494",
      "accountId": {
        "id": "10"
      },
      "instrumentId": {
        "marketId": "ROFX",
        "symbol": "DLR/DIC23"
      },
      "price": 189,
      "orderQty": 10,
      "ordType": "LIMIT",
      "side": "BUY",
      "timeInForce": "DAY",
      "transactTime": "20160304-17:38:15",
      "avgPx": 0,
      "lastPx": 0,
      "lastQty": 0,
      "cumQty": 0,
      "leavesQty": 0,
      "status": "CANCELLED",
      "text": "Cancelada"
    }
  ]
}
```
|

### Estado de orden por Execution ID

Consulta que devuelve el estado de la orden (por Execution ID). Es decir, esta consulta va a devolver la orden asociada a un Execution ID pudiendo identificar que orden esta involucrada en la respectiva operación.

**HTTP request**

`GET https://api.remarkets.primary.com.ar/rest/order/byExecId`

**Parámetros URL**
| Parameter Name | Value | Description |
| :--- | :--- | :--- |
| execId | String | Identificador de Ejecución, ejemplo: T3567006 |

**Ejemplo**
| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/order/byExecId?execId=T1234567` |
| **Respuesta** |
```json
{
  "status": "OK",
  "orders": [
    {
      "orderId": "125366222",
      "clOrdId": "699154",
      "proprietary": "PBCP",
      "execId": "T1234567",
      "accountId": {
        "id": "2345"
      },
      "instrumentId": {
        "marketId": "ROFX",
        "symbol": "DLR/DIC23"
      },
      "price": 189,
      "orderQty": 50,
      "ordType": "LIMIT",
      "side": "BUY",
      "timeInForce": "DAY",
      "transactTime": "20230830-14:23:35",
      "avgPx": 189,
      "lastPx": 189,
      "lastQty": 50,
      "cumQty": 50,
      "leavesQty": 0,
      "status": "FILLED",
      "text": "Operada "
    }
  ]
}
```
|

### Suscribirse a Execution Reports a través de WebSocket

Los siguientes mensajes le permiten al usuario suscribirse para recibir mensajes de Execution Reports sobre las órdenes asociadas a una cuenta determinada. Se permite suscribirse para ordenes asociadas a una, varias o todas las cuentas del usuario.

**Para una cuenta**

Con este mensaje se podrá recibir los Execution Reports de todas las órdenes ingresadas con la cuenta indicada en el mensaje.
**Mensaje enviado:**
```json
{
  "type":"os",
  "account": {
    "id":"40"
  }
}
```

**Para varias cuentas**

Con este mensaje se podrá recibir los Execution Reports de todas las órdenes ingresadas con las cuentas indicadas en el mensaje.
**Mensaje enviado:**
```json
{
  "type":"os",
  "accounts":[
    {
      "id":"40"
    },
    {
      "id":"4000"
    }
  ]
}
```

**Para todas las cuentas**

Con este mensaje se podrá recibir los Execution Reports de todas las órdenes ingresadas con las cuentas asociadas al usuario.
**Mensaje enviado:**
```json
{
  "type":"os"
}
```

Se encuentra disponible el parámetro “snapshotOnlyActive” para recibir los Execution Reports de las ordenes activas (en estado NEW o PARTIALLY_FILLED) para todas las cuentas, mas de una cuenta o para una única cuenta.
**Mensaje enviado:**
```json
{
  "type":"os",
  "snapshotOnlyActive":true
}
```

### Mensaje para Execution Reports

Este es el mensaje que envía el servidor a todos los que estén suscriptos a los Execution Reports de la orden indicada.
**Mensaje enviado:**
```json
{
  "type": "or",
  "orderReport": {
    "orderId": "1128056",
    "clOrdId": "user14545967430231",
    "proprietary": "PBCP",
    "execId": "160127155448-fix1-1368",
    "accountId": {
      "id": "30"
    },
    "instrumentId": {
      "marketId": "ROFX",
      "symbol": "DLR/DIC23"
    },
    "price": 189,
    "orderQty": 10,
    "ordType": "LIMIT",
    "side": "BUY",
    "timeInForce": "DAY",
    "transactTime": "20230204-11:41:54",
    "avgPx": 0,
    "lastPx": 0,
    "lastQty": 0,
    "cumQty": 0,
    "leavesQty": 10,
    "status": "CANCELLED",
    "text": "Reemplazada"
  }
}
```

---

## Market Data

Estas API permiten acceder a datos Históricos y en Tiempo Real sobre cualquier instrumento negociado en el mercado. También la API ofrece un método para buscar todos los instrumentos que se han operado.
Para cotizaciones en tiempo real será necesario que el consumo se haga a través de Websocket.

### MarketData en tiempo real a través de REST

**HTTP request**

`GET https://api.remarkets.primary.com.ar/rest/marketdata/get`

**Parámetros URL**

| Parameter Name | Value | Description |
| :--- | :--- | :--- |
| marketId | String | El id del Mercado al que pertenece el instrumento sobre el que se pide info. Valores permitidos: ROFX. Valido para MD de MATBA ROFEX y mercados externos. |
| symbol | String | El símbolo del instrumento. Ejemplo: DLR/DIC23- para futuro de dólar vencimiento Diciembre 2023. |
| entries | List | Es una lista separada por comas que indican la información solicitada sobre ese contrato. Valores permitidos: BI: Bid, OF: Offer, LA: Last, último precio operado, OP: Open, precio de apertura, CL: Close, precio de cierre, SE: Settlement, OI: Open Interest, ACP: Precio de cierre del día de la fecha para instrumentos externos a MATBA ROFEX. Ejemplo: BI,OF,LA,OP,CL,SE,OI |
| depth | Integer | Parámetro opcional que indica la profundidad del book que se desea recibir en la market data. Por defecto es 1. |

**Ejemplo**

| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/marketdata/get?marketId=ROFX&symbol=DLR/DIC23&entries=BI,OF,LA,OP,CL,SE,OI&depth=3` |
| **Respuesta** |
```json
{
  "status": "OK",
  "marketData": {
    "SE": {
      "price": 180.3,
      "size": null,
      "date": 1669852800000
    },
    "LA": {
      "price": 179.85,
      "size": 4,
      "date": 1669995044232
    },
    "OI": {
      "price": null,
      "size": 217596,
      "date": 1664150400000
    },
    "OF": [
      {
        "price": 179.8,
        "size": 1000
      },
      {
        "price": 180.35,
        "size": 1000
      }
    ],
    "OP": 180.35,
    "CL": {
      "price": 180.35,
      "size": null,
      "date": 1669852800000
    },
    "BI": [
      {
        "price": 179.75,
        "size": 275
      },
      {
        "price": 178.95,
        "size": 514
      }
    ]
  },
  "depth": 2,
  "aggregated": true
}
```
|

### Suscribirse a MarketData en tiempo real a través de WebSocket

Utilizando el protocolo Web Socket es posible recibir Market Data de los instrumentos especificados de manera asíncrona cuando esta cambie sin necesidad de hacer un request cada vez que necesitemos.

Para recibir este tipo de mensajes hay que suscribirse indicando los instrumentos de los cuales queremos recibir MD. El servidor enviara un mensaje de MD por cada instrumento al que nos suscribimos cada vez que este cambie.

Con este mensaje nos suscribimos para recibir MD de los instrumentos especificados, el servidor solamente enviará los datos especificados en la lista “entries”. El parámetro "depth" indica la profundidad del book que se desea recibir, por defecto se devuelve el top of book, es decir profundidad 1.

**Mensaje enviado:**
```json
{
  "type":"smd",
  "level":1,
  "entries":[
    "OF"
  ],
  "products":[
    {
      "symbol":"DLR/DIC23",
      "marketId": "ROFX"
    },
    {
      "symbol":"SOJ.ROS/MAY23",
      "marketId": "ROFX"
    }
  ],
  "depth":2
}
```

### Mensaje de Market Data

Este es el mensaje que envía el servidor a todos los que estén suscriptos a MD del instrumento indicado. En este caso utilizamos el ejemplo de suscripción a market data con profundidad 2.

**Mensaje recibido:**
```json
{
  "type": "md",
  "instrumentId": {
    "marketId": "ROFX",
    "symbol": "DLR/DIC23"
  },
  "marketData": {
    "OF": [
      {
        "price": 189,
        "size": 21
      },
      {
        "price": 188,
        "size": 13
      }
    ]
  }
}
```

### Descripción de MarketData Entries

A continuación se presentan los datos del mercado que son posibles consultar por medio de las API tanto REST como Web Socket. Al momento de consultar market data es posible indicar qué tipo de market data se quiere recibir, esto normalmente es una lista separada por comas de los siguientes símbolos:

| Símbolo | Significado | Descripción |
| :--- | :--- | :--- |
| BI | BIDS | Mejor oferta de compra en el Book |
| OF | OFFERS | Mejor oferta de venta en el Book |
| LA | LAST | Último precio operado en el mercado |
| OP | OPENING PRICE | Precio de apertura |
| CL | CLOSING PRICE | Precio de cierre de la rueda de negociación anterior |
| SE | SETTLEMENT PRICE | Precio de ajuste (solo para futuros) |
| HI | TRADING SESSION HIGH PRICE | Precio máximo de la rueda |
| LO | TRADING SESSION LOW PRICE | Precio mínimo de la rueda |
| TV | TRADE VOLUME | Volumen operado en contratos/nominales para ese security |
| OI | OPEN INTEREST | Interés abierto (solo para futuros) |
| IV | INDEX VALUE | Valor del índice (solo para índices) |
| EV | TRADE EFFECTIVE VOLUME | Volumen efectivo de negociación para ese security |
| NV | NOMINAL VOLUME | Volumen nominal de negociación para ese security |
| ACP | AUCTION PRICE | Precio de cierre del día corriente |

Tanto el Entry EV como NV solo van a devolver información en el caso de que se utilice para un instrumento de ByMA, para consultar el Volumen de un instrumento MATBA ROFEX deberían incluír el Entry TV.

### MarketData Histórica

La API para acceder a datos históricos del mercado permite consultar los trades que se hayan realizado para un contrato, en una fecha puntual o un rango de fechas.

**HTTP request**

`https://api.remarkets.primary.com.ar/rest/data/getTrades`

**Parámetros URL**

| Parameter Name | Value | Description |
| :--- | :--- | :--- |
| marketId | String | Identificador del Mercado a consultar, valores posibles: ROFX = MATBA ROFEX |
| symbol | String | Símbolo del contrato a consultar. |
| date | String | Fecha a consultar, formato: YYYY-MM-DD. Ejemplo: 2023-05-08. |
| dateFrom | String | Fecha desde, formato: YYYY-MM-DD. Ejemplo: 2023-05-08. |
| dateTo | String | Fecha hasta, formato: YYYY-MM-DD. Ejemplo: 2023-05-09. |
| external | Boolean | Parámetro para consultar información histórica de instrumentos de mercados externos a Matba Rofex, cuando se consulta por un instrumento de otro mercado el valor de este parámetro debe ser "true". Parámetro opcional. Valor por defecto: "false". |
| environment | String | Indica el entorno para cual se consulta la MarketData Histórica, en el ambiente de reMarkets el parámetro debe ser "REMARKETS". |

**Ejemplo**

Para obtener los trades históricos que hayan sucedido el 05/08/2023 para el contrato DLR/DIC23
| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/data/getTrades?marketId=ROFX&symbol=DLR/DIC23&date=2023-08-05` |
| **Respuesta** |
```json
{
  "status": "OK",
  "symbol": "DLR/DIC23",
  "market": "ROFX",
  "trades": [
    {
      "symbol": "DLR/DIC23",
      "servertime": 1659714475948,
      "size": 25,
      "price": 183.6,
      "datetime": "2023-08-05 15:47:55.948"
    },
    {
      "symbol": "DLR/DIC23",
      "servertime": 1659715021103,
      "size": 1,
      "price": 183.56,
      "datetime": "2023-08-05 15:57:01.103"
    }
  ]
}
```
|

---

## Primary Risk API

Primary Risk API es la interfaz que permite comunicarse con el sistema RIMA (Risk Manager).

La API actualmente soporta métodos para:
*   Consultar las posiciones de una cuenta
*   Consultar el detalle de las posiciones de una cuenta

### Consultar las posiciones de una cuenta

**Consultar posiciones de una cuenta**

**HTTP Request**

`GET https://api.remarkets.primary.com.ar/rest/risk/position/getPositions/{accountName}`

*   `{accountName}` = cuenta a consultar.

**Autenticación**

HTTP Basic Authentication

**Parámetros URL**

| Parameter Name | Value | Description |
| :--- | :--- | :--- |
| accountName | String | Nombre de la Cuenta. Ejemplo: 30. |

**Ejemplo**

| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/risk/position/getPositions/REM7374` |
| **Respuesta** |
```json
{
  "status": "OK",
  "positions": [
    {
      "instrument": {
        "symbolReference": "AAPL",
        "settlType": 0
      },
      "symbol": "MERV XMEV AAPL CI",
      "buySize": 5,
      "buyPrice": 3092,
      "sellSize": 0,
      "sellPrice": 0,
      "totalDailyDiff": 0,
      "totalDiff": 15460,
      "tradingSymbol": "MERV XMEV AAPL CI",
      "originalBuyPrice": 0,
      "originalSellPrice": 0
    },
    {
      "instrument": {
        "symbolReference": "SOJ.ROS/MAY23 380 C"
      },
      "symbol": "SOJ.ROS/MAY23 380 C",
      "buySize": 0,
      "buyPrice": 0,
      "sellSize": 2,
      "sellPrice": 383,
      "totalDailyDiff": 72767.56,
      "totalDiff": -3832.44,
      "tradingSymbol": "SOJ.ROS/MAY23 380 C",
      "originalBuyPrice": 0,
      "originalSellPrice": 0
    },
    {
      "instrument": {
        "symbolReference": "AL30",
        "settlType": 0
      },
      "symbol": "MERV XMEV AL30 CI",
      "buySize": 100,
      "buyPrice": 0.5702,
      "sellSize": 0,
      "sellPrice": 0,
      "totalDailyDiff": 7442.98,
      "totalDiff": 7500,
      "tradingSymbol": "MERV XMEV AL30 CI",
      "originalBuyPrice": 0,
      "originalSellPrice": 0
    },
    {
      "instrument": {
        "symbolReference": "DLR022023"
      },
      "symbol": "DLR/FEB23",
      "buySize": 120,
      "buyPrice": 211.4,
      "sellSize": 0,
      "sellPrice": 0,
      "totalDailyDiff": -168000,
      "totalDiff": 25200000,
      "tradingSymbol": "DLR/FEB23",
      "originalBuyPrice": 0,
      "originalSellPrice": 0
    },
    {
      "instrument": {
        "symbolReference": "AL29",
        "settlType": 0
      },
      "symbol": "MERV XMEV AL29 CI",
      "buySize": 1000,
      "buyPrice": 0.54525,
      "sellSize": 0,
      "sellPrice": 0,
      "totalDailyDiff": 36734.75,
      "totalDiff": 37280,
      "tradingSymbol": "MERV XMEV AL29 CI",
      "originalBuyPrice": 0,
      "originalSellPrice": 0
    },
    {
      "instrument": {
        "symbolReference": "SOJ.ROS/ENE23 412 C"
      },
      "symbol": "SOJ.ROS/ENE23 412 C",
      "buySize": 3,
      "buyPrice": 412,
      "sellSize": 0,
      "sellPrice": 0,
      "totalDailyDiff": -121649.71,
      "totalDiff": 1950.29,
      "tradingSymbol": "SOJ.ROS/ENE23 412 C",
      "originalBuyPrice": 0,
      "originalSellPrice": 0
    }
  ]
}
```
|

### Consultar detalle de posiciones

**HTTP Request**

`GET https://api.remarkets.primary.com.ar/rest/risk/detailedPosition/{accountName}`

*   `{accountName}` = cuenta a consultar.

**Parámetros URL**

| Parameter Name | Value | Description |
| :--- | :--- | :--- |
| accountName | String | Nombre de la Cuenta. Ejemplo: 30. |

**Ejemplo**

| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/risk/detailedPosition/REM7374` |
| **Respuesta** |
```json
{
  "status": "OK",
  "detailedPosition": {
    "account": "REM7374",
    "totalDailyDiffPlain": -184777,
    "totalMarketValue": 60240,
    "report": {
      "FUTURE OPTION CALL": {
        "SOJ.ROS/MAY23 380 C": {
          "detailedPositions": [
            {
              "symbolReference": "SOJ.ROS/MAY23 380 C",
              "contractType": "FUTURE_OPTION_CALL",
              "priceConversionFactor": 1,
              "contractSize": 100,
              "marketPrice": 380.5,
              "currency": "USD G",
              "exchangeRate": 167.77,
              "contractMultiplier": 1,
              "totalInitialSize": -2,
              "buyInitialSize": 0,
              "sellInitialSize": 2,
              "buyInitialPrice": 0,
              "sellInitialPrice": 383,
              "totalFilledSize": 0,
              "buyFilledSize": 0,
              "sellFilledSize": 0,
              "buyFilledPrice": 0,
              "sellFilledPrice": 0,
              "totalCurrentSize": -2,
              "buyCurrentSize": 0,
              "sellCurrentSize": 2,
              "detailedDailyDiff": {
                "buyPricePPPDiff": 380,
                "sellPricePPPDiff": 380,
                "totalDailyDiff": -100,
                "buyDailyDiff": 0,
                "sellDailyDiff": -100,
                "totalDailyDiffPlain": -16777,
                "buyDailyDiffPlain": 0,
                "sellDailyDiffPlain": -16777
              }
            }
          ],
          "instrumentInitialSize": -2,
          "instrumentFilledSize": 0,
          "instrumentCurrentSize": -2
        },
        "SOJ.ROS/ENE23 412 C": {
          "detailedPositions": [
            {
              "symbolReference": "SOJ.ROS/ENE23 412 C",
              "contractType": "FUTURE_OPTION_CALL",
              "priceConversionFactor": 1,
              "contractSize": 100,
              "marketPrice": 410,
              "currency": "USD G",
              "exchangeRate": 167.77,
              "contractMultiplier": 1,
              "totalInitialSize": 3,
              "buyInitialSize": 3,
              "sellInitialSize": 0,
              "buyInitialPrice": 412,
              "sellInitialPrice": 0,
              "totalFilledSize": 0,
              "buyFilledSize": 0,
              "sellFilledSize": 0,
              "buyFilledPrice": 0,
              "sellFilledPrice": 0,
              "totalCurrentSize": 3,
              "buyCurrentSize": 3,
              "sellCurrentSize": 0,
              "detailedDailyDiff": {
                "buyPricePPPDiff": 412,
                "sellPricePPPDiff": 412,
                "totalDailyDiff": 0,
                "buyDailyDiff": 0,
                "sellDailyDiff": 0,
                "totalDailyDiffPlain": 0,
                "buyDailyDiffPlain": 0,
                "sellDailyDiffPlain": 0
              }
            }
          ],
          "instrumentInitialSize": 3,
          "instrumentFilledSize": 0,
          "instrumentCurrentSize": 3
        }
      }
    },
    "lastCalculation": 1669996294136
  }
}
```
|

### Consultar reporte de cuenta

**HTTP Request**

`GET https://api.remarkets.primary.com.ar/rest/risk/accountReport/{accountName}`

*   `{accountName}` = cuenta a consultar.

**Parámetros URL**

| Parameter Name | Value | Description |
| :--- | :--- | :--- |
| accountName | String | Nombre de la Cuenta. Ejemplo: 30. |

**Ejemplo**
| | |
| :--- | :--- |
| **Http Request** | `https://api.remarkets.primary.com.ar/rest/risk/accountReport/REM7374` |
| **Respuesta** |
```json
{
  "status": "OK",
  "accountData": {
    "accountName": "REM7374",
    "marketMember": "PrimaryVenture",
    "marketMemberIdentity": "PMYVTR",
    "collateral": 0,
    "margin": 2923811.299985,
    "availableToCollateral": 100202251.700015,
    "detailedAccountReports": {
      "0": {
        "currencyBalance": {
          "detailedCurrencyBalance": {
            "USD MtR": {
              "consumed": 0,
              "available": 0
            },
            "ARS BCRA": {
              "consumed": 0,
              "available": 0
            },
            "ARS": {
              "consumed": 0,
              "available": 100000000
            },
            "EUR": {
              "consumed": 0,
              "available": 0
            },
            "U$S": {
              "consumed": 0,
              "available": 0
            },
            "USD G": {
              "consumed": 0,
              "available": 0
            },
            "USD R": {
              "consumed": 0,
              "available": 0
            },
            "USD C": {
              "consumed": 0,
              "available": 0
            },
            "USD D": {
              "consumed": 0,
              "available": 10000
            }
          }
        },
        "availableToOperate": {
          "cash": {
            "totalCash": 103250600,
            "detailedCash": {
              "USD MtR": 0,
              "ARS BCRA": 0,
              "ARS": 100000000,
              "EUR": 0,
              "U$S": 0,
              "USD G": 0,
              "USD R": 0,
              "USD C": 0,
              "USD D": 10000
            }
          },
          "movements": 0,
          "credit": null,
          "total": 103065823,
          "pendingMovements": 0
        },
        "settlementDate": 1669950000000
      }
    },
    "hasError": false,
    "errorDetail": null,
    "lastCalculation": 1669996836647,
    "portfolio": 60240,
    "ordersMargin": 0,
    "currentCash": 103065823,
    "dailyDiff": -184777,
    "uncoveredMargin": 0
  }
}
```
|

---

## Anexo - Errores

Lista de algunos errores que devuelve la API.

| Tipo de Error | Respuesta |
| :--- | :--- |
| Intentar cargar una orden con una cuenta a la que no se tiene acceso. |
```json
{
  "status": "ERROR",
  "description": "No tiene acceso a la cuenta 30",
  "message": null
}
```
|
| Intentar cancelar una orden que no existe. |
```json
{
  "status": "ERROR",
  "description": "Order user1144733478:api doesn't exist",
  "message": null
}
```
|
| Intentar utilizar un símbolo de instrumento que no exista. |
```json
{
  "status": "ERROR",
  "description": "Product DOEne15:ROFX doesn't exist",
  "message": null
}
```
|
| Error al intentar acceder a un método que no existe o al cual no se tiene acceso. |
```json
{
  "status": "ERROR",
  "message": "Access Denied"
}
```
<br>
```json
{
  "status": "ERROR",
  "description": "Ruta invalida",
  "message": ""
}
```
|

---

## Anexo - Diccionario de Campos

| Campo | Descripción |
| :--- | :--- |
| `accountName` | Nombre de la Cuenta. Ejemplo: 30. |
| `avgPx` | Precio promedio operado de la orden. |
| `cficode` | Código CFI del contrato |
| `clOrdId` | ID del request de una orden. |
| `contractMultiplier` | Tamaño del contrato |
| `cumQty` | Cantidad operada de la orden. |
| `currency` | Moneda de liquidación del contrato |
| `dateFrom` | Fecha desde |
| `datetime` | Fecha y hora en que sucedió el trade. |
| `dateTo` | Fecha hasta |
| `depth` | Profundidad de la MarketData. Valores posibles: 1, 2, 3, 4, 5 |
| `displayQty` | Especifica la cantidad a divulgar para las órdenes iceberg |
| `displayQuantity` | Especifica la cantidad a divulgar de una orden iceberg |
| `execId` | ID de ejecución de una orden. |
| `expireDate` | Especifica la fecha de vencimiento de una orden GTD |
| `highLimitPrice` | Límite máximo de precio para operar el contrato |
| `iceberg` | Especifica si una orden es del tipo Iceberg |
| `id` | En los execution reports, especifica el nombre de la cuenta involucrada en la orden. |
| `instrumentPricePrecision` | Precisión de decimales en el precio. |
| `instrumentSizePrecision` | Precisión de decimales en la cantidad. |
| `lastPx` | Último precio al que se operó la orden. |
| `lastQty` | Última cantidad operada de la orden |
| `leavesQty` | Cantidad remanente de la orden |
| `lowerLimit` | Rango mínimo de precio de un tick de precio dinámico |
| `lowLimitPrice` | Límite mínimo de precio para operar el contrato |
| `marketId` | Identificador del mercado. Valores posibles: ROFX: MATBA ROFEX, MERV: Mercados externos a MATBA ROFEX |
| `marketSegmentId` | Identificador de cada segmento |
| `maturityDate` | Fecha de vencimiento/maduración del contrato |
| `maxTradeVol` | Volumen máximo al que se puede operar el contrato |
| `minPriceIncrement` | Incremento mínimo de precio (tick price) |
| `minTradeVol` | Volumen mínimo al que se puede operar el contrato |
| `orderId` | ID de la orden. Si la orden fue rechazada o todavía no llegó al mercado, este campo es null. |
| `orderQty` | Tamaño de la orden |
| `orderTypes` | Listado de tipos de órdenes soportados por un contrato |
| `ordType` | Tipo de orden. Valores posibles: LIMIT, MARKET, STOP_LIMIT, STOP_LIMIT_MERVAL |
| `price` | En órdenes, corresponde la precio de la misma. En MarketData corresponde al precio del correspondiente entry. |
| `priceConvertionFactor` | Valor por el cual se multiplica el precio de un contrato para llevarlo a precio unitario |
| `proprietary` | Usuario FIX que envió la orden. Si las órdenes se envían vía API REST/WS, el propietario de la orden puede ser "PBCP" o "ISV_PBCP". |
| `roundLot` | Multiplicador |
| `securityDescription` | Descripción del instrumento |
| `securityId` | ID del Security |
| `securityType` | Tipo del Security |
| `segments` | Listado de segmentos |
| `serverTime` | Timestamp en el que sucedió el registro en cuestión. Por ejemplo, en el método getTrades, este campo especifica el momento en que surgió el trade. |
| `settlType` | Tipo de liquidación del contrato |
| `side` | Lado de la orden. Valores posibles: BUY, SELL |
| `size` | En un registro de MarketData, especifica la cantidad en el correspondiente entry. |
| `status` | En una orden, especifica el estado de la misma. Valores posibles: NEW, PENDING_NEW, PENDING_REPLACE, PENDING_CANCEL, REJECTED, PENDING_APPROVAL, CANCELLED, REPLACED |
| `symbol` | Nombre del contrato en el sistema de negociación |
| `text` | Campo de texto que refiere el motivo del estado de la orden. |
| `tick` | Tick de precio |
| `tickPriceRanges` | Listado de ticks de precios dinámicos de cada contrato |
| `tickSize` | Incremento de cantidad |
| `timeInForce` | Tiempo de vida de la orden. Valores posibles: DAY, IOC, FOK, GTD |
| `timeInForces` | Listado de TIFs soportados por un contrato |
| `timestamp` | Marca de tiempo de un registro en particular |
| `transactTime` | Fecha y hora de la transacción |
| `upperLimit` | Rango máximo de precio de un tick de precio dinámico |
| `wsclOrdId` | ID del request de la orden que se envió vía WebSocket. |

