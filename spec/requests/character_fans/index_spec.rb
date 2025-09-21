# typed: false
# frozen_string_literal: true

RSpec.describe "GET /characters/:character_id/fans", type: :request do
  it "キャラクターのファン一覧が表示されること" do
    character = create(:character)
    user = create(:registered_user)
    create(:character_favorite, user:, character:)

    get "/characters/#{character.id}/fans"

    expect(response.status).to eq(200)
    expect(response.body).to include(character.name)
    expect(response.body).to include(user.profile.name)
  end

  it "ファンがいない場合でもアクセスできること" do
    character = create(:character)

    get "/characters/#{character.id}/fans"

    expect(response.status).to eq(200)
    expect(response.body).to include(character.name)
  end

  it "複数のファンがいる場合、全員が表示されること" do
    character = create(:character)
    user1 = create(:registered_user)
    user2 = create(:registered_user)
    create(:character_favorite, user: user1, character:)
    create(:character_favorite, user: user2, character:)

    get "/characters/#{character.id}/fans"

    expect(response.status).to eq(200)
    expect(response.body).to include(character.name)
    expect(response.body).to include(user1.profile.name)
    expect(response.body).to include(user2.profile.name)
  end

  it "存在しないキャラクターIDの場合、404エラーが返されること" do
    expect {
      get "/characters/9999999/fans"
    }.to raise_error(ActiveRecord::RecordNotFound)
  end
end
