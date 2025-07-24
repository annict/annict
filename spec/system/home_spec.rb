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

    # Turboが動作していることを確認
    expect(page).to have_css("[data-turbo-visit-control]", visible: :hidden)
  end

  it "ログインしたユーザーがトップページにアクセスできること" do
    user = FactoryBot.create(:user)
    sign_in_as(user)

    visit root_path

    expect(page).to have_content(user.username)
    expect(page).to have_http_status(:ok)
  end
end
