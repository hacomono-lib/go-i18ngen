package i18n

import (
	"testing"
)

func TestEntityTextsGodoc(t *testing.T) {
	// Test that EntityTexts utility variables work correctly
	userText := EntityTexts.User
	if userText.Localize("en") != "User" {
		t.Errorf("Expected 'User', got %s", userText.Localize("en"))
	}
	if userText.Localize("ja") != "ユーザー" {
		t.Errorf("Expected 'ユーザー', got %s", userText.Localize("ja"))
	}

	productText := EntityTexts.Product
	if productText.Localize("en") != "Product" {
		t.Errorf("Expected 'Product', got %s", productText.Localize("en"))
	}
	if productText.Localize("ja") != "製品" {
		t.Errorf("Expected '製品', got %s", productText.Localize("ja"))
	}
}

func TestReasonTextsGodoc(t *testing.T) {
	// Test that ReasonTexts utility variables work correctly
	reasonText := ReasonTexts.AlreadyDeleted
	if reasonText.Localize("en") != "already deleted" {
		t.Errorf("Expected 'already deleted', got %s", reasonText.Localize("en"))
	}
	if reasonText.Localize("ja") != "すでに削除されています" {
		t.Errorf("Expected 'すでに削除されています', got %s", reasonText.Localize("ja"))
	}
}

func TestCompleteMessageWithGodocUtilities(t *testing.T) {
	// Test using utility variables to create complete messages
	msg := NewEntityNotFound(EntityTexts.User, ReasonTexts.AlreadyDeleted)

	enResult := msg.Localize("en")
	expected := "User not found: already deleted"
	if enResult != expected {
		t.Errorf("Expected '%s', got '%s'", expected, enResult)
	}

	jaResult := msg.Localize("ja")
	expected = "ユーザーが見つかりません: すでに削除されています"
	if jaResult != expected {
		t.Errorf("Expected '%s', got '%s'", expected, jaResult)
	}
}
