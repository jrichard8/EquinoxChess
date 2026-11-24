package search

const quietDefenseMargin = 80

type quietDefense struct {
	move  uint32
	score int
}

func (d *quietDefense) consider(move uint32, score int) {
	if d.move == 0 || score > d.score {
		d.move, d.score = move, score
	}
}

func (d quietDefense) withinWindow(alpha int) bool {
	return d.move != 0 && d.score > alpha-quietDefenseMargin
}
