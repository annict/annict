# typed: false
# frozen_string_literal: true

RSpec.describe "GET /settings/account", type: :request do
  it "ログイン済みユーザーはページが表示されること" do
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/settings/account"

    expect(response.status).to eq(200)
    expect(response.body).to include("基本情報")

    deletion_link = Nokogiri::HTML(response.body).css("a").find do |link|
      link.text.strip == I18n.t("messages.sayonara.delete_account")
    end
    expect(deletion_link).to be_present
    expect(deletion_link["href"]).to eq("/settings/account/deletion/new")
    expect(deletion_link["data-method"]).to be_nil
    expect(deletion_link["data-confirm"]).to be_nil
  end

  it "未ログインユーザーはログインページにリダイレクトされること" do
    get "/settings/account"

    expect(response).to redirect_to(new_user_session_path)
  end
end
