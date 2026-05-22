package util

import "testing"

func TestHumanDuration_Zero(t *testing.T) {
	if got := HumanDuration(0); got != "0秒" {
		t.Errorf("HumanDuration(0) = %q, want \"0秒\"", got)
	}
}

func TestHumanDuration_SecondsOnly(t *testing.T) {
	if got := HumanDuration(45); got != "45秒" {
		t.Errorf("HumanDuration(45) = %q, want \"45秒\"", got)
	}
}

func TestHumanDuration_Minutes(t *testing.T) {
	if got := HumanDuration(120); got != "2分" {
		t.Errorf("HumanDuration(120) = %q, want \"2分\"", got)
	}
}

func TestHumanDuration_Hours(t *testing.T) {
	if got := HumanDuration(3600); got != "1小时" {
		t.Errorf("HumanDuration(3600) = %q, want \"1小时\"", got)
	}
}

func TestHumanDuration_Full(t *testing.T) {
	if got := HumanDuration(3661); got != "1小时1分1秒" {
		t.Errorf("HumanDuration(3661) = %q, want \"1小时1分1秒\"", got)
	}
}

func TestHumanDuration_HoursOnly(t *testing.T) {
	if got := HumanDuration(7200); got != "2小时" {
		t.Errorf("HumanDuration(7200) = %q, want \"2小时\"", got)
	}
}

func TestHumanDuration_MinutesOnly(t *testing.T) {
	if got := HumanDuration(61); got != "1分1秒" {
		t.Errorf("HumanDuration(61) = %q, want \"1分1秒\"", got)
	}
}