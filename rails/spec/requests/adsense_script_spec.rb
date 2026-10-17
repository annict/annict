# typed: false
# frozen_string_literal: true

# 広告のスクリプトは全レイアウト共通の `application/_head.html.erb` で読み込む。
# サポーターは広告を出さないことを特典として申し込んでいるため、ログイン中のサポーターには読み込ませない。
# 広告は本番環境でだけ出す (`display_ads?`) ため、本番環境として振る舞わせて確かめる。
RSpec.describe "AdSenseのスクリプトの読み込み", type: :request do
  let(:adsense_script_url) { "pagead2.googlesyndication.com" }

  before do
    allow(Rails.env).to receive(:production?).and_return(true)
  end

  it "ログインしていないとき、AdSenseのスクリプトを読み込むこと" do
    work = create(:work)

    get "/works/#{work.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(adsense_script_url)
  end

  it "サポーターでないユーザーがログインしているとき、AdSenseのスクリプトを読み込むこと" do
    user = create(:registered_user)
    work = create(:work)
    login_as(user, scope: :user)

    get "/works/#{work.id}"

    expect(response.status).to eq(200)
    expect(response.body).to include(adsense_script_url)
  end

  it "サポーターがログインしているとき、AdSenseのスクリプトを読み込まないこと" do
    user = create(:registered_user, :with_stripe_supporter)
    work = create(:work)
    login_as(user, scope: :user)

    get "/works/#{work.id}"

    expect(response.status).to eq(200)
    expect(response.body).not_to include(adsense_script_url)
  end
end
