package services

func CelsiusToFahrenheit(celsius float64) float64 {
	return (celsius * 1.8) + 32
}

func CelsiusToKelvin(celsius float64) float64 {
	return celsius + 273.15
}
