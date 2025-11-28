# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/works/:work_id/trailers/new", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    work = create(:work)

    get "/db/works/#{work.id}/trailers/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "エディター権限を持たないユーザーがログインしているとき、アクセスできないこと" do
    work = create(:work)
    user = create(:registered_user)
    login_as(user, scope: :user)

    get "/db/works/#{work.id}/trailers/new"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "エディター権限を持つユーザーがログインしているとき、PV登録ページが表示されること" do
    work = create(:work)
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    get "/db/works/#{work.id}/trailers/new"

    expect(response.status).to eq(200)
    expect(response.body).to include("PV登録")
  end
end
