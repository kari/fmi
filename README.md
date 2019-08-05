# FMI
![Go report](https://goreportcard.com/badge/github.com/kari/fmi)

Tämä Go-kirjasto hakee Ilmatieteen laitoksen rajapintojen kautta viimeisimmät säähavainnot halutulle paikalle. Hyödyllinen esimerkiksi IRC-bottia varten.

## Käyttö

```go
import (
    "fmt"

    "github.com/kari/fmi"
)

func main() {
    fmt.Println(fmi.Weather("turku"))
    // Viimeisimmät säähavainnot paikassa Turku: lämpötila 18.5°C, puolipilvistä, heikkoa länsituulta 4 m/s (6 m/s), ilmankosteus 56%
}
```

Katso examples/ -kansiosta lisää esimerkkejä.

## Lähteet

* [Ilmatieteen laitoksen latauspalveun pikaohje](https://ilmatieteenlaitos.fi/latauspalvelun-pikaohje)
* [BCIRC/py/fmi.py](https://github.com/Jonuz/BCIRC/blob/master/py/fmi.py)

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

[MIT](https://choosealicense.com/licenses/mit/)
