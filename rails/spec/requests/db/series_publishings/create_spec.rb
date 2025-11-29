# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/series/:id/publishing", type: :request do
  it "ログインしていないとき、アクセスできないこと" do
    series = create(:series, :unpublished)

    post "/db/series/#{series.id}/publishing"
    series.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(series.published?).to eq(false)
  end

  it "エディター権限がないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    series = create(:series, :unpublished)
    login_as(user, scope: :user)

    post "/db/series/#{series.id}/publishing"
    series.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(series.published?).to eq(false)
  end

  it "エディター権限があるユーザーがログインしているとき、シリーズを公開できること" do
    user = create(:registered_user, :with_editor_role)
    series = create(:series, :unpublished)
    login_as(user, scope: :user)

    expect(series.published?).to eq(false)

    post "/db/series/#{series.id}/publishing"
    series.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("公開しました")
    expect(series.published?).to eq(true)
  end

  it "存在しないシリーズIDを指定したとき、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    expect {
      post "/db/series/non-existent-id/publishing"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "既に公開済みのシリーズを指定したとき、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    series = create(:series, :published)
    login_as(user, scope: :user)

    expect {
      post "/db/series/#{series.id}/publishing"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "削除済みのシリーズを指定したとき、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    series = create(:series, :unpublished, deleted_at: Time.current)
    login_as(user, scope: :user)

    expect {
      post "/db/series/#{series.id}/publishing"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
