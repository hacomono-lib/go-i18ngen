package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Test that verifies actual runtime localization using the pre-generated testdata package
func TestDirectGoI18nRuntime(t *testing.T) {
	// This test uses the pre-generated testdata package for actual runtime validation
	// Code generation is tested separately to keep tests clean and focused

	// Now test the actual runtime behavior using the generated code
	t.Run("ValidationError_Japanese", func(t *testing.T) {
		fieldInput := NewFieldValue("メール")
		fieldDisplay := NewFieldValue("メールアドレス")
		msg := NewValidationError(fieldInput, fieldDisplay)

		result := msg.Localize("ja")
		expected := "メールのメールアドレス検証エラーです"
		require.Equal(t, expected, result, "Japanese validation error should be correct")
	})

	t.Run("ValidationError_English", func(t *testing.T) {
		fieldInput := NewFieldValue("email")
		fieldDisplay := NewFieldValue("email address")
		msg := NewValidationError(fieldInput, fieldDisplay)

		result := msg.Localize("en")
		expected := "email validation error for email address"
		require.Equal(t, expected, result, "English validation error should be correct")
	})

	t.Run("EntityNotFound_WithPlaceholders", func(t *testing.T) {
		entity := EntityTexts.User
		reason := ReasonTexts.AlreadyDeleted
		msg := NewEntityNotFound(entity, reason)

		// Test Japanese
		resultJa := msg.Localize("ja")
		expectedJa := "ユーザーが見つかりません: すでに削除されています"
		require.Equal(t, expectedJa, resultJa, "Japanese entity not found should be correct")

		// Test English
		resultEn := msg.Localize("en")
		expectedEn := "User not found: already deleted"
		require.Equal(t, expectedEn, resultEn, "English entity not found should be correct")
	})

	t.Run("WelcomeMessage_SuffixNotation", func(t *testing.T) {
		nameUser := NewNameValue("田中")
		nameOwner := NewNameValue("佐藤")
		msg := NewWelcomeMessage(nameUser, nameOwner)

		// Test Japanese
		resultJa := msg.Localize("ja")
		expectedJa := "田中さん、佐藤さんのアカウントへようこそ！"
		require.Equal(t, expectedJa, resultJa, "Japanese welcome message should be correct")

		// Test English
		resultEn := msg.Localize("en")
		expectedEn := "Welcome 田中, to 佐藤's account!"
		require.Equal(t, expectedEn, resultEn, "English welcome message should be correct")
	})

	t.Run("FallbackBehavior", func(t *testing.T) {
		entity := EntityTexts.Product
		reason := ReasonTexts.AlreadyDeleted
		msg := NewEntityNotFound(entity, reason)

		// Test fallback for unsupported locale
		result := msg.Localize("fr")
		require.NotEmpty(t, result, "Should return fallback result for unsupported locale")

		// Debug: Show what the fallback actually returns
		resultJa := msg.Localize("ja")
		resultEn := msg.Localize("en")
		t.Logf("Fallback result for 'fr': %s", result)
		t.Logf("Japanese result: %s", resultJa)
		t.Logf("English result: %s", resultEn)

		// Should return some reasonable fallback (might be error message or one of the available locales)
		require.NotContains(t, result, "panic", "Should not panic")
	})

	t.Run("PluralizationSupport", func(t *testing.T) {
		// Test ItemCount pluralization (English has one/other forms)
		entity := EntityTexts.Product

		// Test singular (1 item)
		msgSingular := NewItemCount(entity).WithPluralCount(1)
		resultEn1 := msgSingular.Localize("en")
		require.Equal(t, "Product item", resultEn1, "English singular should use 'one' form")

		// Test plural (5 items)
		msgPlural := NewItemCount(entity).WithPluralCount(5)
		resultEn5 := msgPlural.Localize("en")
		require.Equal(t, "Product items (5)", resultEn5, "English plural should use 'other' form")

		// Test Japanese (no pluralization, same form for all counts)
		resultJa1 := msgSingular.Localize("ja")
		resultJa5 := msgPlural.Localize("ja")
		require.Equal(t, "製品 アイテム (1個)", resultJa1, "Japanese should show count correctly")
		require.Equal(t, "製品 アイテム (5個)", resultJa5, "Japanese should show count correctly")

		// Test UserCount (simpler case - just numbers)
		userCount1 := NewUserCount().WithPluralCount(1)
		userCount3 := NewUserCount().WithPluralCount(3)

		// English pluralization
		require.Equal(t, "1 user", userCount1.Localize("en"), "English user count singular")
		require.Equal(t, "3 users", userCount3.Localize("en"), "English user count plural")

		// Japanese (no pluralization)
		require.Equal(t, "1人のユーザー", userCount1.Localize("ja"), "Japanese user count")
		require.Equal(t, "3人のユーザー", userCount3.Localize("ja"), "Japanese user count")

		// Test edge cases
		userCount0 := NewUserCount().WithPluralCount(0)
		require.Equal(t, "0 users", userCount0.Localize("en"), "Zero should use plural form in English")
	})

	t.Run("LocalizableInterface", func(t *testing.T) {
		// Test that generated types implement the Localizable interface properly
		entity := EntityTexts.Product
		require.Equal(t, "product", entity.ID(), "Entity should have correct ID")

		// Test localization
		resultJa := entity.Localize("ja")
		require.Equal(t, "製品", resultJa, "Entity should localize correctly in Japanese")

		resultEn := entity.Localize("en")
		require.Equal(t, "Product", resultEn, "Entity should localize correctly in English")
	})

	t.Log("✅ Direct go-i18n runtime test passed successfully!")
}
