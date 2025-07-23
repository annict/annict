# typed: false
# frozen_string_literal: true

RSpec.describe "GET /userland/projects/new", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    get "/userland/projects/new"

    expect(response).to redirect_to(new_user_session_path)
  end

  it "ログインしているとき、ステータスコード200を返すこと" do
    user = FactoryBot.create(:user, :with_profile)

    login_as(user, scope: :user)

    get "/userland/projects/new"

    expect(response.status).to eq(200)
  end

  it "ログインしているとき、新規プロジェクトフォームが表示されること" do
    user = FactoryBot.create(:user, :with_profile)

    login_as(user, scope: :user)

    get "/userland/projects/new"

    expect(response.body).to include("userland_project[name]")
    expect(response.body).to include("userland_project[summary]")
    expect(response.body).to include("userland_project[description]")
    expect(response.body).to include("userland_project[url]")
    expect(response.body).to include("userland_project[userland_category_id]")
  end
end
