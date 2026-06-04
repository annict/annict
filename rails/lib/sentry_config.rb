# typed: false
# frozen_string_literal: true

# SentryConfig holds small helpers that translate environment-derived values
# into Sentry SDK configuration. Extracted from
# `config/initializers/sentry.rb` so that the parsing rules can be covered by
# unit tests without re-evaluating the initializer.
#
# [Ja] 環境変数由来の値を Sentry SDK 設定に変換するヘルパー群。
# `config/initializers/sentry.rb` から切り出すことで、initializer 全体を
# 再評価せずにパース規則をユニットテストできるようにする。
module SentryConfig
  DEFAULT_TRACES_SAMPLE_RATE = 0.5

  # Resolve the traces sample rate from a raw value (typically
  # `ENV["ANNICT_SENTRY_TRACES_SAMPLE_RATE"]`). Falls back to `default` when
  # the value is missing, malformed, or outside the [0.0, 1.0] range so a typo
  # in the deploy environment does not silently disable performance tracing
  # (same policy as the Go version's parseSentryTracesSampleRate).
  #
  # [Ja] サンプリングレートの生値 (通常は `ENV["ANNICT_SENTRY_TRACES_SAMPLE_RATE"]`)
  # を解決する。値が空・不正・範囲外の場合は `default` にフォールバックし、
  # デプロイ環境のタイポでパフォーマンストレースが静かに無効化されないようにする
  # (Go 版の parseSentryTracesSampleRate とポリシーを統一)。
  def self.resolve_traces_sample_rate(raw, default: DEFAULT_TRACES_SAMPLE_RATE)
    parsed =
      case raw
      when Numeric
        raw.to_f
      when String
        Float(raw, exception: false) if raw.present?
      end

    if parsed&.between?(0.0, 1.0)
      parsed
    else
      default
    end
  end
end
