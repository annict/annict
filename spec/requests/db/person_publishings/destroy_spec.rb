# typed: false
# frozen_string_literal: true

RSpec.describe "DELETE /db/people/:id/publishing", type: :request do
  it "ログインしていないとき、ログインページにリダイレクトすること" do
    person = create(:person, :published)

    delete "/db/people/#{person.id}/publishing"
    person.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("ログインしてください")
    expect(person.published?).to eq(true)
  end

  it "エディター権限がないユーザーがログインしているとき、アクセスできないこと" do
    user = create(:registered_user)
    person = create(:person, :published)

    login_as(user, scope: :user)

    delete "/db/people/#{person.id}/publishing"
    person.reload

    expect(response.status).to eq(302)
    expect(flash[:alert]).to eq("アクセスできません")
    expect(person.published?).to eq(true)
  end

  it "エディター権限があるユーザーがログインしているとき、人物を非公開にできること" do
    user = create(:registered_user, :with_editor_role)
    person = create(:person, :published)

    login_as(user, scope: :user)

    expect(person.published?).to eq(true)

    delete "/db/people/#{person.id}/publishing"
    person.reload

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("非公開にしました")
    expect(person.published?).to eq(false)
  end

  it "存在しない人物IDを指定したとき、404エラーになること" do
    user = create(:registered_user, :with_editor_role)

    login_as(user, scope: :user)

    delete "/db/people/non-existent-id/publishing"

    expect(response).to have_http_status(:not_found)
  end

  it "すでに非公開の人物を指定したとき、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    person = create(:person, :unpublished)

    login_as(user, scope: :user)

    delete "/db/people/#{person.id}/publishing"

    expect(response).to have_http_status(:not_found)
  end

  it "削除済みの人物を指定したとき、404エラーになること" do
    user = create(:registered_user, :with_editor_role)
    person = create(:person, :published)
    person.destroy!

    login_as(user, scope: :user)

    delete "/db/people/#{person.id}/publishing"

    expect(response).to have_http_status(:not_found)
  end
end
