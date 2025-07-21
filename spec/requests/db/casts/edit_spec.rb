# typed: false
# frozen_string_literal: true

RSpec.describe "GET /db/casts/:id/edit", type: :request do
  it "ログインしていない場合、アクセスできず認証エラーメッセージが表示されること" do
    cast = create(:cast)

    get "/db/casts/#{cast.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
  end

  it "編集者権限を持たないユーザーの場合、アクセスできずエラーメッセージが表示されること" do
    user = create(:registered_user)
    cast = create(:cast)

    login_as(user, scope: :user)
    get "/db/casts/#{cast.id}/edit"

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
  end

  it "編集者権限を持つユーザーの場合、キャスト編集フォームが表示されること" do
    user = create(:registered_user, :with_editor_role)
    cast = create(:cast)

    login_as(user, scope: :user)
    get "/db/casts/#{cast.id}/edit"

    expect(response.status).to eq(200)
    expect(response.body).to include(cast.character.name)
  end
end
