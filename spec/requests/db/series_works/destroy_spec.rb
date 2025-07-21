# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/series_works/:id", type: :request do
  it "ログインしていないとき、アクセスできないこと" do
    series_work = FactoryBot.create(:series_work, :not_deleted)

    expect(SeriesWork.count).to eq(1)

    delete "/db/series_works/#{series_work.id}"
    series_work.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")

    expect(SeriesWork.count).to eq(1)
  end

  it "一般ユーザーでログインしているとき、アクセスできないこと" do
    user = FactoryBot.create(:registered_user)
    series_work = FactoryBot.create(:series_work, :not_deleted)

    login_as(user, scope: :user)

    expect(SeriesWork.count).to eq(1)

    delete "/db/series_works/#{series_work.id}"
    series_work.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")

    expect(SeriesWork.count).to eq(1)
  end

  it "エディターでログインしているとき、アクセスできないこと" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    series_work = FactoryBot.create(:series_work, :not_deleted)

    login_as(user, scope: :user)

    expect(SeriesWork.count).to eq(1)

    delete "/db/series_works/#{series_work.id}"
    series_work.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")

    expect(SeriesWork.count).to eq(1)
  end

  it "管理者でログインしているとき、シリーズと作品の関連をソフトデリートできること" do
    user = FactoryBot.create(:registered_user, :with_admin_role)
    series_work = FactoryBot.create(:series_work, :not_deleted)

    login_as(user, scope: :user)

    expect(SeriesWork.count).to eq(1)

    delete "/db/series_works/#{series_work.id}"

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("削除しました")

    expect(SeriesWork.count).to eq(0)
  end
end
