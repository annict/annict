# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/series/:id", type: :request do
  it "ログインしていないとき、アクセスできずログインページにリダイレクトすること" do
    series = create(:series, :not_deleted)
    expect(Series.count).to eq(1)

    delete "/db/series/#{series.id}"
    series.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(Series.count).to eq(1)
  end

  it "編集者でないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    series = create(:series, :not_deleted)
    login_as(user, scope: :user)

    expect(Series.count).to eq(1)

    delete "/db/series/#{series.id}"
    series.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Series.count).to eq(1)
  end

  it "編集者がログインしているとき、アクセスできないこと" do
    user = create(:registered_user, :with_editor_role)
    series = create(:series, :not_deleted)
    login_as(user, scope: :user)

    expect(Series.count).to eq(1)

    delete "/db/series/#{series.id}"
    series.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(Series.count).to eq(1)
  end

  it "管理者がログインしているとき、シリーズをソフトデリートできること" do
    user = create(:registered_user, :with_admin_role)
    series = create(:series, :not_deleted)
    login_as(user, scope: :user)

    expect(Series.count).to eq(1)

    delete "/db/series/#{series.id}"

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("削除しました")
    expect(Series.count).to eq(0)
  end
end
