package main

import (
	"strconv"
	"strings"
)


func userHasSkill(skills []Skill, categorie string) bool {
	target := strings.TrimSpace(categorie)
	for _, s := range skills {
		if strings.EqualFold(strings.TrimSpace(s.Nom), target) {
			return true
		}
	}
	return false
}


func isSelfExchange(ownerID int, requesterID string) bool {
	return strconv.Itoa(ownerID) == requesterID
}


func isValidNote(note int) bool {
	return note >= 1 && note <= 5
}


func isActiveExchangeStatus(status string) bool {
	return status == "pending" || status == "accepted"
}
