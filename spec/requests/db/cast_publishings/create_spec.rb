# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/casts/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    cast = FactoryBot.create(:cast, :unpublished)

    post "/db/casts/#{cast.id}/publishing"
    cast.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(cast.published?).to eq(false)
  end

  it "編集者権限を持たないユーザーがログインしているとき、アクセスできないこと" do
    user = FactoryBot.create(:registered_user)
    cast = FactoryBot.create(:cast, :unpublished)
    login_as(user, scope: :user)

    post "/db/casts/#{cast.id}/publishing"
    cast.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(cast.published?).to eq(false)
  end

  it "編集者権限を持つユーザーがログインしているとき、キャストを公開できること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    cast = FactoryBot.create(:cast, :unpublished)
    login_as(user, scope: :user)

    expect(cast.published?).to eq(false)

    post "/db/casts/#{cast.id}/publishing"
    cast.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("公開しました")
    expect(cast.published?).to eq(true)
  end

  it "すでに公開済みのキャストの場合、404エラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    cast = FactoryBot.create(:cast, :published)
    login_as(user, scope: :user)

    expect(cast.published?).to eq(true)

    expect do
      post "/db/casts/#{cast.id}/publishing"
    end.to raise_error(ActiveRecord::RecordNotFound)
  end

  it "存在しないキャストIDが指定されたとき、404エラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    login_as(user, scope: :user)

    expect do
      post "/db/casts/non-existent-id/publishing"
    end.to raise_error(ActiveRecord::RecordNotFound)
  end
end
