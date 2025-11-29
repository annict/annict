# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/casts/:id", type: :request do
  it "ログインしていないとき、アクセスできずログインページにリダイレクトされること" do
    cast = create(:cast, :not_deleted)
    expect(Cast.count).to eq(1)

    delete "/db/casts/#{cast.id}"
    cast.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(Cast.count).to eq(1)
  end

  it "一般ユーザーでログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    cast = create(:cast, :not_deleted)
    login_as(user, scope: :user)
    expect(Cast.count).to eq(1)

    delete "/db/casts/#{cast.id}"
    cast.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Cast.count).to eq(1)
  end

  it "エディターユーザーでログインしているとき、アクセスできないこと" do
    user = create(:registered_user, :with_editor_role)
    cast = create(:cast, :not_deleted)
    login_as(user, scope: :user)
    expect(Cast.count).to eq(1)

    delete "/db/casts/#{cast.id}"
    cast.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Cast.count).to eq(1)
  end

  it "管理者ユーザーでログインしているとき、キャストを論理削除できること" do
    user = create(:registered_user, :with_admin_role)
    cast = create(:cast, :not_deleted)
    login_as(user, scope: :user)
    expect(Cast.count).to eq(1)

    delete "/db/casts/#{cast.id}"

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("削除しました")
    expect(Cast.count).to eq(0)
  end
end
