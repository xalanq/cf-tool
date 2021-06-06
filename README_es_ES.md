# Codeforces Tool

[![Github release](https://img.shields.io/github/release/xalanq/cf-tool.svg)](https://github.com/xalanq/cf-tool/releases)
[![platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-blue.svg)](https://github.com/xalanq/cf-tool/releases)
[![Build Status](https://travis-ci.org/xalanq/cf-tool.svg?branch=master)](https://travis-ci.org/xalanq/cf-tool)
[![Go Report Card](https://goreportcard.com/badge/github.com/xalanq/cf-tool)](https://goreportcard.com/report/github.com/xalanq/cf-tool)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.12-green.svg)](https://github.com/golang)
[![license](https://img.shields.io/badge/license-MIT-%23373737.svg)](https://raw.githubusercontent.com/xalanq/cf-tool/master/LICENSE)

Codeforces Tool es una herramienta de interfaz de línea de comandos para [Codeforces](https://codeforces.com).

Es rápido, pequeño, multiplataforma y potente.

[Instalación](#instalación) | [Uso](#uso) | [FAQ](#faq) | [中文](./README_zh_CN.md) | [Español](./README_es_ES.md)

## Características

* Soporte de Competencias, Gym de Codeforces, Grupos y acmsguru.
* Admite todos los lenguajes de programación en Codeforces.
* Envíe sus soluciones.
* Observe el estado de los envíos de forma dinámica.
* Observa los problemas.
* Compile y pruebe localmente.
* Clona todos los códigos de alguien.
* Genere códigos a partir de la plantilla especificada (incluida la marca de tiempo, el autor, etc.)
* Enumere las estadísticas de problemas de una competencia.
* Utilice el navegador web predeterminado para abrir las páginas de problemas, la página de clasificación, etc.
* Configure un proxy de red. Configure un host espejo.
* CLI colorido.

Las solicitudes de extracción siempre son bienvenidas.

![](./assets/readme_1.gif)

## Instalación

Puede descargar el archivo binario precompilado [aquí](https://github.com/xalanq/cf-tool/releases).

Entonces disfruta de cf-tool~

O puede compilarlo desde la fuente **(go >= 1.12)**:

```plain
$ go get github.com/xalanq/cf-tool
$ cd $GOPATH/src/github.com/xalanq/cf-tool
$ go build -ldflags "-s -w" cf.go
```

Si no sabe cuál es el `$GOPATH`, consulte aquí <https://github.com/golang/go/wiki/GOPATH>.

## Usage

Simulemos una competencia.

 `cf race 1136` o `cf race https://codeforces.com/contest/1136`


¡Para empezar a competir en el concurso 1136!

Si el concurso aún no ha comenzado, `cf` contará hacia atrás. Si el concurso ha comenzado o la cuenta regresiva termina, `cf` usará el navegador predeterminado para abrir la página del tablero y la página de problemas, y buscará todas las muestras en el local.

 `cd ./cf/contest/1136/a` (Puede ser diferente a esto, observe el mensaje en su pantalla)

Ingrese al directorio del problema A, el directorio debe contener todas las muestras del problema.

 `cf gen` 

Genere un código con la plantilla predeterminada. El nombre de archivo del código es ID de problema por defecto.

 `vim a.cpp` 

Usa Vim para escribir el código (depende de ti mismo).

 `cf test` 

Compile y pruebe todos los casos de prueba.

 `cf submit` 

Envíe el código.

 `cf list` 

Enumere las estadísticas de problemas del concurso.

 `cf stand` 

Abre la página de clasificación del concurso.

```plain
Debe ejecutar "cf config" para configurar su identificador, contraseña y código
plantillas al principio.

Si quieres competir, el mejor comando es "cf race".

Usage:
  cf config
  cf submit [-f <file>] [<specifier>...]
  cf list [<specifier>...]
  cf parse [<specifier>...]
  cf gen [<alias>]
  cf test [<file>]
  cf watch [all] [<specifier>...]
  cf open [<specifier>...]
  cf stand [<specifier>...]
  cf sid [<specifier>...]
  cf race [<specifier>...]
  cf pull [ac] [<specifier>...]
  cf clone [ac] [<handle>]
  cf upgrade

Options:
  -h --help            Muestra esta pantalla.
  --version            Mostrar versión.
  -f <file>, --file <file>, <file>
                       Ruta al archivo. Ej: "a.cpp", "./temp/a.cpp"
  <specifier>          Cualquier texto útil. Ej:
                       "https://codeforces.com/contest/100",
                       "https://codeforces.com/contest/180/problem/A",
                       "https://codeforces.com/group/Cw4JRyRGXR/contest/269760",
                       "1111A", "1111", "a", "Cw4JRyRGXR"
                       Puede combinar varios especificadores para
                       especificar lo que desea.
  <alias>              Alias de la plantilla. Ej: "cpp"
  ac                   El estado del envío es Aceptado.

Examples:
  cf config            Configure la herramienta cf-tool.
  cf submit            cf detectará lo que desea enviar automáticamente.
  cf submit -f a.cpp
  cf submit https://codeforces.com/contest/100/A
  cf submit -f a.cpp 100A 
  cf submit -f a.cpp 100 a
  cf submit contest 100 a
  cf submit gym 100001 a
  cf list              Enumere las estadísticas de todos los problemas de un concurso.
  cf list 1119
  cf parse 100         Obtenga todos los casos de prueba de los problemas del concurso 100 en
                       "{cf}/{contest}/100/".
  cf parse gym 100001a
                       Obtener muestras del problema "a" del Gym 100001 en
                       "{cf}/{gym}/100001/a".
  cf parse gym 100001
                       Obtenga todas las muestras de los problemas del Gym 100001 en
                       "{cf}/{gym}/100001".
  cf parse             Obtenga muestras del problema actual en la ruta actual.
  cf gen               Genere un código a partir de la plantilla predeterminada.
  cf gen cpp           Genere un código de la plantilla cuyo alias es "cpp"
                       en la ruta actual.
  cf test              Ejecute los comandos de una plantilla en la ruta actual.
                       Luego pruebe todas las muestras. Si desea agregar un nuevo
                       caso de prueba, cree dos archivos "inK.txt" y "ansK.txt"
                       donde K es una cadena con 0 ~ 9.
  cf watch             Mira las primeras 10 envios del concurso actual.
  cf watch all         Mira todas los envios del concurso actual.
  cf open 1136a        Utilice el navegador web predeterminado para abrir
                       la página del concurso 1136, problema a.
  cf open gym 100136   Utilice el navegador web predeterminado para abrir
                       la página del Gym 100136.
  cf stand             Utilice el navegador web predeterminado para abrir
                       la página de posiciones de la competencia.
  cf sid 52531875      Utilice el navegador web predeterminado para abrir
                       la página del envío 52531875.
  cf sid               Abra la página del último envío.
  cf race 1136         Si el concurso 1136 aún no ha comenzado, comenzará
                       la cuenta regresiva. Cuando finalice la cuenta regresiva,
                       se abrirán las páginas de todos los problemas y analizará
                       las muestras.
  cf pull 100          Extraiga los últimos códigos del concurso 100 de todos
                       los problemas en "./100/<problem-id>".
  cf pull 100 a        Extraiga el último código del problema "a" del concurso 100
                       en "./100/<problem-id>".
  cf pull ac 100 a     Extraiga el código del problema "a" del concurso 100 "Aceptado"
                       o "Pruebas previas aprobadas".
  cf pull              Extraiga los últimos códigos del problema actual
                       en la ruta actual.
  cf clone xalanq      Clona todos los códigos de xalanq.
  cf upgrade           Actualice "cf" a la última versión de GitHub.

File:
  cf guardará algunos datos en algunos archivos:

  "~/.cf/config"        Archivo de configuración, incluidas plantillas, etc.
  "~/.cf/session"       Archivo de sesión, incluidas cookies, identificador,
                        contraseña, etc.

  "~" es el directorio de inicio del usuario actual en su sistema.

Plantilla:
  Puede insertar algunos marcadores de posición en el código de su plantilla.
  Cuando se genera un código a partir de la plantilla, cf reemplazará todos
  los marcadores de posición siguiendo las siguientes reglas:

  $%U%$   Handle (e.g. xalanq)
  $%Y%$   Year   (e.g. 2019)
  $%M%$   Month  (e.g. 04)
  $%D%$   Day    (e.g. 09)
  $%h%$   Hour   (e.g. 08)
  $%m%$   Minute (e.g. 05)
  $%s%$   Second (e.g. 00)

Script en plantilla:
  La plantilla ejecutará 3 scripts en secuencia cuando ejecute "cf test":
    - before_script   (ejecutar una vez)
    - script          (ejecutar el número de casos de prueba)
    - after_script    (ejecutar una vez)
  Puede establecer "before_script" o "after_script" en una cadena vacía,
  lo que significa que no se ejecuta.
  Tienes que ejecutar tu programa en "script" con el estándar de
  entrada/salida (no es necesario redireccionar).

  Puede insertar algunos marcadores de posición en sus scripts. Cuando se ejecuta un script, cf reemplazará todos los marcadores de posición por las siguientes reglas:

  $%path%$   Ruta al archivo de origen (Excluyendo $%full%$, Ej: "/home/xalanq/")
  $%full%$   Nombre completo del archivo de origen (Ej: "a.cpp")
  $%file%$   Nombre del archivo de origen (Excluyendo el sufijo, Ej: "a")
  $%rand%$   Cadena aleatoria con 8 caracteres (Incluyendo "a-z" "0-9")
```

## Template Example

Los marcadores de posición dentro de la plantilla se reemplazarán con el contenido correspondiente cuando ejecute `cf gen`.

```
$%U%$   Handle (e.g. xalanq)
$%Y%$   Year   (e.g. 2019)
$%M%$   Month  (e.g. 04)
$%D%$   Day    (e.g. 09)
$%h%$   Hour   (e.g. 08)
$%m%$   Minute (e.g. 05)
$%s%$   Second (e.g. 00)
```

```cpp
/* Generated by powerful Codeforces Tool
 * You can download the binary file in here https://github.com/xalanq/cf-tool (Windows, macOS, Linux)
 * Author: $%U%$
 * Time: $%Y%$-$%M%$-$%D%$ $%h%$:$%m%$:$%s%$
**/

#include <bits/stdc++.h>
using namespace std;

typedef long long ll;

int main() {
    ios::sync_with_stdio(false);
    cin.tie(0);
    
    return 0;
}
```

## Preguntas más frecuentes

### Hago doble clic en el programa pero no funciona

Codeforces Tool es una herramienta de línea de comandos. Deberías ejecutarlo en la terminal.

### No puedo usar el comando `cf`

Debe poner el programa `cf` en una ruta (por ejemplo,`/usr/bin/` en Linux) que se ha agregado a la variable de entorno del sistema PATH.

O simplemente google "cómo agregar una ruta a la variable de entorno del sistema PATH".

### ¿Cómo agregar un nuevo caso de prueba?

Cree dos archivos de casos de prueba adicionales `inK.txt` y` ansK.txt` (K es una cadena con 0 ~ 9).

### Habilitar la finalización de pestañas en la terminal

Utilizar este [Infinidat/infi.docopt_completion](https://github.com/Infinidat/infi.docopt_completion).

Nota: Si hay una nueva versión lanzada (especialmente un nuevo comando agregado), debe ejecutar `docopt-completion cf` de nuevo.
