# typed: false
# frozen_string_literal: true

RSpec.describe "GET /@:username/favorite_characters", type: :request do
  it "お気に入りキャラクターがいないとき、アクセスできること" do
    user = create(:registered_user)
    get "/@#{user.username}/favorite_characters"

    expect(response.status).to eq(200)
    expect(response.body).to include("キャラクターはいません")
  end

  it "お気に入りキャラクターがいるとき、キャラクター一覧が表示されること" do
    user = create(:registered_user)
    character1 = create(:character)
    character2 = create(:character)
    create(:character_favorite, user:, character: character1)
    create(:character_favorite, user:, character: character2)

    get "/@#{user.username}/favorite_characters"

    expect(response.status).to eq(200)
    expect(response.body).to include(character1.name)
    expect(response.body).to include(character2.name)
  end

  it "存在しないユーザー名でアクセスしたとき、404エラーが返されること" do
    get "/@nonexistentuser/favorite_characters"

    expect(response.status).to eq(404)
  end

  it "削除されたユーザーでアクセスしたとき、404エラーが返されること" do
    user = create(:registered_user)
    user.destroy_in_batches

    expect {
      get "/@#{user.username}/favorite_characters"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
