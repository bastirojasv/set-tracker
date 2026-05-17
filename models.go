package main

import "time"

type UserProfile struct {
	ID        int
	Name      string
	Color     string // hex e.g. "#00BFFF"
	PhotoPath string // can be empty
}

type SetRecord struct {
	ID            int
	PlayedAt      time.Time
	SetType       string // "FT" or "BO"
	SetNumber     int
	UserScore     int
	OpponentScore int
	OpponentName  string
	OpponentColor string
	UserWon       *bool  // nil = active set, true/false = completed
	Status        string // "active" or "completed"
}

type GameRecord struct {
	ID         int
	SetID      int
	GameNumber int
	Winner     string // "user" or "opponent"
}

// ActiveSetState holds the in-memory state of the current set
type ActiveSetState struct {
	SetID         int
	SetType       string // "FT" or "BO"
	SetNumber     int
	OpponentName  string
	OpponentColor string
	UserScore     int
	OppScore      int
	Games         []string // sequence of "user"/"opponent"
}

func (s *ActiveSetState) IsLibre() bool {
	return s.SetType == "LS"
}

func (s *ActiveSetState) WinsNeeded() int {
	if s.SetType == "FT" {
		return s.SetNumber
	}
	if s.SetType == "LS" {
		return s.SetNumber // máximo de puntos permitido
	}
	// BO: ceil(N/2)
	return (s.SetNumber + 1) / 2
}

func (s *ActiveSetState) IsFinished() bool {
	if s.IsLibre() {
		return false // LS termina manualmente con el botón Finalizar
	}
	needed := s.WinsNeeded()
	return s.UserScore >= needed || s.OppScore >= needed
}

func (s *ActiveSetState) Winner() string {
	if s.IsLibre() {
		if s.UserScore > s.OppScore {
			return "user"
		}
		if s.OppScore > s.UserScore {
			return "opponent"
		}
		return "" // empate
	}
	if s.UserScore >= s.WinsNeeded() {
		return "user"
	}
	if s.OppScore >= s.WinsNeeded() {
		return "opponent"
	}
	return ""
}
