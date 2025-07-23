# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/series/:id/publishing", type: :request do
  it "ログインしていないとき、アクセスできないこと" do
    series = create(:series, :published)

    delete "/db/series/#{series.id}/publishing"
    series.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(series.published?).to eq(true)
  end

  it "編集者ロールを持たないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    series = create(:series, :published)
    login_as(user, scope: :user)

    delete "/db/series/#{series.id}/publishing"
    series.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(series.published?).to eq(true)
  end

  it "編集者ロールを持つユーザーがログインしているとき、シリーズを非公開にできること" do
    user = create(:registered_user, :with_editor_role)
    series = create(:series, :published)
    login_as(user, scope: :user)

    expect(series.published?).to eq(true)

    delete "/db/series/#{series.id}/publishing"
    series.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("非公開にしました")
    expect(series.published?).to eq(false)
  end

  it "編集者ロールを持つユーザーがログインしているとき、存在しないシリーズIDの場合404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    expect do
      delete "/db/series/999999/publishing"
    end.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "編集者ロールを持つユーザーがログインしているとき、既に非公開のシリーズの場合404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    series = create(:series, :unpublished)
    login_as(user, scope: :user)

    expect do
      delete "/db/series/#{series.id}/publishing"
    end.to raise_error(ActiveRecord::RecordNotFound)
  end
end
