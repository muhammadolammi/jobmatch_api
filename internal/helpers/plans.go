package helpers

func GetPlanDailyUsage(planName string) (bool, int32) {
	store := map[string]int32{
		"pro":  20,
		"team": 1000,
	}
	val, ok := store[planName]
	return ok, val
}
