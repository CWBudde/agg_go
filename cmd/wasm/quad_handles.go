package main

func handleQuadMouseDown(x, y float64, quad *[4][2]float64, selected *int) bool {
	*selected = -1
	const r2 = 100.0 // 10px radius
	for i := 0; i < 4; i++ {
		dx := x - quad[i][0]
		dy := y - quad[i][1]
		if dx*dx+dy*dy <= r2 {
			*selected = i
			return true
		}
	}
	return false
}

func handleQuadMouseMove(x, y float64, quad *[4][2]float64, selected *int) bool {
	if *selected < 0 || *selected >= 4 {
		return false
	}
	quad[*selected][0] = x
	quad[*selected][1] = y
	return true
}

func handleQuadMouseUp(selected *int) {
	*selected = -1
}
