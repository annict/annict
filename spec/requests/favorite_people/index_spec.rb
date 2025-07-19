# typed: false
# frozen_string_literal: true

RSpec.describe "GET /@:username/favorite_people", type: :request do
  it "お気に入りがいないとき、アクセスできること" do
    user = create(:registered_user)
    get "/@#{user.username}/favorite_people"

    expect(response.status).to eq(200)
    expect(response.body).to include("人物はいません")
  end

  it "お気に入り人物がいるとき、正常にアクセスできること" do
    user = create(:registered_user)
    person = create(:person)
    create(:person_favorite, user: user, person: person)

    get "/@#{user.username}/favorite_people"

    expect(response.status).to eq(200)
  end
end
