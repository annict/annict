# typed: false

include Warden::Test::Helpers

Warden.test_mode!

RSpec.configure do |config|
  config.include Devise::Test::ControllerHelpers, type: :controller

  # Reset Warden's test_mode state (registered users via `login_as`) after each
  # example. Without this, users logged in via `login_as` leak into subsequent
  # tests under random ordering, which can cause `authenticate_user!` to pass
  # in tests that expected an unauthenticated request (manifesting as
  # `flash[:alert]` becoming "アクセスできません" instead of "ログインしてください").
  # See https://github.com/heartcombo/devise/wiki/How-To:-Test-with-Capybara
  #
  # [Ja] `login_as` で Warden test_mode に登録したユーザーを各テスト終了時に
  # リセットする。これがないと、ランダム順実行時に `login_as` のユーザーが
  # 後続テストへ漏れて、未ログイン前提のテストで `authenticate_user!` が
  # 通ってしまい、`flash[:alert]` が "ログインしてください" ではなく
  # "アクセスできません" になる flaky 失敗が発生する。
  config.after do
    Warden.test_reset!
  end
end
