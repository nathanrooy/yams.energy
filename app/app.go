package app

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
)

const upperBound float64 = 10
const lowerBound float64 = 25

type FoodCandidate struct {
	Item         FoodItem
	Scale        int64
	ItemCalories int64
}

func GenerateDescription(activityCals float64) string {

	// Filter down the food items
	var candidates []FoodCandidate
	for _, item := range FoodItems {
		if item.ServingCalories < activityCals {
			if item.IsScalable {
				scaleUpper := math.Ceil(activityCals / item.ServingCalories)
				scaleLower := math.Floor(activityCals / item.ServingCalories)
				deltaUpper := (scaleUpper * item.ServingCalories) - activityCals
				deltaLower := activityCals - (scaleLower * item.ServingCalories)
				if deltaUpper <= upperBound || deltaLower <= lowerBound {
					if deltaUpper <= deltaLower {
						candidates = append(candidates,
							FoodCandidate{
								Item:         item,
								Scale:        int64(scaleUpper),
								ItemCalories: int64(item.ServingCalories * scaleUpper),
							},
						)
					} else {
						candidates = append(candidates,
							FoodCandidate{
								Item:         item,
								Scale:        int64(scaleLower),
								ItemCalories: int64(item.ServingCalories * scaleLower),
							},
						)
					}
				}
			}
		} else if activityCals-lowerBound <= item.ServingCalories && item.ServingCalories <= activityCals+upperBound {
			candidates = append(candidates, FoodCandidate{Item: item, Scale: 1, ItemCalories: int64(item.ServingCalories)})
		}
	}

	if len(candidates) > 0 {
		itemPick := candidates[rand.Intn(len(candidates))]

		// Construct the actual message
		var calorieMessage string = strconv.FormatFloat(float64(itemPick.Scale)*itemPick.Item.ServingSize, 'f', -1, 64)

		// Add any serving prefixes if they exist
		if itemPick.Item.ServingPrefix != "" {
			calorieMessage += " " + itemPick.Item.ServingPrefix
		}

		// Add plural form of the serving units
		if itemPick.Item.ServingUnits != "" {
			if itemPick.Item.ServingSize*float64(itemPick.Scale) != 1 {
				val, ok := PluralMap[itemPick.Item.ServingUnits]
				if ok {
					calorieMessage += " " + val + " of"
				} else {
					calorieMessage += " " + itemPick.Item.ServingUnits + " of"
				}
			} else {
				calorieMessage += " " + itemPick.Item.ServingUnits + " of"
			}
		}

		// Add the actual item
		if itemPick.Scale == 1 {
			if itemPick.Item.NameSingular != "" {
				calorieMessage += fmt.Sprintf(" %s", itemPick.Item.NameSingular)
			} else {
				calorieMessage += fmt.Sprintf(" %s", itemPick.Item.NamePlural)
			}
		} else if itemPick.Item.NamePlural != "" {
			calorieMessage += fmt.Sprintf(" %s", itemPick.Item.NamePlural)
		} else {
			calorieMessage += fmt.Sprintf(" %s", itemPick.Item.NameSingular)
		}

		// Add item source
		if itemPick.Item.Source != "FDA" {
			calorieMessage += fmt.Sprintf(" from %s", itemPick.Item.Source)
		}

		return calorieMessage + " - yams․energy"
	} else {
		return ""
	}
}
