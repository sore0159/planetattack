package overpower

import (
	"mule/hexagon"
)

func RunGameTurn(source Source) (breaker, logger error) {
	game, err := source.Game()
	if my, bad := Check(err, "run turn resource failure"); bad {
		return my, nil
	}
	planets, err := source.Planets()
	if my, bad := Check(err, "run turn resource failure"); bad {
		return my, nil
	}
	factions, err := source.Factions()
	if my, bad := Check(err, "run turn resource failure"); bad {
		return my, nil
	}
	orders, err := source.Orders()
	if my, bad := Check(err, "run turn resource failure"); bad {
		return my, nil
	}
	ships, err := source.Ships()
	if my, bad := Check(err, "run turn resource failure"); bad {
		return my, nil
	}
	truces, err := source.Truces()
	if my, bad := Check(err, "run turn resource failure"); bad {
		return my, nil
	}
	dbPowerOrders, err := source.PowerOrders()
	if my, bad := Check(err, "run turn resource failure"); bad {
		return my, nil
	}
	// --------- GAME ALREADY OVER -------- //
	if game.HighScore() >= game.ToWin() {
		return nil, nil
	}
	// -------------------------------- //
	var errOccured bool
	loggerM, _ := Check(ErrIgnorable, "run turn problem")
	planetGrid := make(map[hexagon.Coord]Planet, len(planets))
	radar := make(map[int]hexagon.CoordList, len(factions))

	truceMap := map[hexagon.Coord]map[[2]int]bool{}
	for _, tr := range truces {
		loc := tr.Loc()
		if mp, ok := truceMap[loc]; ok {
			mp[[2]int{tr.Fid(), tr.Trucee()}] = true
		} else {
			truceMap[loc] = map[[2]int]bool{[2]int{tr.Fid(), tr.Trucee()}: true}
		}
	}
	var atWar []Planet
	for _, p := range planets {
		loc := p.Loc()
		planetGrid[loc] = p
		pFid, sFid := p.PrimaryFaction(), p.SecondaryFaction()
		if pFid != 0 && sFid != 0 {
			if mp := truceMap[loc]; mp == nil || !mp[[2]int{pFid, sFid}] || !mp[[2]int{sFid, pFid}] {
				atWar = append(atWar, p)
			}
		}
		for _, fid := range []int{pFid, sFid} {
			if fid == 0 {
				continue
			}
			if list, ok := radar[fid]; ok {
				radar[fid] = append(list, loc)
			} else {
				radar[fid] = hexagon.CoordList{loc}
			}
		}
	}
	powerOrders := make([]PowerOrder, 0, len(dbPowerOrders))
	for _, pO := range dbPowerOrders {
		pl, ok := planetGrid[pO.Loc()]
		if !ok {
			errOccured = true
			loggerM.AddContext("bad powerorder", "planet not found", "powerorder", pO)
			continue
		}
		fid := pO.Fid()
		if pl.PrimaryFaction() != fid && pl.SecondaryFaction() != fid {
			errOccured = true
			loggerM.AddContext("bad powerorder", "planet not owner", "powerorder", pO)
			continue
		}
		powerOrders = append(powerOrders, pO)
	}
	turn := game.Turn()
	var auto bool
	for _, f := range factions {
		doneB := f.DoneBuffer()
		if doneB > 0 {
			f.SetDoneBuffer(doneB - 1)
		} else if doneB == 0 {
			auto = true
		}
	}
	// ------ AUTO TURN ------- //
	if !auto {
		game.SetFreeAutos(game.FreeAutos() + 1)
	}
	// ---- PLANETS AT WAR ---- //
	for _, p := range atWar {
		Battle(source, p, nil, turn, truceMap[p.Loc()])
	}
	// ---- SHIPS LAUNCH ---- //
	var secondaryOrders []Order

	for _, o := range orders {
		src, ok2 := planetGrid[o.Source()]
		tar, ok1 := planetGrid[o.Target()]
		if !(ok1 && ok2) {
			errOccured = true
			loggerM.AddContext("bad order", "planets not found", "order", o)
			continue
		}
		size := o.Size()
		if size < 1 {
			errOccured = true
			loggerM.AddContext("bad order", "size <0", "order", o)
			continue
		}
		if fid := o.Fid(); fid == src.SecondaryFaction() {
			secondaryOrders = append(secondaryOrders, o)
			continue
		} else if fid != src.PrimaryFaction() {
			errOccured = true
			loggerM.AddContext("bad order", "bad controller", "order", o)
			continue
		}
		switch src.PrimaryPower() {
		case -1:
			if have := src.Tachyons(); size > have {
				size = have
				src.SetTachyons(0)
			} else {
				src.SetTachyons(have - size)
			}
		case 1:
			if have := src.Antimatter(); size > have {
				size = have
				src.SetAntimatter(0)
			} else {
				src.SetAntimatter(have - size)
			}
		default:
			errOccured = true
			loggerM.AddContext("bad order", "bad power", "order", o)
			continue
		}
		if size > 0 {
			path := src.Loc().PathTo(tar.Loc())
			sh := source.NewShip(src.PrimaryFaction(), size, turn, path)
			ships = append(ships, sh)
			source.NewLaunchRecord(turn, o, sh)
		} else {
			source.NewLaunchRecord(turn, o, nil)
		}
	}

	for _, o := range secondaryOrders {
		src, ok2 := planetGrid[o.Source()]
		tar, ok1 := planetGrid[o.Target()]
		if !(ok1 && ok2) {
			errOccured = true
			loggerM.AddContext("bad order", "planets not found", "order", o)
			continue
		}
		size := o.Size()
		if size < 1 {
			errOccured = true
			loggerM.AddContext("bad order", "size <0", "order", o)
			continue
		}
		switch src.SecondaryPower() {
		case -1:
			if have := src.Tachyons(); size > have {
				size = have
				src.SetTachyons(0)
			} else {
				src.SetTachyons(have - size)
			}
		case 1:
			if have := src.Antimatter(); size > have {
				size = have
				src.SetAntimatter(0)
			} else {
				src.SetAntimatter(have - size)
			}
		default:
			errOccured = true
			loggerM.AddContext("bad order", "bad power", "order", o)
			continue
		}
		if size > 0 {
			path := src.Loc().PathTo(tar.Loc())
			sh := source.NewShip(src.SecondaryFaction(), size, turn, path)
			ships = append(ships, sh)
			source.NewLaunchRecord(turn, o, sh)
		} else {
			source.NewLaunchRecord(turn, o, nil)
		}
	}
	source.DropOrders()
	// ---- SHIPS MOVE ---- //
	// dist, ship index
	landings := map[int][]int{}
	for i, sh := range ships {
		travelled, land := Travelled(sh, turn)
		if len(travelled) < 1 {
			errOccured = true
			loggerM.AddContext("bad ship", "no travel dist", "ship", sh)
			source.DropShip(sh)
			continue
		}
		at := travelled[len(travelled)-1]
		// ----- SHIP MOVEMENT IS SEEN ------ //
		for fid, rList := range radar {
			var destValid, spottedShip bool
			var spotted hexagon.CoordList
			if fid == sh.Fid() {
				spotted, spottedShip = travelled, true
				destValid = true
			} else {
				spotted, spottedShip = RadarCheck(rList, travelled)
			}
			if len(spotted) > 0 {
				var trail []hexagon.Coord
				var loc, dest hexagon.Coord
				locValid := spottedShip && !land
				if locValid {
					loc = at
					trail = spotted[:len(spotted)-1]
				} else {
					trail = spotted
				}
				if destValid {
					path := sh.Path()
					dest = path[len(path)-1]
				}
				var locNC, destNC hexagon.NullCoord
				locNC.Valid = locValid
				locNC.Coord = loc
				destNC.Valid = destValid
				destNC.Coord = dest
				source.NewShipView(sh, fid, turn, locNC, destNC, trail)
			}
		}
		// ---- LANDINGS TAGGED FOR LATER ------ //
		if land {
			dist := len(travelled) - 1
			if list, ok := landings[dist]; ok {
				landings[dist] = append(list, i)
			} else {
				landings[dist] = []int{i}
			}
		}
	}
	//
	// ---- SHIPS LAND ---- //
	// plid, amount
	for i := 1; i < SHIPSPEED+1; i++ {
		shipsLandings, ok := landings[i]
		if !ok {
			continue
		}
		for _, sI := range shuffleInts(shipsLandings) {
			sh := ships[sI]
			path := sh.Path()
			loc := path[len(path)-1]
			p, ok := planetGrid[loc]
			if !ok {
				loggerM.AddContext("bad ship", "landing nonexistant", "ship", sh)
				errOccured = true
			} else {
				Battle(source, p, sh, turn, truceMap[loc])
			}
			source.DropShip(sh)
		}
		delete(landings, i)
	}
	if len(landings) > 0 {
		loggerM.AddContext("bad ships", "landings with improper dist", "landings", landings)
		errOccured = true
	}
	//
	// ------- PLANETS CHANGE POWER -------- //
	for _, pO := range powerOrders {
		pl := planetGrid[pO.Loc()]
		fid := pO.Fid()
		var powNum int
		if upP := pO.UpPower(); upP > 0 {
			powNum = 1
		} else if upP < 0 {
			powNum = -1
		} else {
			continue
		}
		if pl.PrimaryFaction() == fid {
			pl.SetPrimaryPower(powNum)
		} else if pl.SecondaryFaction() == fid {
			pl.SetSecondaryPower(powNum)
		} else {
			continue
		}
	}
	source.ClearPowerOrders()
	// ---- TURN STARTS ---- //
	game.IncTurn()
	turn = game.Turn()
	facScores := make(map[int]int, len(factions))
	for _, pl := range planets {
		// ---- PLANETS ARE SEEN ---- //
		for _, cont := range []int{pl.PrimaryFaction(), pl.SecondaryFaction()} {
			if cont == 0 {
				continue
			}
			facScores[cont] += 1
			source.UpdatePlanetView(cont, turn, pl)
		}
	}
	var highScore int
	toWin := game.ToWin()
	winners := make([]Faction, 0)
	for _, f := range factions {
		score := facScores[f.Fid()]
		if score > highScore {
			highScore = score
		}
		f.SetScore(score)
		if score >= toWin {
			winners = append(winners, f)
		}
	}
	game.SetHighScore(highScore)
	if len(winners) > 0 {
		Ping("TODO: WINNING!", winners)
	}
	if errOccured {
		return nil, loggerM
	} else {
		return nil, nil
	}
}
