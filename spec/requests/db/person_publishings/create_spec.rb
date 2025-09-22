# typed: false
# frozen_string_literal: true

RSpec.describe "POST /db/people/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    person = FactoryBot.create(:person, :unpublished)

    post "/db/people/#{person.id}/publishing"
    person.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(person.published?).to eq(false)
  end

  it "エディター権限を持たないユーザーがログインしているとき、アクセスできないこと" do
    user = FactoryBot.create(:registered_user)
    person = FactoryBot.create(:person, :unpublished)

    login_as(user, scope: :user)

    post "/db/people/#{person.id}/publishing"
    person.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(person.published?).to eq(false)
  end

  it "エディター権限を持つユーザーがログインしているとき、人物を公開できること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    person = FactoryBot.create(:person, :unpublished)

    login_as(user, scope: :user)

    expect(person.published?).to eq(false)

    post "/db/people/#{person.id}/publishing"
    person.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("公開しました")
    expect(person.published?).to eq(true)
  end

  it "エディター権限を持つユーザーがログインしているとき、既に公開されている人物は404エラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    person = FactoryBot.create(:person, :published)

    login_as(user, scope: :user)

    expect(person.published?).to eq(true)

    post "/db/people/#{person.id}/publishing"

    expect(response.status).to eq(404)
  end

  it "エディター権限を持つユーザーがログインしているとき、削除された人物は404エラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)
    person = FactoryBot.create(:person, :unpublished, :deleted)

    login_as(user, scope: :user)

    expect(person.deleted?).to eq(true)

    post "/db/people/#{person.id}/publishing"

    expect(response.status).to eq(404)
  end

  it "エディター権限を持つユーザーがログインしているとき、存在しない人物IDは404エラーになること" do
    user = FactoryBot.create(:registered_user, :with_editor_role)

    login_as(user, scope: :user)

    post "/db/people/99999999/publishing"

    expect(response.status).to eq(404)
  end
end
