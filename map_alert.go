package main

func mapAlert(alert alert) map[string]interface{} {
	a := map[string]interface{}{
		"status":       alert.Status,
		"startsAt":     alert.StartsAt,
		"endsAt":       alert.EndsAt,
		"generatorURL": alert.GeneratorURL,
	}

	for k, v := range alert.Labels.(map[string]interface{}) {
		a["label_"+k] = v
	}

	for k, v := range alert.Annotations.(map[string]interface{}) {
		a["annotation_"+k] = v
	}

	return a
}
