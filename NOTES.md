# Тезисы

существующие модели: 
- `ICON`, `GFS`, `ECMWF`
- `Т169L31` - росгидромет??

## Аббревиатуры

|Аббревиатура|Описание|
|-|-|
|`EPSG`| `European Petroleum Survey Group` публичный реестр геодезических данных, систем пространственной привязки, земных эллипсоидов, преобразований координат и связанных с ними единиц измерения|
|`SRID`| `Spatial Reference System Identifier` идентификатор системы пространственной привязки|
|`WGS 84`| `World Geodetic System` единая всемирная система геодезических координат, определяющая координаты относительно центра масс Земли|
|`EOSDIS`| `Earth Observing System Data and Information System` система данных и информации системы наблюдения Земли|

## Модель GFS

- Вычисляется 4 раза в день (интервал 6 часов)
- 7 дней (лучше 3-4) дальше не достаточно надежно (есть сведения, что для Североамериканского и Сибирского регионов наиболее точен)
- Европейская `ECMWF` лучше американской (не ясно какого года статья так как после обновления модели в 2017 - 2019гг стало сильно лучше)
- `ICON` - хорош по осадкам


#### Ссылки
 - https://vlab.noaa.gov/web/gfs/documentation
 - https://method.meteorf.ru/publ/tr/tr359/tolstih.pdf
 - https://meteolabs.org/article/что-такое-модели-прогноза-погоды/
---
 - https://www.ready.noaa.gov/index.php - Real-time Environmental Applications and Display sYstem, запрос и визуализация данных о погоде
---
- https://www.ecmwf.int - сайт `ECMWF`
- https://confluence.ecmwf.int/display/FUG/1+Introduction - Forecast user guide от ECMWF
- https://github.com/ecmwf/eccodes - библиотека (`C`/`Fortran`/`Python`) ecCodes от ECMWF для работы с файлами `GRIB` и `BUFR`
- https://www.openskiron.org/en/icon-gribs - GRIB файлы для модели ICON
- https://apps.ecmwf.int/datasets/data/tigge/levtype=sfc/type=cf/ - страница запроса GRIB файлов от ECMWF (требуется аккаунт)
---
 - https://meteoinfo.ru/categ-articles/11-actuals-cat/1281-1246618396grib - (GRIB1) данные от Росгидромета?
 - https://meteoinfo.ru/images/media/books-docs/WMO/wmo-N306_vI2_codes.pdf - описание формата GRIB2 на русском языке
 - https://method.meteorf.ru/publ/tr/tr346/rosin.pdf - описание принципа работы модели `Т169L31`
---
 - https://www.meteorf.gov.ru/about/structure/ - структура Росгидромета
---
 - https://habr.com/ru/articles/235283/ - Ликбез по картографическим проекциям
 - https://habr.com/ru/articles/239251/ - Google WEB Mercator `EPSG:3857`
 - https://www.earthdata.nasa.gov/eosdis - EOSDIS
 - https://overpass-turbo.eu/ - Overpass Turbo инструмент для запроса данных из OSM
