package prompt

import (
	"fmt"

	"github.com/auto-brain/garagebrain/internal/model"
)

func BuildPassportPrompt(car model.Car, records []model.ServiceRecord) string {
	prompt := fmt.Sprintf(`Составь краткую сводку по автомобилю %s %s.`, car.Brand, car.Model)

	if car.Year != nil {
		prompt += fmt.Sprintf(" Год выпуска: %d.", *car.Year)
	}
	prompt += fmt.Sprintf(" Текущий пробег: %d км.", car.Mileage)

	if len(records) > 0 {
		prompt += fmt.Sprintf(" Всего записей обслуживания: %d.", len(records))

		totalCost := 0.0
		for _, r := range records {
			if r.Cost != nil {
				totalCost += *r.Cost
			}
			if r.PartsCost != nil {
				totalCost += *r.PartsCost
			}
		}
		if totalCost > 0 {
			prompt += fmt.Sprintf(" Общие затраты на обслуживание: %.2f.", totalCost)
		}
	}

	prompt += `
Сформируй структурированное резюме в формате:
- Общая информация
- История обслуживания (кратко)
- Текущие затраты
- Рекомендации
`

	return prompt
}
