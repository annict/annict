# typed: false
# frozen_string_literal: true

ENV["BUNDLE_GEMFILE"] ||= File.expand_path("../Gemfile", __dir__)

require "bundler/setup" # Set up gems listed in the Gemfile.

# concurrent-ruby gemを1.3.5に上げると発生するNameErrorを防ぐため、明示的に logger を読み込む
# Rails 7.1にアップデートするとこの対応が不要になる
# 詳細: https://stackoverflow.com/questions/79360526/uninitialized-constant-activesupportloggerthreadsafelevellogger-nameerror
require "logger"
