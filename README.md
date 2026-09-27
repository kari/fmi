# FMI

[![Go](https://github.com/kari/fmi/actions/workflows/go.yml/badge.svg)](https://github.com/kari/fmi/actions/workflows/go.yml)

Tämä Go-kirjasto hakee Ilmatieteen laitoksen rajapintojen kautta viimeisimmät säähavainnot halutulle paikalle. Hyödyllinen esimerkiksi IRC-bottia varten.

## Käyttö

```go
import (
    "errors"
    "fmt"
    "log"

    "github.com/kari/fmi"
)

func main() {
  weather, err := fmi.Weather("Turku")
  if errors.Is(err, fmi.ErrUnknownPlace) {
    log.Fatal("tuntematon paikka")
  } else if err != nil {
    log.Fatal(err)
  }
  fmt.Println(weather)
    // Viimeisimmät säähavainnot paikassa Turku: lämpötila 18.5°C, puolipilvistä, heikkoa länsituulta 4 m/s (6 m/s), ilmankosteus 56%
}
```

Kirjasto palauttaa virheet `fmi.ErrNoPlace`, `fmi.ErrFetchFailed`, `fmi.ErrUnknownPlace` ja `fmi.ErrNoObservations`, jotka voi tunnistaa `errors.Is`:llä.

## Rakenteinen tulos

`Current` palauttaa havainnot rakenteisena, jolloin arvot saa suoraan ilman tekstin jäsentämistä:

```go
w, err := fmi.Current(context.Background(), "Turku")
if err != nil {
    log.Fatal(err)
}
fmt.Println(w)                           // sama kuvaus kuin Weatherilla
fmt.Println(w.Observations.Temperature)  // lämpötila numerona
```

`Weather`-rakenteen `String()` tuottaa saman kuvauksen kuin `fmi.Weather`, ja `NaN`-arvo tarkoittaa, että mita puuttuu.

Katso examples/ -kansiosta lisää esimerkkejä.

Huom. FMI:n rajapinta tunnistaa jotkin paikat vain ruotsinkielisellä nimellä (esim. Tammisaari löytyy vain nimellä Ekenäs, ks. [issue #2](https://github.com/kari/fmi/issues/2)).

## Komentorivityökalu

Kirjaston mukana tulee yksinkertainen komentorivityökalu:

```sh
go install github.com/kari/fmi/cmd/saa@latest
saa Turku
```

## Lähteet

* [Ilmatieteen laitoksen latauspalvelun pikaohje](https://ilmatieteenlaitos.fi/latauspalvelun-pikaohje)
* [BCIRC/py/fmi.py](https://github.com/Jonuz/BCIRC/blob/master/py/fmi.py)

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate, and run `just check` (tests and linter) before submitting.

## License

[MIT](https://choosealicense.com/licenses/mit/)

Open data from FMI used under [Creative Commons Attribution 4.0 International license](https://en.ilmatieteenlaitos.fi/open-data-licence)
