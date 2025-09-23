# typed: false
# frozen_string_literal: true

RSpec.describe "GET /userland/projects/:project_id/edit", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    project = FactoryBot.create(:userland_project)

    get "/userland/projects/#{project.id}/edit"

    expect(response).to redirect_to(new_user_session_path)
  end

  it "ログインしているがプロジェクトメンバーでないとき、403エラーが発生すること" do
    user = FactoryBot.create(:user, :with_profile)
    project = FactoryBot.create(:userland_project)

    login_as(user, scope: :user)

    expect {
      get "/userland/projects/#{project.id}/edit"
    }.to raise_error(Pundit::NotAuthorizedError)
  end

  it "ログインしていてプロジェクトメンバーのとき、ステータスコード200を返すこと" do
    user = FactoryBot.create(:user, :with_profile)
    project = FactoryBot.create(:userland_project)
    FactoryBot.create(:userland_project_member, user:, userland_project: project)

    login_as(user, scope: :user)

    get "/userland/projects/#{project.id}/edit"

    expect(response.status).to eq(200)
  end

  it "ログインしていてプロジェクトメンバーのとき、編集フォームが表示されること" do
    user = FactoryBot.create(:user, :with_profile)
    category = FactoryBot.create(:userland_category)
    project = FactoryBot.create(:userland_project, userland_category: category)
    FactoryBot.create(:userland_project_member, user:, userland_project: project)

    login_as(user, scope: :user)

    get "/userland/projects/#{project.id}/edit"

    expect(response.body).to include("userland_project[name]")
    expect(response.body).to include("userland_project[summary]")
    expect(response.body).to include("userland_project[description]")
    expect(response.body).to include("userland_project[url]")
    expect(response.body).to include("userland_project[userland_category_id]")
  end

  it "存在しないプロジェクトIDの場合、404エラーが返されること" do
    user = FactoryBot.create(:user, :with_profile)

    login_as(user, scope: :user)

    get "/userland/projects/999999999/edit"

    expect(response).to have_http_status(:not_found)
  end
end
