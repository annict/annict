# typed: false
# frozen_string_literal: true

RSpec.describe "Home page", type: :system do
  it "トップページが表示されること" do
    visit root_path

    expect(page).to have_content("Annict")
    expect(page).to have_http_status(:ok)
  end

  it "JavaScriptが有効な環境でトップページが表示されること", js: true do
    visit root_path

    expect(page).to have_content("Annict")

    # JavaScriptが読み込まれていることを確認
    wait_for_javascript

    # Turboフレームワークに関連する要素があることを確認
    expect(page).to have_css("meta[name='turbo-cache-control']", visible: :hidden)
  end

  it "ログインしたユーザーがトップページにアクセスできること" do
    # confirmされたユーザーを作成
    user = FactoryBot.create(:registered_user)

    sign_in(user:)

    # ログイン成功のメッセージまたはリダイレクトを確認
    expect(page).to have_current_path(root_path)

    # ユーザー名またはアバターが表示されることを確認
    # ナビゲーションバーなどにユーザー名が表示される
    expect(page).to have_content(user.username)
      .or have_css("img[alt*='#{user.username}']")
      .or have_css("[data-user-id='#{user.id}']")
  end
end
