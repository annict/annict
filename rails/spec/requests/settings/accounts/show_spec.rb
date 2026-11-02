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

  it "設定メニューにGo版の二要素認証の設定画面へのリンクが表示されること" do
    user_record = create(:registered_user)
    login_as(user_record, scope: :user)

    get "/settings/account"

    links = Nokogiri::HTML(response.body).css(".c-nav__link").select do |link|
      link["href"] == "/settings/two_factor_auth"
    end
    expect(links.size).to eq(1)
    expect(links.first.text.strip).to eq(I18n.t("head.title.settings.two_factor_auths.show"))
  end

  it "未ログインユーザーはログインページにリダイレクトされること" do
    get "/settings/account"

    expect(response).to redirect_to(new_user_session_path)
  end
end
