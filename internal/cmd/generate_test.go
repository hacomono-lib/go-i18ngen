package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hacomono-lib/go-i18ngen/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeConfig(t *testing.T) {
	t.Run("コマンド引数がconfig.yamlより優先される", func(t *testing.T) {
		// config.yamlの設定
		cfg := &config.Config{
			Locales:          []string{"ja"},
			Compound:         false,
			MessagesGlob:     "/config/messages/*.json",
			PlaceholdersGlob: "/config/placeholders/*.yaml",
			OutputDir:        "/config/output",
			OutputPackage:    "config_pkg",
		}

		// コマンド引数のフラグ
		flags := &Flags{
			Locales:          []string{"ja", "en"},
			Compound:         true,
			MessagesGlob:     "/cmd/messages/*.json",
			PlaceholdersGlob: "/cmd/placeholders/*.yaml",
			OutputDir:        "/cmd/output",
			OutputPackage:    "cmd_pkg",
		}

		merged := MergeConfig(cfg, flags)

		// コマンド引数の値が優先されることを確認
		assert.Equal(t, []string{"ja", "en"}, merged.Locales)
		assert.True(t, merged.Compound)
		assert.Equal(t, "/cmd/messages/*.json", merged.MessagesGlob)
		assert.Equal(t, "/cmd/placeholders/*.yaml", merged.PlaceholdersGlob)
		assert.Equal(t, "/cmd/output", merged.OutputDir)
		assert.Equal(t, "cmd_pkg", merged.OutputPackage)
	})

	t.Run("コマンド引数が空の場合はconfig.yamlの値を使用", func(t *testing.T) {
		// config.yamlの設定
		cfg := &config.Config{
			Locales:          []string{"ja"},
			Compound:         true,
			MessagesGlob:     "/config/messages/*.json",
			PlaceholdersGlob: "/config/placeholders/*.yaml",
			OutputDir:        "/config/output",
			OutputPackage:    "config_pkg",
		}

		// 空のコマンド引数フラグ
		flags := &Flags{}

		merged := MergeConfig(cfg, flags)

		// config.yamlの値がそのまま使用されることを確認
		assert.Equal(t, []string{"ja"}, merged.Locales)
		assert.True(t, merged.Compound)
		assert.Equal(t, "/config/messages/*.json", merged.MessagesGlob)
		assert.Equal(t, "/config/placeholders/*.yaml", merged.PlaceholdersGlob)
		assert.Equal(t, "/config/output", merged.OutputDir)
		assert.Equal(t, "config_pkg", merged.OutputPackage)
	})

	t.Run("部分的なコマンド引数の上書き", func(t *testing.T) {
		// config.yamlの設定
		cfg := &config.Config{
			Locales:          []string{"ja"},
			Compound:         false,
			MessagesGlob:     "/config/messages/*.json",
			PlaceholdersGlob: "/config/placeholders/*.yaml",
			OutputDir:        "/config/output",
			OutputPackage:    "config_pkg",
		}

		// 一部のコマンド引数のみ指定
		flags := &Flags{
			MessagesGlob: "/cmd/messages/*.json",
			OutputDir:    "/cmd/output",
		}

		merged := MergeConfig(cfg, flags)

		// 指定されたコマンド引数のみ上書きされ、その他はconfig.yamlの値を使用
		assert.Equal(t, []string{"ja"}, merged.Locales)                         // config.yamlの値
		assert.False(t, merged.Compound)                                        // config.yamlの値
		assert.Equal(t, "/cmd/messages/*.json", merged.MessagesGlob)            // コマンド引数で上書き
		assert.Equal(t, "/config/placeholders/*.yaml", merged.PlaceholdersGlob) // config.yamlの値
		assert.Equal(t, "/cmd/output", merged.OutputDir)                        // コマンド引数で上書き
		assert.Equal(t, "config_pkg", merged.OutputPackage)                     // config.yamlの値
	})
}

func TestPathResolutionBehavior(t *testing.T) {
	// 一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "i18ngen_path_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// サブディレクトリを作成
	configDir := filepath.Join(tempDir, "config")
	require.NoError(t, os.MkdirAll(configDir, 0755))

	t.Run("config.yamlファイル内のパスは設定ファイルからの相対パス", func(t *testing.T) {
		// config.yamlファイルを作成
		configPath := filepath.Join(configDir, "test_config.yaml")
		configContent := `messages: "messages/*.json"
placeholders: "placeholders/*.yaml"
output_dir: "output"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// config.yamlを読み込み
		cfg, err := config.LoadConfig(configPath)
		require.NoError(t, err)

		// パスが設定ファイルのディレクトリを基準に解決されることを確認
		expectedMessagesGlob := filepath.Join(configDir, "messages/*.json")
		expectedPlaceholdersGlob := filepath.Join(configDir, "placeholders/*.yaml")
		expectedOutputDir := filepath.Join(configDir, "output")

		assert.Equal(t, expectedMessagesGlob, cfg.MessagesGlob)
		assert.Equal(t, expectedPlaceholdersGlob, cfg.PlaceholdersGlob)
		assert.Equal(t, expectedOutputDir, cfg.OutputDir)
	})

	t.Run("コマンド引数のパスはそのまま使用される（実行ディレクトリからの相対パス）", func(t *testing.T) {
		// 空のconfig.yamlを作成
		configPath := filepath.Join(configDir, "empty_config.yaml")
		configContent := `locales: ["ja"]`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0644))

		// config.yamlを読み込み
		cfg, err := config.LoadConfig(configPath)
		require.NoError(t, err)

		// コマンド引数のフラグ（実行ディレクトリからの相対パス）
		flags := &Flags{
			MessagesGlob:     "cmd_messages/*.json",     // 実行ディレクトリから
			PlaceholdersGlob: "cmd_placeholders/*.yaml", // 実行ディレクトリから
			OutputDir:        "cmd_output",              // 実行ディレクトリから
		}

		merged := MergeConfig(cfg, flags)

		// コマンド引数のパスはそのまま使用される（パス解決されない）
		assert.Equal(t, "cmd_messages/*.json", merged.MessagesGlob)
		assert.Equal(t, "cmd_placeholders/*.yaml", merged.PlaceholdersGlob)
		assert.Equal(t, "cmd_output", merged.OutputDir)
	})
}
