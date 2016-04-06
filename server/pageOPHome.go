package main

import (
	"mule/overpower"
	"net/http"
)

var (
	TPOPHOME = MixTemp("frame", "titlebar", "ophome")
)

func pageOPHome(w http.ResponseWriter, r *http.Request) {
	h := MakeHandler(w, r)
	if !h.LoggedIn {
		http.Redirect(w, r, "/auth/login", http.StatusFound)
		return
	}
	if r.URL.Path != "/overpower/home" {
		http.Redirect(w, r, "/overpower/home", http.StatusFound)
		return
	}
	var hasG bool
	var g overpower.GameDat

	games, err := h.M.Game().Select("owner", h.User.String())
	if my, bad := Check(err, "resource failure on OP home page", "resource", "game", "owner", h.User.String()); bad {
		Bail(w, my)
		return
	}
	if len(games) != 0 {
		g = games[0]
		hasG = true
	}

	var gFacs []overpower.FactionDat
	var gHasF bool
	if hasG {
		gFacs, err = h.M.Faction().SelectWhere(h.GID(g.GID()))
		if my, bad := Check(err, "resource error in homepage", "resource", "faction", "user", h.User, "gid", g.GID()); bad {
			Bail(w, my)
			return
		}
		gHasF = len(gFacs) > 0
	}
	if r.Method == "POST" {
		/*
			if DBLOCK {
				http.Error(w, "GAME DOWN FOR DAYLY MAINT: 10-20MIN", http.StatusInternalServerError)
				return
			}
			action := r.FormValue("action")
			switch action {
			case "nextturn", "setautos":
				if !hasG {
					http.Error(w, "USER HAS NO GAME TO PROGRESS", http.StatusBadRequest)
					return
				}
				ints, ok := GetInts(r, "turn")
				if !ok {
					http.Error(w, "MALFORMED TURN DATA", http.StatusBadRequest)
					return
				}
				turn := ints[0]
				if turn != g.Turn() {
					http.Error(w, "BAD TURN DATA", http.StatusBadRequest)
					return
				}
				if action == "nextturn" {
					if g.Turn() < 1 {
						http.Error(w, "GAME NOT YET BEGUN", http.StatusBadRequest)
						return
					}
					err = RunGameTurn(g.Gid())
					if my, bad := Check(err, "failure on running turn", "gid", g.Gid()); bad {
						Bail(w, my)
						return
					}
				} else {
					dayBool := [7]bool{}
					dayBool[0] = r.FormValue("sunday") == "on"
					dayBool[1] = r.FormValue("monday") == "on"
					dayBool[2] = r.FormValue("tuesday") == "on"
					dayBool[3] = r.FormValue("wednesday") == "on"
					dayBool[4] = r.FormValue("thursday") == "on"
					dayBool[5] = r.FormValue("friday") == "on"
					dayBool[6] = r.FormValue("saturday") == "on"
					g.SetAutoDays(dayBool)
					err = OPDB.UpdateGames(g)
					if my, bad := Check(err, "failure on updating game", "game", g); bad {
						Bail(w, my)
						return
					}
				}
			case "startgame":
				if !hasG {
					h.SetError("USER HAS NO GAME TO START")
					break
				}
				if g.Turn() > 0 {
					h.SetError("GAME ALREADY IN PROGRESS")
					break
				}
				if len(gFacs) < 1 {
					h.SetError("GAME HAS NO PLAYERS")
					break
				}
				exodus := r.FormValue("exodus") == "on"
				err := StartGame(g.Gid(), exodus)
				if my, bad := Check(err, "startgame failure", "gid", g.Gid()); bad {
					Log(my)
					h.SetError("DATABASE ERROR STARTING GAME")
					break
				}
			case "newgame":
				if hasG {
					h.SetError("USER ALREADY HAS GAME IN PROGRESS")
					break
				}
				name, password := r.FormValue("gamename"), r.FormValue("password")
				if !ValidText(name) {
					h.SetError("INVALID GAME NAME")
					break
				}
				if password != "" && !ValidText(password) {
					h.SetError("INVALID GAME PASSWORD")
					break
				}
				facName := r.FormValue("facname")
				if facName != "" && !ValidText(facName) {
					h.SetError("INVALID FACTION NAME")
					break
				}
				ints, ok := GetInts(r, "towin")
				if !ok {
					h.SetError("INVALID GAME WINPERCENT")
					break
				}
				towin := ints[0]
				if towin < 2 {
					h.SetError("INVALID GAME WINPERCENT")
					break
				}
				err = OPDB.MakeGame(h.User.String(), name, password, towin)
				if my, bad := Check(err, "make game failure", "user", h.User, "name", name, "password", password, "towin", towin); bad {
					Log(my)
					h.SetError("DATABASE ERROR IN GAME CREATION")
					break
				}
				if facName != "" {
					g, err := OPDB.GetGame("owner", h.User.String())
					if my, bad := Check(err, "make game retrieval failure", "user", h.User, "game"); bad {
						Log(my)
						h.SetError("DATABASE ERROR IN FACTION CREATION")
						break

					}
					err = OPDB.MakeFaction(g.Gid(), h.User.String(), facName)
					if my, bad := Check(err, "make faction failure", "user", h.User, "facname", facName, "gid", g.Gid()); bad {
						Log(my)
						h.SetError("DATABASE ERROR IN FACTION CREATION")
						break
					}
				}
			case "dropgame":
				if !hasG {
					h.SetError("USER HAS NO GAME IN PROGRESS")
					break
				}
				err = OPDB.DropGames("gid", g.Gid())
				if my, bad := Check(err, "drop game failure", "gid", g.Gid()); bad {
					Log(my)
					h.SetError("DATABASE ERROR IN GAME DESTRUCTION")
					break
				}
			default:
				h.SetError("UNKNOWN ACTION TYPE")
			}
			if h.Error == "" {
				http.Redirect(w, r, r.URL.Path, http.StatusFound)
				return
			}
		*/
	}
	//
	m := h.DefaultApp()
	m["user"] = h.User.String()
	if hasG {
		m["game"] = g
		m["active"] = g.Turn() > 0
	}
	if gHasF {
		m["gfactions"] = gFacs
		days := g.AutoDays()
		var any bool
		for _, b := range days {
			if b {
				any = true
				break
			}
		}
		if !any {
			m["noauto"] = true
		}
	}
	oFacs, err := h.M.Faction().Select("owner", h.User.String())
	if my, bad := Check(err, "resource error in homepage", "resource", "faction", "owner", h.User); bad {
		Log(my)
		h.SetError("DATABASE ERROR")
	}
	oHasF := len(oFacs) > 0
	if oHasF {
		facGames := make([]overpower.GameDat, 0, len(oFacs))
		for _, f := range oFacs {
			games, err := h.M.Game().SelectWhere(h.GID(f.GID()))
			if my, bad := Check(err, "resource error in homepage", "gid", f.GID(), "fac", f, "owner", h.User); bad {
				Log(my)
				h.SetError("DATABASE ERROR")
			} else {
				facGames = append(facGames, games...)
			}
		}
		m["ofactions"] = oFacs
		m["ogames"] = facGames
	}
	h.Apply(TPOPHOME, w)
}
