# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/people/:id/edit", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトされること" do
    person = create(:person)

    get "/db/people/#{person.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "エディター権限を持たないユーザーがログインしているとき、アクセスが拒否されること" do
    user = create(:registered_user)
    person = create(:person)
    login_as(user, scope: :user)

    get "/db/people/#{person.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "エディター権限を持つユーザーがログインしているとき、編集フォームが表示されること" do
    user = create(:registered_user, :with_editor_role)
    person = create(:person)
    login_as(user, scope: :user)

    get "/db/people/#{person.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include(person.name)
  end

  it "管理者権限を持つユーザーがログインしているとき、編集フォームが表示されること" do
    user = create(:registered_user, :with_admin_role)
    person = create(:person)
    login_as(user, scope: :user)

    get "/db/people/#{person.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include(person.name)
  end

  it "存在しない人物の編集ページにアクセスしたとき、404エラーが発生すること" do
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    expect {
      get "/db/people/999999/edit"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除済みの人物の編集ページにアクセスしたとき、404エラーが発生すること" do
    user = create(:registered_user, :with_editor_role)
    person = create(:person, deleted_at: Time.current)
    login_as(user, scope: :user)

    expect {
      get "/db/people/#{person.id}/edit"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
